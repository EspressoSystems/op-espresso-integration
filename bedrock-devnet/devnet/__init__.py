import argparse
import logging
import os
import subprocess
import json
import socket
import calendar
import datetime
import time
import shutil
import http.client
from multiprocessing import Process, Queue

import devnet.log_setup

pjoin = os.path.join

parser = argparse.ArgumentParser(description='Bedrock devnet launcher')
parser.add_argument('--monorepo-dir', help='Directory of the monorepo', default=os.getcwd())
parser.add_argument('--allocs', help='Only create the allocs and exit', type=bool, action=argparse.BooleanOptionalAction)
parser.add_argument('--test', help='Tests the deployment, must already be deployed', type=bool, action=argparse.BooleanOptionalAction)
parser.add_argument('--l2', help='Which L2 to run', type=str, default='op1')
parser.add_argument('--l2-provider-url', help='URL for the L2 RPC node', type=str, default='http://localhost:19545')
parser.add_argument('--deploy-l2', help='Deploy the L2 onto a running L1 and sequencer network', type=bool, action=argparse.BooleanOptionalAction)
parser.add_argument('--deploy-config', help='Deployment config, relative to packages/contracts-bedrock/deploy-config', default='devnetL1.json')
parser.add_argument('--deployment', help='Path to deployment output files, relative to packages/contracts-bedrock/deployments', default='devnetL1')
parser.add_argument('--devnet-dir', help='Output path for devnet config, relative to --monorepo-dir', default='.devnet')
parser.add_argument('--espresso', help='Run on Espresso Sequencer', type=bool, action=argparse.BooleanOptionalAction)
parser.add_argument('--skip-build', help='Skip building docker images', type=bool, action=argparse.BooleanOptionalAction)
parser.add_argument('--build', help='Only build docker images', type=bool, action=argparse.BooleanOptionalAction)

log = logging.getLogger()

class Bunch:
    def __init__(self, **kwds):
        self.__dict__.update(kwds)

class ChildProcess:
    def __init__(self, func, *args):
        self.errq = Queue()
        self.process = Process(target=self._func, args=(func, args))

    def _func(self, func, args):
        try:
            func(*args)
        except Exception as e:
            self.errq.put(str(e))

    def start(self):
        self.process.start()

    def join(self):
        self.process.join()

    def get_error(self):
        return self.errq.get() if not self.errq.empty() else None


def main():
    args = parser.parse_args()

    monorepo_dir = os.path.abspath(args.monorepo_dir)
    devnet_dir = pjoin(monorepo_dir, args.devnet_dir)
    contracts_bedrock_dir = pjoin(monorepo_dir, 'packages', 'contracts-bedrock')
    deployment_dir = pjoin(contracts_bedrock_dir, 'deployments', args.deployment)
    op_node_dir = pjoin(args.monorepo_dir, 'op-node')
    ops_bedrock_dir = pjoin(monorepo_dir, 'ops-bedrock')
    deploy_config_dir = pjoin(contracts_bedrock_dir, 'deploy-config'),
    devnet_config_path = pjoin(contracts_bedrock_dir, 'deploy-config', args.deploy_config)
    ops_chain_ops = pjoin(monorepo_dir, 'op-chain-ops')
    sdk_dir = pjoin(monorepo_dir, 'packages', 'sdk')

    paths = Bunch(
      mono_repo_dir=monorepo_dir,
      devnet_dir=devnet_dir,
      contracts_bedrock_dir=contracts_bedrock_dir,
      deployment_dir=deployment_dir,
      l1_deployments_path=pjoin(deployment_dir, '.deploy'),
      deploy_config_dir=deploy_config_dir,
      devnet_config_path=devnet_config_path,
      op_node_dir=op_node_dir,
      ops_bedrock_dir=ops_bedrock_dir,
      ops_chain_ops=ops_chain_ops,
      sdk_dir=sdk_dir,
      genesis_l1_path=pjoin(devnet_dir, 'genesis-l1.json'),
      genesis_l2_path=pjoin(devnet_dir, 'genesis-l2.json'),
      allocs_path=pjoin(devnet_dir, 'allocs-l1.json'),
      addresses_json_path=pjoin(devnet_dir, 'addresses.json'),
      sdk_addresses_json_path=pjoin(devnet_dir, 'sdk-addresses.json'),
      rollup_config_path=pjoin(devnet_dir, 'rollup.json')
    )

    if args.test:
      log.info('Testing deployed devnet')
      devnet_test(paths, args.l2_provider_url)
      return

    os.makedirs(devnet_dir, exist_ok=True)

    if args.allocs:
        devnet_l1_genesis(paths, args.deploy_config)
        return

    if args.skip_build:
        log.warn('Skipping building docker images')
    else:
        log.info('Building docker images')
        run_command(['docker', 'compose', 'build', '--progress', 'plain'], cwd=paths.ops_bedrock_dir, env={
            'PWD': paths.ops_bedrock_dir,
            'DEVNET_DIR': paths.devnet_dir
        })

    if args.build:
        log.info("Finished building")
        return

    log.info('Devnet starting')
    devnet_deploy(paths, args)


def deploy_contracts(paths, deploy_config: str):
    wait_up(8545)
    wait_for_rpc_server('127.0.0.1:8545')
    res = eth_accounts('127.0.0.1:8545')

    response = json.loads(res)
    account = response['result'][0]

    fqn = 'scripts/Deploy.s.sol:Deploy'
    run_command([
        'forge', 'script', fqn, '--sender', account,
        '--rpc-url', 'http://127.0.0.1:8545', '--broadcast',
        '--unlocked'
    ], env={
        'DEPLOYMENT_CONTEXT': deploy_config.removesuffix('.json')
    }, cwd=paths.contracts_bedrock_dir)

    shutil.copy(paths.l1_deployments_path, paths.addresses_json_path)

    log.info('Syncing contracts.')
    run_command([
        'forge', 'script', fqn, '--sig', 'sync()',
        '--rpc-url', 'http://127.0.0.1:8545'
    ], env={
        'DEPLOYMENT_CONTEXT': deploy_config.removesuffix('.json')
    }, cwd=paths.contracts_bedrock_dir)



def devnet_l1_genesis(paths, deploy_config: str):
    log.info('Generating L1 genesis state')
    geth = subprocess.Popen([
        'geth', '--dev', '--http', '--http.api', 'eth,debug',
        '--verbosity', '4', '--gcmode', 'archive', '--dev.gaslimit', '30000000'
    ])

    forge = ChildProcess(deploy_contracts, paths, deploy_config)
    forge.start()
    forge.join()
    err = forge.get_error()
    if err:
        raise Exception(f"Exception occurred in child process: {err}")

    res = debug_dumpBlock('127.0.0.1:8545')
    response = json.loads(res)
    allocs = response['result']

    write_json(paths.allocs_path, allocs)
    geth.terminate()


# Bring up the devnet where the contracts are deployed to L1
def devnet_deploy(paths, args):
    espresso = args.espresso
    l2 = args.l2
    l2_provider_url = args.l2_provider_url

    if os.path.exists(paths.genesis_l1_path) and os.path.isfile(paths.genesis_l1_path):
        log.info('L1 genesis already generated.')
    elif not args.deploy_l2:
        # Generate the L1 genesis, unless we are deploying an L2 onto an existing L1.
        log.info('Generating L1 genesis.')
        if os.path.exists(paths.allocs_path) == False:
            devnet_l1_genesis(paths, args.deploy_config)

        devnet_config_backup = pjoin(paths.devnet_dir, 'devnetL1.json.bak')
        shutil.copy(paths.devnet_config_path, devnet_config_backup)
        deploy_config = read_json(paths.devnet_config_path)
        deploy_config['l1GenesisBlockTimestamp'] = '{:#x}'.format(int(time.time()))
        write_json(paths.devnet_config_path, deploy_config)
        outfile_l1 = pjoin(paths.devnet_dir, 'genesis-l1.json')

        run_command([
            'go', 'run', 'cmd/main.go', 'genesis', 'l1',
            '--deploy-config', paths.devnet_config_path,
            '--l1-allocs', paths.allocs_path,
            '--l1-deployments', paths.addresses_json_path,
            '--outfile.l1', outfile_l1,
        ], cwd=paths.op_node_dir)

    if args.deploy_l2:
        # L1 and sequencer already exist, just deploy the L1 contracts for the new L2.
        deploy_contracts(paths, args.deploy_config)
    else:
        # Deploy L1 and sequencer network.
        log.info('Starting L1.')
        run_command(['docker', 'compose', 'up', '-d', 'l1'], cwd=paths.ops_bedrock_dir, env={
            'PWD': paths.ops_bedrock_dir,
            'DEVNET_DIR': paths.devnet_dir
        })
        wait_up(8545)
        wait_for_rpc_server('127.0.0.1:8545')

        if espresso:
            log.info('Starting Espresso sequencer.')
            espresso_services = [
                'orchestrator',
                'da-server',
                'consensus-server',
                'commitment-task',
                'sequencer0',
                'sequencer1',
            ]
            run_command(['docker-compose', 'up', '-d'] + espresso_services, cwd=paths.ops_bedrock_dir, env={
                'PWD': paths.ops_bedrock_dir,
                'DEVNET_DIR': paths.devnet_dir
            })

    # Re-build the L2 genesis unconditionally in Espresso mode, since we require the timestamps to be recent.
    if not espresso and os.path.exists(paths.genesis_l2_path) and os.path.isfile(paths.genesis_l2_path):
        log.info('L2 genesis and rollup configs already generated.')
    else:
        log.info('Generating L2 genesis and rollup configs.')
        run_command([
            'go', 'run', 'cmd/main.go', 'genesis', 'l2',
            '--l1-rpc', 'http://localhost:8545',
            '--deploy-config', paths.devnet_config_path,
            '--deployment-dir', paths.deployment_dir,
            '--outfile.l2', pjoin(paths.devnet_dir, 'genesis-l2.json'),
            '--outfile.rollup', pjoin(paths.devnet_dir, 'rollup.json')
        ], cwd=paths.op_node_dir)

    rollup_config = read_json(paths.rollup_config_path)
    addresses = read_json(paths.addresses_json_path)

    log.info('Bringing up L2.')
    run_command(['docker', 'compose', 'up', '-d', f'{l2}-l2', f'{l2}-geth-proxy'], cwd=paths.ops_bedrock_dir, env={
        'PWD': paths.ops_bedrock_dir,
        'DEVNET_DIR': paths.devnet_dir
    })

    l2_provider_port = int(l2_provider_url.split(':')[-1])
    l2_provider_http = l2_provider_url.removeprefix('http://')
    wait_up(l2_provider_port)
    wait_for_rpc_server(l2_provider_http)

    l2_output_oracle = addresses['L2OutputOracleProxy']
    log.info(f'Using L2OutputOracle {l2_output_oracle}')
    batch_inbox_address = rollup_config['batch_inbox_address']
    log.info(f'Using batch inbox {batch_inbox_address}')

    log.info('Bringing up everything else.')
    command = ['docker', 'compose', 'up', '-d']
    if args.deploy_l2:
        # If we are deploying onto an existing L1, don't restart the services that are already
        # running.
        command.append('--no-recreate')
    services = [f'{l2}-node', f'{l2}-proposer', f'{l2}-batcher']
    run_command(command + services, cwd=paths.ops_bedrock_dir, env={
        'PWD': paths.ops_bedrock_dir,
        'L2OO_ADDRESS': l2_output_oracle,
        'SEQUENCER_BATCH_INBOX_ADDRESS': batch_inbox_address,
        'DEVNET_DIR': paths.devnet_dir
    })

    # TODO the block explorer doesn't support running multiple instances simultaneously, so only
    # start it with the first L2, not ones deployed after the fact. We should support each L2 having
    # its own block explorer and then run this command unconditionally.
    if not args.deploy_l2:
        log.info('Starting block explorer')
        run_command(['docker-compose', 'up', '-d', f'{l2}-blockscout'], cwd=paths.ops_bedrock_dir)

    log.info('Devnet ready.')


def eth_accounts(url):
    log.info(f'Fetch eth_accounts {url}')
    conn = http.client.HTTPConnection(url)
    headers = {'Content-type': 'application/json'}
    body = '{"id":2, "jsonrpc":"2.0", "method": "eth_accounts", "params":[]}'
    conn.request('POST', '/', body, headers)
    response = conn.getresponse()
    data = response.read().decode()
    conn.close()
    return data


def debug_dumpBlock(url):
    log.info(f'Fetch debug_dumpBlock {url}')
    conn = http.client.HTTPConnection(url)
    headers = {'Content-type': 'application/json'}
    body = '{"id":3, "jsonrpc":"2.0", "method": "debug_dumpBlock", "params":["latest"]}'
    conn.request('POST', '/', body, headers)
    response = conn.getresponse()
    data = response.read().decode()
    conn.close()
    return data


def wait_for_rpc_server(url):
    log.info(f'Waiting for RPC server at {url}')

    conn = http.client.HTTPConnection(url)
    headers = {'Content-type': 'application/json'}
    body = '{"id":1, "jsonrpc":"2.0", "method": "eth_chainId", "params":[]}'

    while True:
        try:
            conn.request('POST', '/', body, headers)
            response = conn.getresponse()
            conn.close()
            if response.status < 300:
                log.info(f'RPC server at {url} ready')
                return
        except Exception as e:
            log.info(f'Error connecting to RPC: {e}')
            log.info(f'Waiting for RPC server at {url}')
            time.sleep(1)

def devnet_test(paths, l2_provider_url):
    # Check the L2 config
    run_command(
        ['go', 'run', 'cmd/check-l2/main.go', '--l2-rpc-url', l2_provider_url, '--l1-rpc-url', 'http://localhost:8545'],
        cwd=paths.ops_chain_ops,
    )

    run_command(
         ['npx', 'hardhat',  'deposit-erc20', '--network',  'devnetL1', '--l1-contracts-json-path', paths.addresses_json_path, '--l2-provider-url', l2_provider_url],
         cwd=paths.sdk_dir,
         timeout=8*60,
    )

    run_command(
         ['npx', 'hardhat',  'deposit-eth', '--network',  'devnetL1', '--l1-contracts-json-path', paths.addresses_json_path, '--l2-provider-url', l2_provider_url],
         cwd=paths.sdk_dir,
         timeout=8*60,
    )

def run_command(args, check=True, shell=False, cwd=None, env=None, timeout=None):
    env = env if env else {}
    return subprocess.run(
        args,
        check=check,
        shell=shell,
        env={
            **os.environ,
            **env
        },
        cwd=cwd,
        timeout=timeout
    )


def wait_up(port, retries=10, wait_secs=1):
    for i in range(0, retries):
        log.info(f'Trying 127.0.0.1:{port}')
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        try:
            s.connect(('127.0.0.1', int(port)))
            s.shutdown(2)
            log.info(f'Connected 127.0.0.1:{port}')
            return True
        except Exception:
            time.sleep(wait_secs)

    raise Exception(f'Timed out waiting for port {port}.')


def write_json(path, data):
    with open(path, 'w+') as f:
        json.dump(data, f, indent='  ')


def read_json(path):
    with open(path, 'r') as f:
        return json.load(f)

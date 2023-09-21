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
parser.add_argument('--faucet-url', help='URL for the L2 faucet', type=str, default='http://localhost:18111')
parser.add_argument('--deploy-l2', help='Deploy the L2 onto a running L1 and sequencer network', type=bool, action=argparse.BooleanOptionalAction)
parser.add_argument('--deploy-config', help='Deployment config, relative to packages/contracts-bedrock/deploy-config', default='devnetL1.json')
parser.add_argument('--deploy-config-template', help='Deployment config template, relative to packages/contracts-bedrock/deploy-config', default='devnetL1-template.json')
parser.add_argument('--deployment', help='Path to deployment output files, relative to packages/contracts-bedrock/deployments', default='devnetL1')
parser.add_argument('--devnet-dir', help='Output path for devnet config, relative to --monorepo-dir', default='.devnet')
parser.add_argument('--espresso', help='Run on Espresso Sequencer', type=bool, action=argparse.BooleanOptionalAction)
parser.add_argument("--compose-file", help="Compose file to use for demo images", type=str, default="docker-compose.yml")

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
    devnet_config_template_path = pjoin(contracts_bedrock_dir, 'deploy-config', args.deploy_config_template)
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
      devnet_config_template_path=devnet_config_template_path,
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
        devnet_test(paths, args.l2_provider_url, args.faucet_url)
        return

    os.makedirs(devnet_dir, exist_ok=True)

    if args.allocs:
        devnet_l1_genesis(paths, args.deploy_config)
        return

    log.info('Devnet starting')
    devnet_deploy(paths, args)

    log.info('Deploying ERC20 contract')
    deploy_erc20(paths, args.l2_provider_url)


def deploy_contracts(paths, deploy_config: str, deploy_l2: bool):
    wait_up(8545)
    wait_for_rpc_server('127.0.0.1:8545')
    res = eth_accounts('127.0.0.1:8545')

    response = json.loads(res)
    account = response['result'][0]
    log.info(f'Deploying with {account}')

    # The create2 account is shared by both L2s, so don't redeploy it unless necessary
    # We check to see if the create2 deployer exists by querying its balance
    res = run_command(
        ["cast", "balance", "0x3fAB184622Dc19b6109349B94811493BF2a45362"],
        capture_output=True,
    )
    deployer_balance = int(res.stdout.strip())
    if deployer_balance == 0:
        # send some ether to the create2 deployer account
        run_command([
            'cast', 'send', '--from', account,
            '--rpc-url', 'http://127.0.0.1:8545',
            '--unlocked', '--value', '1ether', '0x3fAB184622Dc19b6109349B94811493BF2a45362'
        ], env={}, cwd=paths.contracts_bedrock_dir)

        # deploy the create2 deployer
        run_command([
          'cast', 'publish', '--rpc-url', 'http://127.0.0.1:8545',
          '0xf8a58085174876e800830186a08080b853604580600e600039806000f350fe7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe03601600081602082378035828234f58015156039578182fd5b8082525050506014600cf31ba02222222222222222222222222222222222222222222222222222222222222222a02222222222222222222222222222222222222222222222222222222222222222'
        ], env={}, cwd=paths.contracts_bedrock_dir)

    deploy_env = {
        'DEPLOYMENT_CONTEXT': deploy_config.removesuffix('.json')
    }
    if deploy_l2:
        # If deploying an L2 onto an existing L1, use a different deployer salt so the contracts
        # will not collide with those of the existing L2.
        deploy_env['IMPL_SALT'] = os.urandom(32).hex()

    fqn = 'scripts/Deploy.s.sol:Deploy'
    run_command([
        'forge', 'script', fqn, '--sender', account,
        '--rpc-url', 'http://127.0.0.1:8545', '--broadcast',
        '--unlocked'
    ], env=deploy_env, cwd=paths.contracts_bedrock_dir)

    shutil.copy(paths.l1_deployments_path, paths.addresses_json_path)

    log.info('Syncing contracts.')
    run_command([
        'forge', 'script', fqn, '--sig', 'sync()',
        '--rpc-url', 'http://127.0.0.1:8545'
    ], env=deploy_env, cwd=paths.contracts_bedrock_dir)

def init_devnet_l1_deploy_config(paths, update_timestamp=False):
    deploy_config = read_json(paths.devnet_config_template_path)
    if update_timestamp:
        deploy_config['l1GenesisBlockTimestamp'] = '{:#x}'.format(int(time.time()))
    write_json(paths.devnet_config_path, deploy_config)

def devnet_l1_genesis(paths, deploy_config: str):
    log.info('Generating L1 genesis state')

    # Abort if there is an existing geth process listening on localhost:8545. It
    # may cause the op-node to fail to start due to a bad genesis block.
    geth_up = False
    try:
        geth_up = wait_up(8545, retries=1, wait_secs=0)
    except:
        pass
    if geth_up:
        raise Exception('Existing process is listening on localhost:8545, please kill it and try again. (e.g. `pkill geth`)')

    init_devnet_l1_deploy_config(paths)

    geth = subprocess.Popen([
        'geth', '--dev', '--http', '--http.api', 'eth,debug',
        '--verbosity', '4', '--gcmode', 'archive', '--dev.gaslimit', '30000000',
        '--rpc.allow-unprotected-txs'
    ])

    forge = ChildProcess(deploy_contracts, paths, deploy_config, False)
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
    compose_file = args.compose_file

    if os.path.exists(paths.genesis_l1_path) and os.path.isfile(paths.genesis_l1_path):
        log.info('L1 genesis already generated.')
    elif not args.deploy_l2:
        # Generate the L1 genesis, unless we are deploying an L2 onto an existing L1.
        log.info('Generating L1 genesis.')
        if os.path.exists(paths.allocs_path) == False:
            devnet_l1_genesis(paths, args.deploy_config)

        # It's odd that we want to regenerate the devnetL1.json file with
        # an updated timestamp different than the one used in the devnet_l1_genesis
        # function.  But, without it, CI flakes on this test rather consistently.
        # If someone reads this comment and understands why this is being done, please
        # update this comment to explain.
        init_devnet_l1_deploy_config(paths, update_timestamp=True)
        outfile_l1 = pjoin(paths.devnet_dir, 'genesis-l1.json')
        run_command([
            'go', 'run', 'cmd/main.go', 'genesis', 'l1',
            '--deploy-config', paths.devnet_config_path,
            '--l1-allocs', paths.allocs_path,
            '--l1-deployments', paths.addresses_json_path,
            '--outfile.l1', outfile_l1,
        ], cwd=paths.op_node_dir)

    if args.deploy_l2:
        # L1 and sequencer already exist, just create the deploy config and deploy the L1 contracts
        # for the new L2.
        init_devnet_l1_deploy_config(paths, update_timestamp=True)
        deploy_contracts(paths, args.deploy_config, args.deploy_l2)
    else:
        # Deploy L1 and sequencer network.
        log.info('Starting L1.')
        run_command(['docker', 'compose', '-f', compose_file, 'up', '-d', 'l1'], cwd=paths.ops_bedrock_dir, env={
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
            run_command(['docker-compose', '-f', compose_file, 'up', '-d'] + espresso_services, cwd=paths.ops_bedrock_dir, env={
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
    run_command(['docker', 'compose', '-f', compose_file, 'up', '-d', f'{l2}-l2', f'{l2}-geth-proxy'], cwd=paths.ops_bedrock_dir, env={
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
    command = ['docker', 'compose', '-f', compose_file, 'up', '-d']
    if args.deploy_l2:
        # If we are deploying onto an existing L1, don't restart the services that are already
        # running.
        command.append('--no-recreate')
    services = [f'{l2}-node', f'{l2}-proposer', f'{l2}-batcher', f'{l2}-faucet']
    run_command(command + services, cwd=paths.ops_bedrock_dir, env={
        'PWD': paths.ops_bedrock_dir,
        'L2OO_ADDRESS': l2_output_oracle,
        'SEQUENCER_BATCH_INBOX_ADDRESS': batch_inbox_address,
        'DEVNET_DIR': paths.devnet_dir
    })

    log.info('Starting block explorer')
    run_command(['docker-compose', '-f', compose_file, 'up', '-d', f'{l2}-blockscout'], cwd=paths.ops_bedrock_dir)

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

def deploy_erc20(paths, l2_provider_url):
    run_command(
         ['npx', 'hardhat',  'deploy-erc20', '--network',  'devnetL1', '--l2-provider-url', l2_provider_url],
         cwd=paths.sdk_dir,
         timeout=60,
    )

def devnet_test(paths, l2_provider_url, faucet_url):
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

    run_command(
         ['npx', 'hardhat',  'faucet-request', '--network',  'devnetL1', '--l2-provider-url', l2_provider_url, '--faucet-url', faucet_url],
         cwd=paths.sdk_dir,
         timeout=8*60,
    )

def run_command(args, check=True, shell=False, cwd=None, env=None, timeout=None, capture_output=False):
    env = env if env else {}
    return subprocess.run(
        args,
        check=check,
        shell=shell,
        capture_output=capture_output,
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

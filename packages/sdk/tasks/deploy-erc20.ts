import { task, types } from 'hardhat/config'
import { HardhatRuntimeEnvironment } from 'hardhat/types'
import '@nomiclabs/hardhat-ethers'
import 'hardhat-deploy'
import { Contract, Wallet, providers, utils } from 'ethers'
import Artifact__WETH9 from '@eth-optimism/contracts-bedrock/forge-artifacts/WETH9.sol/WETH9.json'

import {
  CrossChainMessenger,
  MessageStatus,
  CONTRACT_ADDRESSES,
  OEContractsLike,
  DEFAULT_L2_CONTRACT_ADDRESSES,
} from '../src'

const deployWETH9 = async (
  hre: HardhatRuntimeEnvironment,
  wrap: boolean,
  signer: Wallet,
): Promise<Contract> => {
  const Factory__WETH9 = new hre.ethers.ContractFactory(
    Artifact__WETH9.abi,
    Artifact__WETH9.bytecode.object,
    signer
  )

  console.log('Sending deployment transaction')
  const WETH9 = await Factory__WETH9.deploy()
  const receipt = await WETH9.deployTransaction.wait()
  console.log(`WETH9 deployed in tx: ${receipt.transactionHash}`)

  if (wrap) {
    const deposit = await signer.sendTransaction({
      value: utils.parseEther('12345'),
      to: WETH9.address,
    })
    await deposit.wait()
  }

  return WETH9
}

task('deploy-erc20', 'Deploy WETH9 onto L2.')
  .addParam(
    'l2ProviderUrl',
    'L2 provider URL.',
    'http://localhost:9545',
    types.string
  )
  .setAction(async (args, hre) => {
    // Use signer 11, hopefully not used by anything else.
    const wallet = Wallet.fromMnemonic(
      "test test test test test test test test test test test junk",
      "m/44'/60'/0'/0/10"
    );
    console.log("Using deployer wallet address: ", wallet.address)

    const l2Provider = new providers.StaticJsonRpcProvider(args.l2ProviderUrl)
    const l2Signer = wallet.connect(l2Provider);

    console.log('Deploying WETH9 to L2')
    const WETH_L2 = await deployWETH9(hre, true, l2Signer)
    console.log(`Deployed to L2 ${WETH_L2.address}`)
  })

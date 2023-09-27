import { task, types } from 'hardhat/config'
import { providers, utils } from 'ethers'
import fetch from 'node-fetch'

const GRANT_AMOUNT = utils.parseEther('0.1')

task('faucet-request', 'Deposits WETH9 onto L2.')
  .addParam(
    'l2ProviderUrl',
    'L2 provider URL.',
    'http://localhost:9545',
    types.string
  )
  .addParam(
    'faucetUrl',
    'L2 faucet URL',
    'http://localhost:18112',
    types.string
  )
  .setAction(async (args, hre) => {
    const signers = await hre.ethers.getSigners()
    if (signers.length === 0) {
      throw new Error('No configured signers')
    }
    // Use the last configured signer so we don't conflict with some other task
    const provider = new providers.StaticJsonRpcProvider(args.l2ProviderUrl)
    const signer = new hre.ethers.Wallet(
      hre.network.config.accounts[0],
      provider
    )
    const address = await signer.getAddress()
    console.log(`Using signer ${address}`)

    // Check the signer's balance before we do anything
    const initialBalance = await signer.getBalance()
    console.log(`Signer initial balance is ${initialBalance}`)

    // Request from the faucet
    const res = await fetch(`${args.faucetUrl}/faucet/request/${address}`, {
      method: 'POST',
    })
    if (res.status !== 200) {
      throw new Error(`Request failed with status ${res.status}`)
    }

    // Wait for our balance to increase.
    for (;;) {
      const balance = await signer.getBalance()
      console.log(`Signer balance is ${balance}`)
      if (balance >= initialBalance + GRANT_AMOUNT) {
        break
      }
      await new Promise((resolve) => setTimeout(resolve, 1000))
    }
  })

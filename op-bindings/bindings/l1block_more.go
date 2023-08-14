// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"encoding/json"

	"github.com/ethereum-optimism/optimism/op-bindings/solc"
)

const L1BlockStorageLayoutJSON = "{\"storage\":[{\"astId\":1000,\"contract\":\"contracts/L2/L1Block.sol:L1Block\",\"label\":\"number\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_uint64\"},{\"astId\":1001,\"contract\":\"contracts/L2/L1Block.sol:L1Block\",\"label\":\"timestamp\",\"offset\":8,\"slot\":\"0\",\"type\":\"t_uint64\"},{\"astId\":1002,\"contract\":\"contracts/L2/L1Block.sol:L1Block\",\"label\":\"basefee\",\"offset\":0,\"slot\":\"1\",\"type\":\"t_uint256\"},{\"astId\":1003,\"contract\":\"contracts/L2/L1Block.sol:L1Block\",\"label\":\"hash\",\"offset\":0,\"slot\":\"2\",\"type\":\"t_bytes32\"},{\"astId\":1004,\"contract\":\"contracts/L2/L1Block.sol:L1Block\",\"label\":\"sequenceNumber\",\"offset\":0,\"slot\":\"3\",\"type\":\"t_uint64\"},{\"astId\":1005,\"contract\":\"contracts/L2/L1Block.sol:L1Block\",\"label\":\"batcherHash\",\"offset\":0,\"slot\":\"4\",\"type\":\"t_bytes32\"},{\"astId\":1006,\"contract\":\"contracts/L2/L1Block.sol:L1Block\",\"label\":\"l1FeeOverhead\",\"offset\":0,\"slot\":\"5\",\"type\":\"t_uint256\"},{\"astId\":1007,\"contract\":\"contracts/L2/L1Block.sol:L1Block\",\"label\":\"l1FeeScalar\",\"offset\":0,\"slot\":\"6\",\"type\":\"t_uint256\"}],\"types\":{\"t_bytes32\":{\"encoding\":\"inplace\",\"label\":\"bytes32\",\"numberOfBytes\":\"32\"},\"t_uint256\":{\"encoding\":\"inplace\",\"label\":\"uint256\",\"numberOfBytes\":\"32\"},\"t_uint64\":{\"encoding\":\"inplace\",\"label\":\"uint64\",\"numberOfBytes\":\"8\"}}}"

var L1BlockStorageLayout = new(solc.StorageLayout)

var L1BlockDeployedBin = "0x608060405234801561001057600080fd5b50600436106100c95760003560e01c80638381f58a11610081578063b80777ea1161005b578063b80777ea14610170578063e591b28214610190578063e81b2c6d146101d057600080fd5b80638381f58a1461014a5780638b239f731461015e5780639e8c49661461016757600080fd5b80635cf24969116100b25780635cf24969146100ff57806364ca23ef146101085780637f122dcf1461013557600080fd5b806309bd5a60146100ce57806354fd4d50146100ea575b600080fd5b6100d760025481565b6040519081526020015b60405180910390f35b6100f26101d9565b6040516100e1919061052a565b6100d760015481565b60035461011c9067ffffffffffffffff1681565b60405167ffffffffffffffff90911681526020016100e1565b610148610143366004610598565b61027c565b005b60005461011c9067ffffffffffffffff1681565b6100d760055481565b6100d760065481565b60005461011c9068010000000000000000900467ffffffffffffffff1681565b6101ab73deaddeaddeaddeaddeaddeaddeaddeaddead000181565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020016100e1565b6100d760045481565b60606102047f00000000000000000000000000000000000000000000000000000000000000006103bd565b61022d7f00000000000000000000000000000000000000000000000000000000000000006103bd565b6102567f00000000000000000000000000000000000000000000000000000000000000006103bd565b6040516020016102689392919061066e565b604051602081830303815290604052905090565b3373deaddeaddeaddeaddeaddeaddeaddeaddead000114610323576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603b60248201527f4c31426c6f636b3a206f6e6c7920746865206465706f7369746f72206163636f60448201527f756e742063616e20736574204c3120626c6f636b2076616c7565730000000000606482015260840160405180910390fd5b50506000805467ffffffffffffffff98891668010000000000000000027fffffffffffffffffffffffffffffffff00000000000000000000000000000000909116998916999099179890981790975560019490945560029290925560038054919094167fffffffffffffffffffffffffffffffffffffffffffffffff00000000000000009190911617909255600491909155600555600655565b60608160000361040057505060408051808201909152600181527f3000000000000000000000000000000000000000000000000000000000000000602082015290565b8160005b811561042a578061041481610713565b91506104239050600a8361077a565b9150610404565b60008167ffffffffffffffff8111156104455761044561078e565b6040519080825280601f01601f19166020018201604052801561046f576020820181803683370190505b5090505b84156104f2576104846001836107bd565b9150610491600a866107d4565b61049c9060306107e8565b60f81b8183815181106104b1576104b1610800565b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a9053506104eb600a8661077a565b9450610473565b949350505050565b60005b838110156105155781810151838201526020016104fd565b83811115610524576000848401525b50505050565b60208152600082518060208401526105498160408501602087016104fa565b601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169190910160400192915050565b803567ffffffffffffffff8116811461059357600080fd5b919050565b6000806000806000806000806000806101208b8d0312156105b857600080fd5b6105c18b61057b565b99506105cf60208c0161057b565b985060408b0135975060608b013596506105eb60808c0161057b565b955060a08b0135945060c08b0135935060e08b013592506101008b013567ffffffffffffffff8082111561061e57600080fd5b818d0191508d601f83011261063257600080fd5b81358181111561064157600080fd5b8e602082850101111561065357600080fd5b6020830194508093505050509295989b9194979a5092959850565b600084516106808184602089016104fa565b80830190507f2e0000000000000000000000000000000000000000000000000000000000000080825285516106bc816001850160208a016104fa565b600192019182015283516106d78160028401602088016104fa565b0160020195945050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8203610744576107446106e4565b5060010190565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b6000826107895761078961074b565b500490565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6000828210156107cf576107cf6106e4565b500390565b6000826107e3576107e361074b565b500690565b600082198211156107fb576107fb6106e4565b500190565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fdfea164736f6c634300080f000a"

func init() {
	if err := json.Unmarshal([]byte(L1BlockStorageLayoutJSON), L1BlockStorageLayout); err != nil {
		panic(err)
	}

	layouts["L1Block"] = L1BlockStorageLayout
	deployedBytecodes["L1Block"] = L1BlockDeployedBin
}

// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"encoding/json"

	"github.com/ethereum-optimism/optimism/op-bindings/solc"
)

const L1BlockStorageLayoutJSON = "{\"storage\":[{\"astId\":1000,\"contract\":\"src/L2/L1Block.sol:L1Block\",\"label\":\"number\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_uint64\"},{\"astId\":1001,\"contract\":\"src/L2/L1Block.sol:L1Block\",\"label\":\"timestamp\",\"offset\":8,\"slot\":\"0\",\"type\":\"t_uint64\"},{\"astId\":1002,\"contract\":\"src/L2/L1Block.sol:L1Block\",\"label\":\"basefee\",\"offset\":0,\"slot\":\"1\",\"type\":\"t_uint256\"},{\"astId\":1003,\"contract\":\"src/L2/L1Block.sol:L1Block\",\"label\":\"hash\",\"offset\":0,\"slot\":\"2\",\"type\":\"t_bytes32\"},{\"astId\":1004,\"contract\":\"src/L2/L1Block.sol:L1Block\",\"label\":\"sequenceNumber\",\"offset\":0,\"slot\":\"3\",\"type\":\"t_uint64\"},{\"astId\":1005,\"contract\":\"src/L2/L1Block.sol:L1Block\",\"label\":\"blobBaseFeeScalar\",\"offset\":8,\"slot\":\"3\",\"type\":\"t_uint32\"},{\"astId\":1006,\"contract\":\"src/L2/L1Block.sol:L1Block\",\"label\":\"baseFeeScalar\",\"offset\":12,\"slot\":\"3\",\"type\":\"t_uint32\"},{\"astId\":1007,\"contract\":\"src/L2/L1Block.sol:L1Block\",\"label\":\"batcherHash\",\"offset\":0,\"slot\":\"4\",\"type\":\"t_bytes32\"},{\"astId\":1008,\"contract\":\"src/L2/L1Block.sol:L1Block\",\"label\":\"l1FeeOverhead\",\"offset\":0,\"slot\":\"5\",\"type\":\"t_uint256\"},{\"astId\":1009,\"contract\":\"src/L2/L1Block.sol:L1Block\",\"label\":\"l1FeeScalar\",\"offset\":0,\"slot\":\"6\",\"type\":\"t_uint256\"},{\"astId\":1010,\"contract\":\"src/L2/L1Block.sol:L1Block\",\"label\":\"blobBaseFee\",\"offset\":0,\"slot\":\"7\",\"type\":\"t_uint256\"},{\"astId\":1011,\"contract\":\"src/L2/L1Block.sol:L1Block\",\"label\":\"espresso\",\"offset\":0,\"slot\":\"8\",\"type\":\"t_bool\"},{\"astId\":1012,\"contract\":\"src/L2/L1Block.sol:L1Block\",\"label\":\"espressoL1ConfDepth\",\"offset\":1,\"slot\":\"8\",\"type\":\"t_uint64\"}],\"types\":{\"t_bool\":{\"encoding\":\"inplace\",\"label\":\"bool\",\"numberOfBytes\":\"1\"},\"t_bytes32\":{\"encoding\":\"inplace\",\"label\":\"bytes32\",\"numberOfBytes\":\"32\"},\"t_uint256\":{\"encoding\":\"inplace\",\"label\":\"uint256\",\"numberOfBytes\":\"32\"},\"t_uint32\":{\"encoding\":\"inplace\",\"label\":\"uint32\",\"numberOfBytes\":\"4\"},\"t_uint64\":{\"encoding\":\"inplace\",\"label\":\"uint64\",\"numberOfBytes\":\"8\"}}}"

var L1BlockStorageLayout = new(solc.StorageLayout)

var L1BlockDeployedBin = "0x608060405234801561001057600080fd5b506004361061011b5760003560e01c80638381f58a116100b2578063c598591811610081578063e591b28211610066578063e591b282146102a5578063e81b2c6d146102e5578063f8206140146102ee57600080fd5b8063c59859181461026c578063dc59462c1461028c57600080fd5b80638381f58a146102265780638b239f731461023a5780639e8c496614610243578063b80777ea1461024c57600080fd5b80635cf24969116100ee5780635cf24969146101a257806361fba0ca146101ab57806364ca23ef146101c857806368d5dca6146101f557600080fd5b806309bd5a6014610120578063440a5e201461013c57806354b7325c1461014657806354fd4d5014610159575b600080fd5b61012960025481565b6040519081526020015b60405180910390f35b6101446102f7565b005b610144610154366004610587565b610355565b6101956040518060400160405280600581526020017f312e322e3000000000000000000000000000000000000000000000000000000081525081565b60405161013391906105ca565b61012960015481565b6008546101b89060ff1681565b6040519015158152602001610133565b6003546101dc9067ffffffffffffffff1681565b60405167ffffffffffffffff9091168152602001610133565b6003546102119068010000000000000000900463ffffffff1681565b60405163ffffffff9091168152602001610133565b6000546101dc9067ffffffffffffffff1681565b61012960055481565b61012960065481565b6000546101dc9068010000000000000000900467ffffffffffffffff1681565b600354610211906c01000000000000000000000000900463ffffffff1681565b6008546101dc90610100900467ffffffffffffffff1681565b6102c073deaddeaddeaddeaddeaddeaddeaddeaddead000181565b60405173ffffffffffffffffffffffffffffffffffffffff9091168152602001610133565b61012960045481565b61012960075481565b3373deaddeaddeaddeaddeaddeaddeaddeaddead00011461032057633cc50b456000526004601cfd5b60043560801c60035560143560801c60005560243560015560443560075560643560025560843560045560a43560801c600855565b3373deaddeaddeaddeaddeaddeaddeaddeaddead0001146103fc576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603b60248201527f4c31426c6f636b3a206f6e6c7920746865206465706f7369746f72206163636f60448201527f756e742063616e20736574204c3120626c6f636b2076616c7565730000000000606482015260840160405180910390fd5b610409602082018261063d565b600080547fffffffffffffffffffffffffffffffffffffffffffffffff00000000000000001667ffffffffffffffff92909216919091179055610452604082016020830161063d565b6000805467ffffffffffffffff9290921668010000000000000000027fffffffffffffffffffffffffffffffff0000000000000000ffffffffffffffff909216919091179055604081013560015560608101356002556104b860a082016080830161063d565b600380547fffffffffffffffffffffffffffffffffffffffffffffffff00000000000000001667ffffffffffffffff9290921691909117905560a081013560045560c081013560055560e081013560065561051b61012082016101008301610667565b600880547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff001691151591909117905561055c6101408201610120830161063d565b600860016101000a81548167ffffffffffffffff021916908367ffffffffffffffff16021790555050565b60006020828403121561059957600080fd5b813567ffffffffffffffff8111156105b057600080fd5b820161016081850312156105c357600080fd5b9392505050565b600060208083528351808285015260005b818110156105f7578581018301518582016040015282016105db565b81811115610609576000604083870101525b50601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016929092016040019392505050565b60006020828403121561064f57600080fd5b813567ffffffffffffffff811681146105c357600080fd5b60006020828403121561067957600080fd5b813580151581146105c357600080fdfea164736f6c634300080f000a"


func init() {
	if err := json.Unmarshal([]byte(L1BlockStorageLayoutJSON), L1BlockStorageLayout); err != nil {
		panic(err)
	}

	layouts["L1Block"] = L1BlockStorageLayout
	deployedBytecodes["L1Block"] = L1BlockDeployedBin
	immutableReferences["L1Block"] = false
}

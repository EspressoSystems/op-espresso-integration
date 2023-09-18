// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// ResourceMeteringResourceConfig is an auto generated low-level Go binding around an user-defined struct.
type ResourceMeteringResourceConfig struct {
	MaxResourceLimit            uint32
	ElasticityMultiplier        uint8
	BaseFeeMaxChangeDenominator uint8
	MinimumBaseFee              uint32
	SystemTxMaxGas              uint32
	MaximumBaseFee              *big.Int
}

// SystemConfigAddresses is an auto generated low-level Go binding around an user-defined struct.
type SystemConfigAddresses struct {
	L1CrossDomainMessenger       common.Address
	L1ERC721Bridge               common.Address
	L1StandardBridge             common.Address
	L2OutputOracle               common.Address
	OptimismPortal               common.Address
	OptimismMintableERC20Factory common.Address
}

// SystemConfigMetaData contains all meta data concerning the SystemConfig contract.
var SystemConfigMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"enumSystemConfig.UpdateType\",\"name\":\"updateType\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"ConfigUpdate\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BATCH_INBOX_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"L1_CROSS_DOMAIN_MESSENGER_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"L1_ERC_721_BRIDGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"L1_STANDARD_BRIDGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"L2_OUTPUT_ORACLE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"OPTIMISM_MINTABLE_ERC20_FACTORY_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"OPTIMISM_PORTAL_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"UNSAFE_BLOCK_SIGNER_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VERSION\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"batchInbox\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"addr_\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"batcherHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"espresso\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"gasLimit\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_overhead\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_scalar\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"_batcherHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"_gasLimit\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"_espresso\",\"type\":\"bool\"},{\"internalType\":\"address\",\"name\":\"_unsafeBlockSigner\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"uint32\",\"name\":\"maxResourceLimit\",\"type\":\"uint32\"},{\"internalType\":\"uint8\",\"name\":\"elasticityMultiplier\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"baseFeeMaxChangeDenominator\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"minimumBaseFee\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"systemTxMaxGas\",\"type\":\"uint32\"},{\"internalType\":\"uint128\",\"name\":\"maximumBaseFee\",\"type\":\"uint128\"}],\"internalType\":\"structResourceMetering.ResourceConfig\",\"name\":\"_config\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"_startBlock\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_batchInbox\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"l1CrossDomainMessenger\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"l1ERC721Bridge\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"l1StandardBridge\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"l2OutputOracle\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"optimismPortal\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"optimismMintableERC20Factory\",\"type\":\"address\"}],\"internalType\":\"structSystemConfig.Addresses\",\"name\":\"_addresses\",\"type\":\"tuple\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l1CrossDomainMessenger\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"addr_\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l1ERC721Bridge\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"addr_\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l1StandardBridge\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"addr_\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l2OutputOracle\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"addr_\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minimumGasLimit\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"optimismMintableERC20Factory\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"addr_\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"optimismPortal\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"addr_\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"overhead\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"resourceConfig\",\"outputs\":[{\"components\":[{\"internalType\":\"uint32\",\"name\":\"maxResourceLimit\",\"type\":\"uint32\"},{\"internalType\":\"uint8\",\"name\":\"elasticityMultiplier\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"baseFeeMaxChangeDenominator\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"minimumBaseFee\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"systemTxMaxGas\",\"type\":\"uint32\"},{\"internalType\":\"uint128\",\"name\":\"maximumBaseFee\",\"type\":\"uint128\"}],\"internalType\":\"structResourceMetering.ResourceConfig\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"scalar\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_batcherHash\",\"type\":\"bytes32\"}],\"name\":\"setBatcherHash\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"_espresso\",\"type\":\"bool\"}],\"name\":\"setEspresso\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_overhead\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_scalar\",\"type\":\"uint256\"}],\"name\":\"setGasConfig\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"_gasLimit\",\"type\":\"uint64\"}],\"name\":\"setGasLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint32\",\"name\":\"maxResourceLimit\",\"type\":\"uint32\"},{\"internalType\":\"uint8\",\"name\":\"elasticityMultiplier\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"baseFeeMaxChangeDenominator\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"minimumBaseFee\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"systemTxMaxGas\",\"type\":\"uint32\"},{\"internalType\":\"uint128\",\"name\":\"maximumBaseFee\",\"type\":\"uint128\"}],\"internalType\":\"structResourceMetering.ResourceConfig\",\"name\":\"_config\",\"type\":\"tuple\"}],\"name\":\"setResourceConfig\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_unsafeBlockSigner\",\"type\":\"address\"}],\"name\":\"setUnsafeBlockSigner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"startBlock\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unsafeBlockSigner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"addr_\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x60806040523480156200001157600080fd5b506040805160c080820183526001808352602080840182905260028486015260006060808601829052608080870183905260a08088018490528851968701895283875293860183905296850182905284018190529483018590528201849052620000919361dead9390928392839290918391829160001990839062000097565b62000cf2565b600054600290610100900460ff16158015620000ba575060005460ff8083169116105b620001235760405162461bcd60e51b815260206004820152602e60248201527f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160448201526d191e481a5b9a5d1a585b1a5e995960921b60648201526084015b60405180910390fd5b6000805461ffff191660ff83161761010017905562000141620003c1565b6200014c8c62000429565b6200015789620004a8565b620001638b8b620004fa565b6200016e886200055e565b6200017986620005fb565b620001848762000655565b620001b983620001b660017f71ac12829d66ee73d8d95bff50b3589745ce57edae70a3fb111a2342464dc59862000bcf565b55565b8151620001ed90620001b660017f383f291819e6d54073bc9a648251d97421076bdd101933c0c022219ce958063762000bcf565b60208201516200022490620001b660017f46adcbebc6be8ce551740c29c47c8798210f23f7f4086c41752944352568d5a862000bcf565b60408201516200025b90620001b660017f9904ba90dde5696cda05c9e0dab5cbaa0fea005ace4d11218a02ac668dad637762000bcf565b60608201516200029290620001b660017fe52a667f71ec761b9b381c7b76ca9b852adf7e8905da0e0ad49986a0a687181662000bcf565b6080820151620002c990620001b660017f4b6c74f9e688cb39801f2112c14a8c57232a3fc5202e1444126d4bce86eb19ad62000bcf565b60a08201516200030090620001b660017fa04c5bb938ca6fc46d95553abf0a76345ce3e722a30bf4f74928b8e7d852320d62000bcf565b6200030b84620006a8565b620003168562000733565b6200032062000a77565b6001600160401b0316886001600160401b03161015620003725760405162461bcd60e51b815260206004820152601f6024820152600080516020620029f883398151915260448201526064016200011a565b6000805461ff001916905560405160ff821681527f7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb38474024989060200160405180910390a1505050505050505050505050565b600054610100900460ff166200041d5760405162461bcd60e51b815260206004820152602b602482015260008051602062002a3883398151915260448201526a6e697469616c697a696e6760a81b60648201526084016200011a565b6200042762000aa4565b565b6200043362000b0b565b6001600160a01b0381166200049a5760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b60648201526084016200011a565b620004a58162000b67565b50565b60678190556040805160208082018490528251808303909101815290820190915260005b600060008051602062002a1883398151915283604051620004ee919062000be9565b60405180910390a35050565b60658290556066819055604080516020810184905290810182905260009060600160408051601f1981840301815291905290506001600060008051602062002a188339815191528360405162000551919062000be9565b60405180910390a3505050565b6200056862000a77565b6001600160401b0316816001600160401b03161015620005ba5760405162461bcd60e51b815260206004820152601f6024820152600080516020620029f883398151915260448201526064016200011a565b606880546001600160401b0319166001600160401b0383169081179091556040805160208082019390935281518082039093018352810190526002620004cc565b62000624817f65a7ed542fb37fe237fdfbdd70b31598523fe5b32879e307bae27a0bd9581c0855565b604080516001600160a01b03831660208201526000910160408051601f1981840301815291905290506003620004cc565b60688054821515680100000000000000000260ff60401b199091161790556040516000906200068e908390602001901515815260200190565b60408051601f1981840301815291905290506004620004cc565b606a5415620007205760405162461bcd60e51b815260206004820152603860248201527f53797374656d436f6e6669673a2063616e6e6f74206f7665727269646520616e60448201527f20616c72656164792073657420737461727420626c6f636b000000000000000060648201526084016200011a565b80156200072c57606a55565b43606a5550565b8060a001516001600160801b0316816060015163ffffffff161115620007c25760405162461bcd60e51b815260206004820152603560248201527f53797374656d436f6e6669673a206d696e206261736520666565206d7573742060448201527f6265206c657373207468616e206d61782062617365000000000000000000000060648201526084016200011a565b6001816040015160ff1611620008335760405162461bcd60e51b815260206004820152602f60248201527f53797374656d436f6e6669673a2064656e6f6d696e61746f72206d757374206260448201526e65206c6172676572207468616e203160881b60648201526084016200011a565b606854608082015182516001600160401b039092169162000855919062000c41565b63ffffffff1611156200089a5760405162461bcd60e51b815260206004820152601f6024820152600080516020620029f883398151915260448201526064016200011a565b6000816020015160ff16116200090b5760405162461bcd60e51b815260206004820152602f60248201527f53797374656d436f6e6669673a20656c6173746963697479206d756c7469706c60448201526e06965722063616e6e6f74206265203608c1b60648201526084016200011a565b8051602082015163ffffffff82169160ff909116906200092d90829062000c6c565b62000939919062000c9e565b63ffffffff1614620009b45760405162461bcd60e51b815260206004820152603760248201527f53797374656d436f6e6669673a20707265636973696f6e206c6f73732077697460448201527f6820746172676574207265736f75726365206c696d697400000000000000000060648201526084016200011a565b805160698054602084015160408501516060860151608087015160a09097015163ffffffff96871664ffffffffff199095169490941764010000000060ff948516021764ffffffffff60281b191665010000000000939092169290920263ffffffff60301b19161766010000000000009185169190910217600160501b600160f01b0319166a01000000000000000000009390941692909202600160701b600160f01b03191692909217600160701b6001600160801b0390921691909102179055565b60695460009062000a9f9063ffffffff6a010000000000000000000082048116911662000ccd565b905090565b600054610100900460ff1662000b005760405162461bcd60e51b815260206004820152602b602482015260008051602062002a3883398151915260448201526a6e697469616c697a696e6760a81b60648201526084016200011a565b620004273362000b67565b6033546001600160a01b03163314620004275760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e657260448201526064016200011a565b603380546001600160a01b038381166001600160a01b0319831681179093556040519116919082907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a35050565b634e487b7160e01b600052601160045260246000fd5b60008282101562000be45762000be462000bb9565b500390565b600060208083528351808285015260005b8181101562000c185785810183015185820160400152820162000bfa565b8181111562000c2b576000604083870101525b50601f01601f1916929092016040019392505050565b600063ffffffff80831681851680830382111562000c635762000c6362000bb9565b01949350505050565b600063ffffffff8084168062000c9257634e487b7160e01b600052601260045260246000fd5b92169190910492915050565b600063ffffffff8083168185168183048111821515161562000cc45762000cc462000bb9565b02949350505050565b60006001600160401b0382811684821680830382111562000c635762000c6362000bb9565b611cf68062000d026000396000f3fe608060405234801561001057600080fd5b506004361061025c5760003560e01c8063935f029e11610145578063dac6e63a116100bd578063f45e65d81161008c578063f8c68de011610071578063f8c68de0146105e7578063fd32aa0f146105ef578063ffa1ad74146105f757600080fd5b8063f45e65d8146105ca578063f68016b7146105d357600080fd5b8063dac6e63a14610593578063e81b2c6d1461059b578063f06d7a3b146105a4578063f2fde38b146105b757600080fd5b8063bc49ce5f11610114578063c71973f6116100f9578063c71973f614610439578063c9b26f611461044c578063cc731b021461045f57600080fd5b8063bc49ce5f14610429578063c4e8ddfa1461043157600080fd5b8063935f029e146103f35780639b7d7f0a14610406578063a71198691461040e578063b40a817c1461041657600080fd5b80634add321d116101d85780635d73369c116101a757806361fba0ca1161018c57806361fba0ca146103a4578063715018a6146103cd5780638da5cb5b146103d557600080fd5b80635d73369c1461039457806361d157681461039c57600080fd5b80634add321d146102fb5780634d9f15591461031c5780634f16540b1461032457806354fd4d501461034b57600080fd5b806318d139181161022f5780631fd19ee1116102145780631fd19ee1146102d757806336eeb0e6146102df57806348cd4cb1146102f257600080fd5b806318d13918146102ba57806319f5cea8146102cf57600080fd5b806306c9265714610261578063078f29cf1461027c5780630a49cb03146102a95780630c18c162146102b1575b600080fd5b6102696105ff565b6040519081526020015b60405180910390f35b61028461062d565b60405173ffffffffffffffffffffffffffffffffffffffff9091168152602001610273565b610284610666565b61026960655481565b6102cd6102c83660046117ee565b610696565b005b6102696106aa565b6102846106d5565b6102cd6102ed36600461198e565b6106ff565b610269606a5481565b610303610a8c565b60405167ffffffffffffffff9091168152602001610273565b610284610ab2565b6102697f65a7ed542fb37fe237fdfbdd70b31598523fe5b32879e307bae27a0bd9581c0881565b6103876040518060400160405280600581526020017f312e372e3000000000000000000000000000000000000000000000000000000081525081565b6040516102739190611b42565b610269610ae2565b610269610b0d565b6068546103bd9068010000000000000000900460ff1681565b6040519015158152602001610273565b6102cd610b38565b60335473ffffffffffffffffffffffffffffffffffffffff16610284565b6102cd610401366004611b55565b610b4c565b610284610b62565b610284610b92565b6102cd610424366004611b77565b610bc2565b610269610bd3565b610284610bfe565b6102cd610447366004611b92565b610c2e565b6102cd61045a366004611bae565b610c3f565b6105236040805160c081018252600080825260208201819052918101829052606081018290526080810182905260a0810191909152506040805160c08101825260695463ffffffff8082168352640100000000820460ff9081166020850152650100000000008304169383019390935266010000000000008104831660608301526a0100000000000000000000810490921660808201526e0100000000000000000000000000009091046fffffffffffffffffffffffffffffffff1660a082015290565b6040516102739190600060c08201905063ffffffff80845116835260ff602085015116602084015260ff6040850151166040840152806060850151166060840152806080850151166080840152506fffffffffffffffffffffffffffffffff60a08401511660a083015292915050565b610284610c50565b61026960675481565b6102cd6105b2366004611bc7565b610c80565b6102cd6105c53660046117ee565b610c91565b61026960665481565b6068546103039067ffffffffffffffff1681565b610269610d45565b610269610d70565b610269600081565b61062a60017fa04c5bb938ca6fc46d95553abf0a76345ce3e722a30bf4f74928b8e7d852320d611c11565b81565b600061066161065d60017f9904ba90dde5696cda05c9e0dab5cbaa0fea005ace4d11218a02ac668dad6377611c11565b5490565b905090565b600061066161065d60017f4b6c74f9e688cb39801f2112c14a8c57232a3fc5202e1444126d4bce86eb19ad611c11565b61069e610d9b565b6106a781610e1c565b50565b61062a60017f46adcbebc6be8ce551740c29c47c8798210f23f7f4086c41752944352568d5a8611c11565b60006106617f65a7ed542fb37fe237fdfbdd70b31598523fe5b32879e307bae27a0bd9581c085490565b600054600290610100900460ff16158015610721575060005460ff8083169116105b6107b2576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602e60248201527f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160448201527f647920696e697469616c697a656400000000000000000000000000000000000060648201526084015b60405180910390fd5b600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00001660ff8316176101001790556107eb610ed8565b6107f48c610c91565b6107fd89610f77565b6108078b8b610f9f565b61081088611030565b61081986610e1c565b6108228761110e565b6108548361085160017f71ac12829d66ee73d8d95bff50b3589745ce57edae70a3fb111a2342464dc598611c11565b55565b81516108859061085160017f383f291819e6d54073bc9a648251d97421076bdd101933c0c022219ce9580637611c11565b60208201516108b99061085160017f46adcbebc6be8ce551740c29c47c8798210f23f7f4086c41752944352568d5a8611c11565b60408201516108ed9061085160017f9904ba90dde5696cda05c9e0dab5cbaa0fea005ace4d11218a02ac668dad6377611c11565b60608201516109219061085160017fe52a667f71ec761b9b381c7b76ca9b852adf7e8905da0e0ad49986a0a6871816611c11565b60808201516109559061085160017f4b6c74f9e688cb39801f2112c14a8c57232a3fc5202e1444126d4bce86eb19ad611c11565b60a08201516109899061085160017fa04c5bb938ca6fc46d95553abf0a76345ce3e722a30bf4f74928b8e7d852320d611c11565b61099284611198565b61099b8561123a565b6109a3610a8c565b67ffffffffffffffff168867ffffffffffffffff161015610a20576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f53797374656d436f6e6669673a20676173206c696d697420746f6f206c6f770060448201526064016107a9565b600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff16905560405160ff821681527f7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb38474024989060200160405180910390a1505050505050505050505050565b6069546000906106619063ffffffff6a0100000000000000000000820481169116611c28565b600061066161065d60017fe52a667f71ec761b9b381c7b76ca9b852adf7e8905da0e0ad49986a0a6871816611c11565b61062a60017f383f291819e6d54073bc9a648251d97421076bdd101933c0c022219ce9580637611c11565b61062a60017fe52a667f71ec761b9b381c7b76ca9b852adf7e8905da0e0ad49986a0a6871816611c11565b610b40610d9b565b610b4a60006116ae565b565b610b54610d9b565b610b5e8282610f9f565b5050565b600061066161065d60017fa04c5bb938ca6fc46d95553abf0a76345ce3e722a30bf4f74928b8e7d852320d611c11565b600061066161065d60017f383f291819e6d54073bc9a648251d97421076bdd101933c0c022219ce9580637611c11565b610bca610d9b565b6106a781611030565b61062a60017f71ac12829d66ee73d8d95bff50b3589745ce57edae70a3fb111a2342464dc598611c11565b600061066161065d60017f46adcbebc6be8ce551740c29c47c8798210f23f7f4086c41752944352568d5a8611c11565b610c36610d9b565b6106a78161123a565b610c47610d9b565b6106a781610f77565b600061066161065d60017f71ac12829d66ee73d8d95bff50b3589745ce57edae70a3fb111a2342464dc598611c11565b610c88610d9b565b6106a78161110e565b610c99610d9b565b73ffffffffffffffffffffffffffffffffffffffff8116610d3c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201527f646472657373000000000000000000000000000000000000000000000000000060648201526084016107a9565b6106a7816116ae565b61062a60017f9904ba90dde5696cda05c9e0dab5cbaa0fea005ace4d11218a02ac668dad6377611c11565b61062a60017f4b6c74f9e688cb39801f2112c14a8c57232a3fc5202e1444126d4bce86eb19ad611c11565b60335473ffffffffffffffffffffffffffffffffffffffff163314610b4a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e657260448201526064016107a9565b610e44817f65a7ed542fb37fe237fdfbdd70b31598523fe5b32879e307bae27a0bd9581c0855565b6040805173ffffffffffffffffffffffffffffffffffffffff8316602082015260009101604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0818403018152919052905060035b60007f1d2b0bda21d56b8bd12d4f94ebacffdfb35f5e226f84b461103bb8beab6353be83604051610ecc9190611b42565b60405180910390a35050565b600054610100900460ff16610f6f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602b60248201527f496e697469616c697a61626c653a20636f6e7472616374206973206e6f74206960448201527f6e697469616c697a696e6700000000000000000000000000000000000000000060648201526084016107a9565b610b4a611725565b6067819055604080516020808201849052825180830390910181529082019091526000610e9b565b606582905560668190556040805160208101849052908101829052600090606001604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08184030181529190529050600160007f1d2b0bda21d56b8bd12d4f94ebacffdfb35f5e226f84b461103bb8beab6353be836040516110239190611b42565b60405180910390a3505050565b611038610a8c565b67ffffffffffffffff168167ffffffffffffffff1610156110b5576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f53797374656d436f6e6669673a20676173206c696d697420746f6f206c6f770060448201526064016107a9565b606880547fffffffffffffffffffffffffffffffffffffffffffffffff00000000000000001667ffffffffffffffff83169081179091556040805160208082019390935281518082039093018352810190526002610e9b565b6068805482151568010000000000000000027fffffffffffffffffffffffffffffffffffffffffffffff00ffffffffffffffff909116179055604051600090611161908390602001901515815260200190565b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe081840301815291905290506004610e9b565b606a5415611228576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603860248201527f53797374656d436f6e6669673a2063616e6e6f74206f7665727269646520616e60448201527f20616c72656164792073657420737461727420626c6f636b000000000000000060648201526084016107a9565b801561123357606a55565b43606a5550565b8060a001516fffffffffffffffffffffffffffffffff16816060015163ffffffff1611156112ea576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603560248201527f53797374656d436f6e6669673a206d696e206261736520666565206d7573742060448201527f6265206c657373207468616e206d61782062617365000000000000000000000060648201526084016107a9565b6001816040015160ff1611611381576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602f60248201527f53797374656d436f6e6669673a2064656e6f6d696e61746f72206d757374206260448201527f65206c6172676572207468616e2031000000000000000000000000000000000060648201526084016107a9565b6068546080820151825167ffffffffffffffff909216916113a29190611c54565b63ffffffff161115611410576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f53797374656d436f6e6669673a20676173206c696d697420746f6f206c6f770060448201526064016107a9565b6000816020015160ff16116114a7576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602f60248201527f53797374656d436f6e6669673a20656c6173746963697479206d756c7469706c60448201527f6965722063616e6e6f742062652030000000000000000000000000000000000060648201526084016107a9565b8051602082015163ffffffff82169160ff909116906114c7908290611c73565b6114d19190611cbd565b63ffffffff1614611564576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603760248201527f53797374656d436f6e6669673a20707265636973696f6e206c6f73732077697460448201527f6820746172676574207265736f75726365206c696d697400000000000000000060648201526084016107a9565b805160698054602084015160408501516060860151608087015160a09097015163ffffffff9687167fffffffffffffffffffffffffffffffffffffffffffffffffffffff00000000009095169490941764010000000060ff94851602177fffffffffffffffffffffffffffffffffffffffffffff0000000000ffffffffff166501000000000093909216929092027fffffffffffffffffffffffffffffffffffffffffffff00000000ffffffffffff1617660100000000000091851691909102177fffff0000000000000000000000000000000000000000ffffffffffffffffffff166a010000000000000000000093909416929092027fffff00000000000000000000000000000000ffffffffffffffffffffffffffff16929092176e0100000000000000000000000000006fffffffffffffffffffffffffffffffff90921691909102179055565b6033805473ffffffffffffffffffffffffffffffffffffffff8381167fffffffffffffffffffffffff0000000000000000000000000000000000000000831681179093556040519116919082907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a35050565b600054610100900460ff166117bc576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602b60248201527f496e697469616c697a61626c653a20636f6e7472616374206973206e6f74206960448201527f6e697469616c697a696e6700000000000000000000000000000000000000000060648201526084016107a9565b610b4a336116ae565b803573ffffffffffffffffffffffffffffffffffffffff811681146117e957600080fd5b919050565b60006020828403121561180057600080fd5b611809826117c5565b9392505050565b803567ffffffffffffffff811681146117e957600080fd5b803580151581146117e957600080fd5b60405160c0810167ffffffffffffffff81118282101715611882577f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b60405290565b803563ffffffff811681146117e957600080fd5b803560ff811681146117e957600080fd5b600060c082840312156118bf57600080fd5b60405160c0810181811067ffffffffffffffff82111715611909577f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b60405290508061191883611888565b81526119266020840161189c565b60208201526119376040840161189c565b604082015261194860608401611888565b606082015261195960808401611888565b608082015260a08301356fffffffffffffffffffffffffffffffff8116811461198157600080fd5b60a0919091015292915050565b60008060008060008060008060008060008b8d036102a08112156119b157600080fd5b6119ba8d6117c5565b9b5060208d01359a5060408d0135995060608d013598506119dd60808e01611810565b97506119eb60a08e01611828565b96506119f960c08e016117c5565b9550611a088e60e08f016118ad565b94506101a08d01359350611a1f6101c08e016117c5565b925060c07ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe2082011215611a5157600080fd5b50611a5a611838565b611a676101e08e016117c5565b8152611a766102008e016117c5565b6020820152611a886102208e016117c5565b6040820152611a9a6102408e016117c5565b6060820152611aac6102608e016117c5565b6080820152611abe6102808e016117c5565b60a0820152809150509295989b509295989b9093969950565b6000815180845260005b81811015611afd57602081850181015186830182015201611ae1565b81811115611b0f576000602083870101525b50601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169290920160200192915050565b6020815260006118096020830184611ad7565b60008060408385031215611b6857600080fd5b50508035926020909101359150565b600060208284031215611b8957600080fd5b61180982611810565b600060c08284031215611ba457600080fd5b61180983836118ad565b600060208284031215611bc057600080fd5b5035919050565b600060208284031215611bd957600080fd5b61180982611828565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600082821015611c2357611c23611be2565b500390565b600067ffffffffffffffff808316818516808303821115611c4b57611c4b611be2565b01949350505050565b600063ffffffff808316818516808303821115611c4b57611c4b611be2565b600063ffffffff80841680611cb1577f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b92169190910492915050565b600063ffffffff80831681851681830481118215151615611ce057611ce0611be2565b0294935050505056fea164736f6c634300080f000a53797374656d436f6e6669673a20676173206c696d697420746f6f206c6f77001d2b0bda21d56b8bd12d4f94ebacffdfb35f5e226f84b461103bb8beab6353be496e697469616c697a61626c653a20636f6e7472616374206973206e6f742069",
}

// SystemConfigABI is the input ABI used to generate the binding from.
// Deprecated: Use SystemConfigMetaData.ABI instead.
var SystemConfigABI = SystemConfigMetaData.ABI

// SystemConfigBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SystemConfigMetaData.Bin instead.
var SystemConfigBin = SystemConfigMetaData.Bin

// DeploySystemConfig deploys a new Ethereum contract, binding an instance of SystemConfig to it.
func DeploySystemConfig(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SystemConfig, error) {
	parsed, err := SystemConfigMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SystemConfigBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SystemConfig{SystemConfigCaller: SystemConfigCaller{contract: contract}, SystemConfigTransactor: SystemConfigTransactor{contract: contract}, SystemConfigFilterer: SystemConfigFilterer{contract: contract}}, nil
}

// SystemConfig is an auto generated Go binding around an Ethereum contract.
type SystemConfig struct {
	SystemConfigCaller     // Read-only binding to the contract
	SystemConfigTransactor // Write-only binding to the contract
	SystemConfigFilterer   // Log filterer for contract events
}

// SystemConfigCaller is an auto generated read-only Go binding around an Ethereum contract.
type SystemConfigCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SystemConfigTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SystemConfigTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SystemConfigFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SystemConfigFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SystemConfigSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SystemConfigSession struct {
	Contract     *SystemConfig     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SystemConfigCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SystemConfigCallerSession struct {
	Contract *SystemConfigCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// SystemConfigTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SystemConfigTransactorSession struct {
	Contract     *SystemConfigTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// SystemConfigRaw is an auto generated low-level Go binding around an Ethereum contract.
type SystemConfigRaw struct {
	Contract *SystemConfig // Generic contract binding to access the raw methods on
}

// SystemConfigCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SystemConfigCallerRaw struct {
	Contract *SystemConfigCaller // Generic read-only contract binding to access the raw methods on
}

// SystemConfigTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SystemConfigTransactorRaw struct {
	Contract *SystemConfigTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSystemConfig creates a new instance of SystemConfig, bound to a specific deployed contract.
func NewSystemConfig(address common.Address, backend bind.ContractBackend) (*SystemConfig, error) {
	contract, err := bindSystemConfig(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SystemConfig{SystemConfigCaller: SystemConfigCaller{contract: contract}, SystemConfigTransactor: SystemConfigTransactor{contract: contract}, SystemConfigFilterer: SystemConfigFilterer{contract: contract}}, nil
}

// NewSystemConfigCaller creates a new read-only instance of SystemConfig, bound to a specific deployed contract.
func NewSystemConfigCaller(address common.Address, caller bind.ContractCaller) (*SystemConfigCaller, error) {
	contract, err := bindSystemConfig(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SystemConfigCaller{contract: contract}, nil
}

// NewSystemConfigTransactor creates a new write-only instance of SystemConfig, bound to a specific deployed contract.
func NewSystemConfigTransactor(address common.Address, transactor bind.ContractTransactor) (*SystemConfigTransactor, error) {
	contract, err := bindSystemConfig(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SystemConfigTransactor{contract: contract}, nil
}

// NewSystemConfigFilterer creates a new log filterer instance of SystemConfig, bound to a specific deployed contract.
func NewSystemConfigFilterer(address common.Address, filterer bind.ContractFilterer) (*SystemConfigFilterer, error) {
	contract, err := bindSystemConfig(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SystemConfigFilterer{contract: contract}, nil
}

// bindSystemConfig binds a generic wrapper to an already deployed contract.
func bindSystemConfig(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SystemConfigABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SystemConfig *SystemConfigRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SystemConfig.Contract.SystemConfigCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SystemConfig *SystemConfigRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SystemConfig.Contract.SystemConfigTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SystemConfig *SystemConfigRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SystemConfig.Contract.SystemConfigTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SystemConfig *SystemConfigCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SystemConfig.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SystemConfig *SystemConfigTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SystemConfig.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SystemConfig *SystemConfigTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SystemConfig.Contract.contract.Transact(opts, method, params...)
}

// BATCHINBOXSLOT is a free data retrieval call binding the contract method 0xbc49ce5f.
//
// Solidity: function BATCH_INBOX_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCaller) BATCHINBOXSLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "BATCH_INBOX_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BATCHINBOXSLOT is a free data retrieval call binding the contract method 0xbc49ce5f.
//
// Solidity: function BATCH_INBOX_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigSession) BATCHINBOXSLOT() ([32]byte, error) {
	return _SystemConfig.Contract.BATCHINBOXSLOT(&_SystemConfig.CallOpts)
}

// BATCHINBOXSLOT is a free data retrieval call binding the contract method 0xbc49ce5f.
//
// Solidity: function BATCH_INBOX_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCallerSession) BATCHINBOXSLOT() ([32]byte, error) {
	return _SystemConfig.Contract.BATCHINBOXSLOT(&_SystemConfig.CallOpts)
}

// L1CROSSDOMAINMESSENGERSLOT is a free data retrieval call binding the contract method 0x5d73369c.
//
// Solidity: function L1_CROSS_DOMAIN_MESSENGER_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCaller) L1CROSSDOMAINMESSENGERSLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "L1_CROSS_DOMAIN_MESSENGER_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// L1CROSSDOMAINMESSENGERSLOT is a free data retrieval call binding the contract method 0x5d73369c.
//
// Solidity: function L1_CROSS_DOMAIN_MESSENGER_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigSession) L1CROSSDOMAINMESSENGERSLOT() ([32]byte, error) {
	return _SystemConfig.Contract.L1CROSSDOMAINMESSENGERSLOT(&_SystemConfig.CallOpts)
}

// L1CROSSDOMAINMESSENGERSLOT is a free data retrieval call binding the contract method 0x5d73369c.
//
// Solidity: function L1_CROSS_DOMAIN_MESSENGER_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCallerSession) L1CROSSDOMAINMESSENGERSLOT() ([32]byte, error) {
	return _SystemConfig.Contract.L1CROSSDOMAINMESSENGERSLOT(&_SystemConfig.CallOpts)
}

// L1ERC721BRIDGESLOT is a free data retrieval call binding the contract method 0x19f5cea8.
//
// Solidity: function L1_ERC_721_BRIDGE_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCaller) L1ERC721BRIDGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "L1_ERC_721_BRIDGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// L1ERC721BRIDGESLOT is a free data retrieval call binding the contract method 0x19f5cea8.
//
// Solidity: function L1_ERC_721_BRIDGE_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigSession) L1ERC721BRIDGESLOT() ([32]byte, error) {
	return _SystemConfig.Contract.L1ERC721BRIDGESLOT(&_SystemConfig.CallOpts)
}

// L1ERC721BRIDGESLOT is a free data retrieval call binding the contract method 0x19f5cea8.
//
// Solidity: function L1_ERC_721_BRIDGE_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCallerSession) L1ERC721BRIDGESLOT() ([32]byte, error) {
	return _SystemConfig.Contract.L1ERC721BRIDGESLOT(&_SystemConfig.CallOpts)
}

// L1STANDARDBRIDGESLOT is a free data retrieval call binding the contract method 0xf8c68de0.
//
// Solidity: function L1_STANDARD_BRIDGE_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCaller) L1STANDARDBRIDGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "L1_STANDARD_BRIDGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// L1STANDARDBRIDGESLOT is a free data retrieval call binding the contract method 0xf8c68de0.
//
// Solidity: function L1_STANDARD_BRIDGE_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigSession) L1STANDARDBRIDGESLOT() ([32]byte, error) {
	return _SystemConfig.Contract.L1STANDARDBRIDGESLOT(&_SystemConfig.CallOpts)
}

// L1STANDARDBRIDGESLOT is a free data retrieval call binding the contract method 0xf8c68de0.
//
// Solidity: function L1_STANDARD_BRIDGE_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCallerSession) L1STANDARDBRIDGESLOT() ([32]byte, error) {
	return _SystemConfig.Contract.L1STANDARDBRIDGESLOT(&_SystemConfig.CallOpts)
}

// L2OUTPUTORACLESLOT is a free data retrieval call binding the contract method 0x61d15768.
//
// Solidity: function L2_OUTPUT_ORACLE_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCaller) L2OUTPUTORACLESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "L2_OUTPUT_ORACLE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// L2OUTPUTORACLESLOT is a free data retrieval call binding the contract method 0x61d15768.
//
// Solidity: function L2_OUTPUT_ORACLE_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigSession) L2OUTPUTORACLESLOT() ([32]byte, error) {
	return _SystemConfig.Contract.L2OUTPUTORACLESLOT(&_SystemConfig.CallOpts)
}

// L2OUTPUTORACLESLOT is a free data retrieval call binding the contract method 0x61d15768.
//
// Solidity: function L2_OUTPUT_ORACLE_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCallerSession) L2OUTPUTORACLESLOT() ([32]byte, error) {
	return _SystemConfig.Contract.L2OUTPUTORACLESLOT(&_SystemConfig.CallOpts)
}

// OPTIMISMMINTABLEERC20FACTORYSLOT is a free data retrieval call binding the contract method 0x06c92657.
//
// Solidity: function OPTIMISM_MINTABLE_ERC20_FACTORY_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCaller) OPTIMISMMINTABLEERC20FACTORYSLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "OPTIMISM_MINTABLE_ERC20_FACTORY_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// OPTIMISMMINTABLEERC20FACTORYSLOT is a free data retrieval call binding the contract method 0x06c92657.
//
// Solidity: function OPTIMISM_MINTABLE_ERC20_FACTORY_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigSession) OPTIMISMMINTABLEERC20FACTORYSLOT() ([32]byte, error) {
	return _SystemConfig.Contract.OPTIMISMMINTABLEERC20FACTORYSLOT(&_SystemConfig.CallOpts)
}

// OPTIMISMMINTABLEERC20FACTORYSLOT is a free data retrieval call binding the contract method 0x06c92657.
//
// Solidity: function OPTIMISM_MINTABLE_ERC20_FACTORY_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCallerSession) OPTIMISMMINTABLEERC20FACTORYSLOT() ([32]byte, error) {
	return _SystemConfig.Contract.OPTIMISMMINTABLEERC20FACTORYSLOT(&_SystemConfig.CallOpts)
}

// OPTIMISMPORTALSLOT is a free data retrieval call binding the contract method 0xfd32aa0f.
//
// Solidity: function OPTIMISM_PORTAL_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCaller) OPTIMISMPORTALSLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "OPTIMISM_PORTAL_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// OPTIMISMPORTALSLOT is a free data retrieval call binding the contract method 0xfd32aa0f.
//
// Solidity: function OPTIMISM_PORTAL_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigSession) OPTIMISMPORTALSLOT() ([32]byte, error) {
	return _SystemConfig.Contract.OPTIMISMPORTALSLOT(&_SystemConfig.CallOpts)
}

// OPTIMISMPORTALSLOT is a free data retrieval call binding the contract method 0xfd32aa0f.
//
// Solidity: function OPTIMISM_PORTAL_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCallerSession) OPTIMISMPORTALSLOT() ([32]byte, error) {
	return _SystemConfig.Contract.OPTIMISMPORTALSLOT(&_SystemConfig.CallOpts)
}

// UNSAFEBLOCKSIGNERSLOT is a free data retrieval call binding the contract method 0x4f16540b.
//
// Solidity: function UNSAFE_BLOCK_SIGNER_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCaller) UNSAFEBLOCKSIGNERSLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "UNSAFE_BLOCK_SIGNER_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// UNSAFEBLOCKSIGNERSLOT is a free data retrieval call binding the contract method 0x4f16540b.
//
// Solidity: function UNSAFE_BLOCK_SIGNER_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigSession) UNSAFEBLOCKSIGNERSLOT() ([32]byte, error) {
	return _SystemConfig.Contract.UNSAFEBLOCKSIGNERSLOT(&_SystemConfig.CallOpts)
}

// UNSAFEBLOCKSIGNERSLOT is a free data retrieval call binding the contract method 0x4f16540b.
//
// Solidity: function UNSAFE_BLOCK_SIGNER_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCallerSession) UNSAFEBLOCKSIGNERSLOT() ([32]byte, error) {
	return _SystemConfig.Contract.UNSAFEBLOCKSIGNERSLOT(&_SystemConfig.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_SystemConfig *SystemConfigCaller) VERSION(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_SystemConfig *SystemConfigSession) VERSION() (*big.Int, error) {
	return _SystemConfig.Contract.VERSION(&_SystemConfig.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_SystemConfig *SystemConfigCallerSession) VERSION() (*big.Int, error) {
	return _SystemConfig.Contract.VERSION(&_SystemConfig.CallOpts)
}

// BatchInbox is a free data retrieval call binding the contract method 0xdac6e63a.
//
// Solidity: function batchInbox() view returns(address addr_)
func (_SystemConfig *SystemConfigCaller) BatchInbox(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "batchInbox")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// BatchInbox is a free data retrieval call binding the contract method 0xdac6e63a.
//
// Solidity: function batchInbox() view returns(address addr_)
func (_SystemConfig *SystemConfigSession) BatchInbox() (common.Address, error) {
	return _SystemConfig.Contract.BatchInbox(&_SystemConfig.CallOpts)
}

// BatchInbox is a free data retrieval call binding the contract method 0xdac6e63a.
//
// Solidity: function batchInbox() view returns(address addr_)
func (_SystemConfig *SystemConfigCallerSession) BatchInbox() (common.Address, error) {
	return _SystemConfig.Contract.BatchInbox(&_SystemConfig.CallOpts)
}

// BatcherHash is a free data retrieval call binding the contract method 0xe81b2c6d.
//
// Solidity: function batcherHash() view returns(bytes32)
func (_SystemConfig *SystemConfigCaller) BatcherHash(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "batcherHash")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BatcherHash is a free data retrieval call binding the contract method 0xe81b2c6d.
//
// Solidity: function batcherHash() view returns(bytes32)
func (_SystemConfig *SystemConfigSession) BatcherHash() ([32]byte, error) {
	return _SystemConfig.Contract.BatcherHash(&_SystemConfig.CallOpts)
}

// BatcherHash is a free data retrieval call binding the contract method 0xe81b2c6d.
//
// Solidity: function batcherHash() view returns(bytes32)
func (_SystemConfig *SystemConfigCallerSession) BatcherHash() ([32]byte, error) {
	return _SystemConfig.Contract.BatcherHash(&_SystemConfig.CallOpts)
}

// Espresso is a free data retrieval call binding the contract method 0x61fba0ca.
//
// Solidity: function espresso() view returns(bool)
func (_SystemConfig *SystemConfigCaller) Espresso(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "espresso")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Espresso is a free data retrieval call binding the contract method 0x61fba0ca.
//
// Solidity: function espresso() view returns(bool)
func (_SystemConfig *SystemConfigSession) Espresso() (bool, error) {
	return _SystemConfig.Contract.Espresso(&_SystemConfig.CallOpts)
}

// Espresso is a free data retrieval call binding the contract method 0x61fba0ca.
//
// Solidity: function espresso() view returns(bool)
func (_SystemConfig *SystemConfigCallerSession) Espresso() (bool, error) {
	return _SystemConfig.Contract.Espresso(&_SystemConfig.CallOpts)
}

// GasLimit is a free data retrieval call binding the contract method 0xf68016b7.
//
// Solidity: function gasLimit() view returns(uint64)
func (_SystemConfig *SystemConfigCaller) GasLimit(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "gasLimit")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GasLimit is a free data retrieval call binding the contract method 0xf68016b7.
//
// Solidity: function gasLimit() view returns(uint64)
func (_SystemConfig *SystemConfigSession) GasLimit() (uint64, error) {
	return _SystemConfig.Contract.GasLimit(&_SystemConfig.CallOpts)
}

// GasLimit is a free data retrieval call binding the contract method 0xf68016b7.
//
// Solidity: function gasLimit() view returns(uint64)
func (_SystemConfig *SystemConfigCallerSession) GasLimit() (uint64, error) {
	return _SystemConfig.Contract.GasLimit(&_SystemConfig.CallOpts)
}

// L1CrossDomainMessenger is a free data retrieval call binding the contract method 0xa7119869.
//
// Solidity: function l1CrossDomainMessenger() view returns(address addr_)
func (_SystemConfig *SystemConfigCaller) L1CrossDomainMessenger(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "l1CrossDomainMessenger")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L1CrossDomainMessenger is a free data retrieval call binding the contract method 0xa7119869.
//
// Solidity: function l1CrossDomainMessenger() view returns(address addr_)
func (_SystemConfig *SystemConfigSession) L1CrossDomainMessenger() (common.Address, error) {
	return _SystemConfig.Contract.L1CrossDomainMessenger(&_SystemConfig.CallOpts)
}

// L1CrossDomainMessenger is a free data retrieval call binding the contract method 0xa7119869.
//
// Solidity: function l1CrossDomainMessenger() view returns(address addr_)
func (_SystemConfig *SystemConfigCallerSession) L1CrossDomainMessenger() (common.Address, error) {
	return _SystemConfig.Contract.L1CrossDomainMessenger(&_SystemConfig.CallOpts)
}

// L1ERC721Bridge is a free data retrieval call binding the contract method 0xc4e8ddfa.
//
// Solidity: function l1ERC721Bridge() view returns(address addr_)
func (_SystemConfig *SystemConfigCaller) L1ERC721Bridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "l1ERC721Bridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L1ERC721Bridge is a free data retrieval call binding the contract method 0xc4e8ddfa.
//
// Solidity: function l1ERC721Bridge() view returns(address addr_)
func (_SystemConfig *SystemConfigSession) L1ERC721Bridge() (common.Address, error) {
	return _SystemConfig.Contract.L1ERC721Bridge(&_SystemConfig.CallOpts)
}

// L1ERC721Bridge is a free data retrieval call binding the contract method 0xc4e8ddfa.
//
// Solidity: function l1ERC721Bridge() view returns(address addr_)
func (_SystemConfig *SystemConfigCallerSession) L1ERC721Bridge() (common.Address, error) {
	return _SystemConfig.Contract.L1ERC721Bridge(&_SystemConfig.CallOpts)
}

// L1StandardBridge is a free data retrieval call binding the contract method 0x078f29cf.
//
// Solidity: function l1StandardBridge() view returns(address addr_)
func (_SystemConfig *SystemConfigCaller) L1StandardBridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "l1StandardBridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L1StandardBridge is a free data retrieval call binding the contract method 0x078f29cf.
//
// Solidity: function l1StandardBridge() view returns(address addr_)
func (_SystemConfig *SystemConfigSession) L1StandardBridge() (common.Address, error) {
	return _SystemConfig.Contract.L1StandardBridge(&_SystemConfig.CallOpts)
}

// L1StandardBridge is a free data retrieval call binding the contract method 0x078f29cf.
//
// Solidity: function l1StandardBridge() view returns(address addr_)
func (_SystemConfig *SystemConfigCallerSession) L1StandardBridge() (common.Address, error) {
	return _SystemConfig.Contract.L1StandardBridge(&_SystemConfig.CallOpts)
}

// L2OutputOracle is a free data retrieval call binding the contract method 0x4d9f1559.
//
// Solidity: function l2OutputOracle() view returns(address addr_)
func (_SystemConfig *SystemConfigCaller) L2OutputOracle(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "l2OutputOracle")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L2OutputOracle is a free data retrieval call binding the contract method 0x4d9f1559.
//
// Solidity: function l2OutputOracle() view returns(address addr_)
func (_SystemConfig *SystemConfigSession) L2OutputOracle() (common.Address, error) {
	return _SystemConfig.Contract.L2OutputOracle(&_SystemConfig.CallOpts)
}

// L2OutputOracle is a free data retrieval call binding the contract method 0x4d9f1559.
//
// Solidity: function l2OutputOracle() view returns(address addr_)
func (_SystemConfig *SystemConfigCallerSession) L2OutputOracle() (common.Address, error) {
	return _SystemConfig.Contract.L2OutputOracle(&_SystemConfig.CallOpts)
}

// MinimumGasLimit is a free data retrieval call binding the contract method 0x4add321d.
//
// Solidity: function minimumGasLimit() view returns(uint64)
func (_SystemConfig *SystemConfigCaller) MinimumGasLimit(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "minimumGasLimit")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// MinimumGasLimit is a free data retrieval call binding the contract method 0x4add321d.
//
// Solidity: function minimumGasLimit() view returns(uint64)
func (_SystemConfig *SystemConfigSession) MinimumGasLimit() (uint64, error) {
	return _SystemConfig.Contract.MinimumGasLimit(&_SystemConfig.CallOpts)
}

// MinimumGasLimit is a free data retrieval call binding the contract method 0x4add321d.
//
// Solidity: function minimumGasLimit() view returns(uint64)
func (_SystemConfig *SystemConfigCallerSession) MinimumGasLimit() (uint64, error) {
	return _SystemConfig.Contract.MinimumGasLimit(&_SystemConfig.CallOpts)
}

// OptimismMintableERC20Factory is a free data retrieval call binding the contract method 0x9b7d7f0a.
//
// Solidity: function optimismMintableERC20Factory() view returns(address addr_)
func (_SystemConfig *SystemConfigCaller) OptimismMintableERC20Factory(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "optimismMintableERC20Factory")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OptimismMintableERC20Factory is a free data retrieval call binding the contract method 0x9b7d7f0a.
//
// Solidity: function optimismMintableERC20Factory() view returns(address addr_)
func (_SystemConfig *SystemConfigSession) OptimismMintableERC20Factory() (common.Address, error) {
	return _SystemConfig.Contract.OptimismMintableERC20Factory(&_SystemConfig.CallOpts)
}

// OptimismMintableERC20Factory is a free data retrieval call binding the contract method 0x9b7d7f0a.
//
// Solidity: function optimismMintableERC20Factory() view returns(address addr_)
func (_SystemConfig *SystemConfigCallerSession) OptimismMintableERC20Factory() (common.Address, error) {
	return _SystemConfig.Contract.OptimismMintableERC20Factory(&_SystemConfig.CallOpts)
}

// OptimismPortal is a free data retrieval call binding the contract method 0x0a49cb03.
//
// Solidity: function optimismPortal() view returns(address addr_)
func (_SystemConfig *SystemConfigCaller) OptimismPortal(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "optimismPortal")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OptimismPortal is a free data retrieval call binding the contract method 0x0a49cb03.
//
// Solidity: function optimismPortal() view returns(address addr_)
func (_SystemConfig *SystemConfigSession) OptimismPortal() (common.Address, error) {
	return _SystemConfig.Contract.OptimismPortal(&_SystemConfig.CallOpts)
}

// OptimismPortal is a free data retrieval call binding the contract method 0x0a49cb03.
//
// Solidity: function optimismPortal() view returns(address addr_)
func (_SystemConfig *SystemConfigCallerSession) OptimismPortal() (common.Address, error) {
	return _SystemConfig.Contract.OptimismPortal(&_SystemConfig.CallOpts)
}

// Overhead is a free data retrieval call binding the contract method 0x0c18c162.
//
// Solidity: function overhead() view returns(uint256)
func (_SystemConfig *SystemConfigCaller) Overhead(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "overhead")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Overhead is a free data retrieval call binding the contract method 0x0c18c162.
//
// Solidity: function overhead() view returns(uint256)
func (_SystemConfig *SystemConfigSession) Overhead() (*big.Int, error) {
	return _SystemConfig.Contract.Overhead(&_SystemConfig.CallOpts)
}

// Overhead is a free data retrieval call binding the contract method 0x0c18c162.
//
// Solidity: function overhead() view returns(uint256)
func (_SystemConfig *SystemConfigCallerSession) Overhead() (*big.Int, error) {
	return _SystemConfig.Contract.Overhead(&_SystemConfig.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SystemConfig *SystemConfigCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SystemConfig *SystemConfigSession) Owner() (common.Address, error) {
	return _SystemConfig.Contract.Owner(&_SystemConfig.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SystemConfig *SystemConfigCallerSession) Owner() (common.Address, error) {
	return _SystemConfig.Contract.Owner(&_SystemConfig.CallOpts)
}

// ResourceConfig is a free data retrieval call binding the contract method 0xcc731b02.
//
// Solidity: function resourceConfig() view returns((uint32,uint8,uint8,uint32,uint32,uint128))
func (_SystemConfig *SystemConfigCaller) ResourceConfig(opts *bind.CallOpts) (ResourceMeteringResourceConfig, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "resourceConfig")

	if err != nil {
		return *new(ResourceMeteringResourceConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(ResourceMeteringResourceConfig)).(*ResourceMeteringResourceConfig)

	return out0, err

}

// ResourceConfig is a free data retrieval call binding the contract method 0xcc731b02.
//
// Solidity: function resourceConfig() view returns((uint32,uint8,uint8,uint32,uint32,uint128))
func (_SystemConfig *SystemConfigSession) ResourceConfig() (ResourceMeteringResourceConfig, error) {
	return _SystemConfig.Contract.ResourceConfig(&_SystemConfig.CallOpts)
}

// ResourceConfig is a free data retrieval call binding the contract method 0xcc731b02.
//
// Solidity: function resourceConfig() view returns((uint32,uint8,uint8,uint32,uint32,uint128))
func (_SystemConfig *SystemConfigCallerSession) ResourceConfig() (ResourceMeteringResourceConfig, error) {
	return _SystemConfig.Contract.ResourceConfig(&_SystemConfig.CallOpts)
}

// Scalar is a free data retrieval call binding the contract method 0xf45e65d8.
//
// Solidity: function scalar() view returns(uint256)
func (_SystemConfig *SystemConfigCaller) Scalar(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "scalar")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Scalar is a free data retrieval call binding the contract method 0xf45e65d8.
//
// Solidity: function scalar() view returns(uint256)
func (_SystemConfig *SystemConfigSession) Scalar() (*big.Int, error) {
	return _SystemConfig.Contract.Scalar(&_SystemConfig.CallOpts)
}

// Scalar is a free data retrieval call binding the contract method 0xf45e65d8.
//
// Solidity: function scalar() view returns(uint256)
func (_SystemConfig *SystemConfigCallerSession) Scalar() (*big.Int, error) {
	return _SystemConfig.Contract.Scalar(&_SystemConfig.CallOpts)
}

// StartBlock is a free data retrieval call binding the contract method 0x48cd4cb1.
//
// Solidity: function startBlock() view returns(uint256)
func (_SystemConfig *SystemConfigCaller) StartBlock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "startBlock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StartBlock is a free data retrieval call binding the contract method 0x48cd4cb1.
//
// Solidity: function startBlock() view returns(uint256)
func (_SystemConfig *SystemConfigSession) StartBlock() (*big.Int, error) {
	return _SystemConfig.Contract.StartBlock(&_SystemConfig.CallOpts)
}

// StartBlock is a free data retrieval call binding the contract method 0x48cd4cb1.
//
// Solidity: function startBlock() view returns(uint256)
func (_SystemConfig *SystemConfigCallerSession) StartBlock() (*big.Int, error) {
	return _SystemConfig.Contract.StartBlock(&_SystemConfig.CallOpts)
}

// UnsafeBlockSigner is a free data retrieval call binding the contract method 0x1fd19ee1.
//
// Solidity: function unsafeBlockSigner() view returns(address addr_)
func (_SystemConfig *SystemConfigCaller) UnsafeBlockSigner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "unsafeBlockSigner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// UnsafeBlockSigner is a free data retrieval call binding the contract method 0x1fd19ee1.
//
// Solidity: function unsafeBlockSigner() view returns(address addr_)
func (_SystemConfig *SystemConfigSession) UnsafeBlockSigner() (common.Address, error) {
	return _SystemConfig.Contract.UnsafeBlockSigner(&_SystemConfig.CallOpts)
}

// UnsafeBlockSigner is a free data retrieval call binding the contract method 0x1fd19ee1.
//
// Solidity: function unsafeBlockSigner() view returns(address addr_)
func (_SystemConfig *SystemConfigCallerSession) UnsafeBlockSigner() (common.Address, error) {
	return _SystemConfig.Contract.UnsafeBlockSigner(&_SystemConfig.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_SystemConfig *SystemConfigCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_SystemConfig *SystemConfigSession) Version() (string, error) {
	return _SystemConfig.Contract.Version(&_SystemConfig.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_SystemConfig *SystemConfigCallerSession) Version() (string, error) {
	return _SystemConfig.Contract.Version(&_SystemConfig.CallOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x36eeb0e6.
//
// Solidity: function initialize(address _owner, uint256 _overhead, uint256 _scalar, bytes32 _batcherHash, uint64 _gasLimit, bool _espresso, address _unsafeBlockSigner, (uint32,uint8,uint8,uint32,uint32,uint128) _config, uint256 _startBlock, address _batchInbox, (address,address,address,address,address,address) _addresses) returns()
func (_SystemConfig *SystemConfigTransactor) Initialize(opts *bind.TransactOpts, _owner common.Address, _overhead *big.Int, _scalar *big.Int, _batcherHash [32]byte, _gasLimit uint64, _espresso bool, _unsafeBlockSigner common.Address, _config ResourceMeteringResourceConfig, _startBlock *big.Int, _batchInbox common.Address, _addresses SystemConfigAddresses) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "initialize", _owner, _overhead, _scalar, _batcherHash, _gasLimit, _espresso, _unsafeBlockSigner, _config, _startBlock, _batchInbox, _addresses)
}

// Initialize is a paid mutator transaction binding the contract method 0x36eeb0e6.
//
// Solidity: function initialize(address _owner, uint256 _overhead, uint256 _scalar, bytes32 _batcherHash, uint64 _gasLimit, bool _espresso, address _unsafeBlockSigner, (uint32,uint8,uint8,uint32,uint32,uint128) _config, uint256 _startBlock, address _batchInbox, (address,address,address,address,address,address) _addresses) returns()
func (_SystemConfig *SystemConfigSession) Initialize(_owner common.Address, _overhead *big.Int, _scalar *big.Int, _batcherHash [32]byte, _gasLimit uint64, _espresso bool, _unsafeBlockSigner common.Address, _config ResourceMeteringResourceConfig, _startBlock *big.Int, _batchInbox common.Address, _addresses SystemConfigAddresses) (*types.Transaction, error) {
	return _SystemConfig.Contract.Initialize(&_SystemConfig.TransactOpts, _owner, _overhead, _scalar, _batcherHash, _gasLimit, _espresso, _unsafeBlockSigner, _config, _startBlock, _batchInbox, _addresses)
}

// Initialize is a paid mutator transaction binding the contract method 0x36eeb0e6.
//
// Solidity: function initialize(address _owner, uint256 _overhead, uint256 _scalar, bytes32 _batcherHash, uint64 _gasLimit, bool _espresso, address _unsafeBlockSigner, (uint32,uint8,uint8,uint32,uint32,uint128) _config, uint256 _startBlock, address _batchInbox, (address,address,address,address,address,address) _addresses) returns()
func (_SystemConfig *SystemConfigTransactorSession) Initialize(_owner common.Address, _overhead *big.Int, _scalar *big.Int, _batcherHash [32]byte, _gasLimit uint64, _espresso bool, _unsafeBlockSigner common.Address, _config ResourceMeteringResourceConfig, _startBlock *big.Int, _batchInbox common.Address, _addresses SystemConfigAddresses) (*types.Transaction, error) {
	return _SystemConfig.Contract.Initialize(&_SystemConfig.TransactOpts, _owner, _overhead, _scalar, _batcherHash, _gasLimit, _espresso, _unsafeBlockSigner, _config, _startBlock, _batchInbox, _addresses)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SystemConfig *SystemConfigTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SystemConfig *SystemConfigSession) RenounceOwnership() (*types.Transaction, error) {
	return _SystemConfig.Contract.RenounceOwnership(&_SystemConfig.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SystemConfig *SystemConfigTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _SystemConfig.Contract.RenounceOwnership(&_SystemConfig.TransactOpts)
}

// SetBatcherHash is a paid mutator transaction binding the contract method 0xc9b26f61.
//
// Solidity: function setBatcherHash(bytes32 _batcherHash) returns()
func (_SystemConfig *SystemConfigTransactor) SetBatcherHash(opts *bind.TransactOpts, _batcherHash [32]byte) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setBatcherHash", _batcherHash)
}

// SetBatcherHash is a paid mutator transaction binding the contract method 0xc9b26f61.
//
// Solidity: function setBatcherHash(bytes32 _batcherHash) returns()
func (_SystemConfig *SystemConfigSession) SetBatcherHash(_batcherHash [32]byte) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetBatcherHash(&_SystemConfig.TransactOpts, _batcherHash)
}

// SetBatcherHash is a paid mutator transaction binding the contract method 0xc9b26f61.
//
// Solidity: function setBatcherHash(bytes32 _batcherHash) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetBatcherHash(_batcherHash [32]byte) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetBatcherHash(&_SystemConfig.TransactOpts, _batcherHash)
}

// SetEspresso is a paid mutator transaction binding the contract method 0xf06d7a3b.
//
// Solidity: function setEspresso(bool _espresso) returns()
func (_SystemConfig *SystemConfigTransactor) SetEspresso(opts *bind.TransactOpts, _espresso bool) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setEspresso", _espresso)
}

// SetEspresso is a paid mutator transaction binding the contract method 0xf06d7a3b.
//
// Solidity: function setEspresso(bool _espresso) returns()
func (_SystemConfig *SystemConfigSession) SetEspresso(_espresso bool) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetEspresso(&_SystemConfig.TransactOpts, _espresso)
}

// SetEspresso is a paid mutator transaction binding the contract method 0xf06d7a3b.
//
// Solidity: function setEspresso(bool _espresso) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetEspresso(_espresso bool) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetEspresso(&_SystemConfig.TransactOpts, _espresso)
}

// SetGasConfig is a paid mutator transaction binding the contract method 0x935f029e.
//
// Solidity: function setGasConfig(uint256 _overhead, uint256 _scalar) returns()
func (_SystemConfig *SystemConfigTransactor) SetGasConfig(opts *bind.TransactOpts, _overhead *big.Int, _scalar *big.Int) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setGasConfig", _overhead, _scalar)
}

// SetGasConfig is a paid mutator transaction binding the contract method 0x935f029e.
//
// Solidity: function setGasConfig(uint256 _overhead, uint256 _scalar) returns()
func (_SystemConfig *SystemConfigSession) SetGasConfig(_overhead *big.Int, _scalar *big.Int) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetGasConfig(&_SystemConfig.TransactOpts, _overhead, _scalar)
}

// SetGasConfig is a paid mutator transaction binding the contract method 0x935f029e.
//
// Solidity: function setGasConfig(uint256 _overhead, uint256 _scalar) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetGasConfig(_overhead *big.Int, _scalar *big.Int) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetGasConfig(&_SystemConfig.TransactOpts, _overhead, _scalar)
}

// SetGasLimit is a paid mutator transaction binding the contract method 0xb40a817c.
//
// Solidity: function setGasLimit(uint64 _gasLimit) returns()
func (_SystemConfig *SystemConfigTransactor) SetGasLimit(opts *bind.TransactOpts, _gasLimit uint64) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setGasLimit", _gasLimit)
}

// SetGasLimit is a paid mutator transaction binding the contract method 0xb40a817c.
//
// Solidity: function setGasLimit(uint64 _gasLimit) returns()
func (_SystemConfig *SystemConfigSession) SetGasLimit(_gasLimit uint64) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetGasLimit(&_SystemConfig.TransactOpts, _gasLimit)
}

// SetGasLimit is a paid mutator transaction binding the contract method 0xb40a817c.
//
// Solidity: function setGasLimit(uint64 _gasLimit) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetGasLimit(_gasLimit uint64) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetGasLimit(&_SystemConfig.TransactOpts, _gasLimit)
}

// SetResourceConfig is a paid mutator transaction binding the contract method 0xc71973f6.
//
// Solidity: function setResourceConfig((uint32,uint8,uint8,uint32,uint32,uint128) _config) returns()
func (_SystemConfig *SystemConfigTransactor) SetResourceConfig(opts *bind.TransactOpts, _config ResourceMeteringResourceConfig) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setResourceConfig", _config)
}

// SetResourceConfig is a paid mutator transaction binding the contract method 0xc71973f6.
//
// Solidity: function setResourceConfig((uint32,uint8,uint8,uint32,uint32,uint128) _config) returns()
func (_SystemConfig *SystemConfigSession) SetResourceConfig(_config ResourceMeteringResourceConfig) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetResourceConfig(&_SystemConfig.TransactOpts, _config)
}

// SetResourceConfig is a paid mutator transaction binding the contract method 0xc71973f6.
//
// Solidity: function setResourceConfig((uint32,uint8,uint8,uint32,uint32,uint128) _config) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetResourceConfig(_config ResourceMeteringResourceConfig) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetResourceConfig(&_SystemConfig.TransactOpts, _config)
}

// SetUnsafeBlockSigner is a paid mutator transaction binding the contract method 0x18d13918.
//
// Solidity: function setUnsafeBlockSigner(address _unsafeBlockSigner) returns()
func (_SystemConfig *SystemConfigTransactor) SetUnsafeBlockSigner(opts *bind.TransactOpts, _unsafeBlockSigner common.Address) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setUnsafeBlockSigner", _unsafeBlockSigner)
}

// SetUnsafeBlockSigner is a paid mutator transaction binding the contract method 0x18d13918.
//
// Solidity: function setUnsafeBlockSigner(address _unsafeBlockSigner) returns()
func (_SystemConfig *SystemConfigSession) SetUnsafeBlockSigner(_unsafeBlockSigner common.Address) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetUnsafeBlockSigner(&_SystemConfig.TransactOpts, _unsafeBlockSigner)
}

// SetUnsafeBlockSigner is a paid mutator transaction binding the contract method 0x18d13918.
//
// Solidity: function setUnsafeBlockSigner(address _unsafeBlockSigner) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetUnsafeBlockSigner(_unsafeBlockSigner common.Address) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetUnsafeBlockSigner(&_SystemConfig.TransactOpts, _unsafeBlockSigner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SystemConfig *SystemConfigTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SystemConfig *SystemConfigSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SystemConfig.Contract.TransferOwnership(&_SystemConfig.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SystemConfig *SystemConfigTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SystemConfig.Contract.TransferOwnership(&_SystemConfig.TransactOpts, newOwner)
}

// SystemConfigConfigUpdateIterator is returned from FilterConfigUpdate and is used to iterate over the raw logs and unpacked data for ConfigUpdate events raised by the SystemConfig contract.
type SystemConfigConfigUpdateIterator struct {
	Event *SystemConfigConfigUpdate // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SystemConfigConfigUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SystemConfigConfigUpdate)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SystemConfigConfigUpdate)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SystemConfigConfigUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SystemConfigConfigUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SystemConfigConfigUpdate represents a ConfigUpdate event raised by the SystemConfig contract.
type SystemConfigConfigUpdate struct {
	Version    *big.Int
	UpdateType uint8
	Data       []byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterConfigUpdate is a free log retrieval operation binding the contract event 0x1d2b0bda21d56b8bd12d4f94ebacffdfb35f5e226f84b461103bb8beab6353be.
//
// Solidity: event ConfigUpdate(uint256 indexed version, uint8 indexed updateType, bytes data)
func (_SystemConfig *SystemConfigFilterer) FilterConfigUpdate(opts *bind.FilterOpts, version []*big.Int, updateType []uint8) (*SystemConfigConfigUpdateIterator, error) {

	var versionRule []interface{}
	for _, versionItem := range version {
		versionRule = append(versionRule, versionItem)
	}
	var updateTypeRule []interface{}
	for _, updateTypeItem := range updateType {
		updateTypeRule = append(updateTypeRule, updateTypeItem)
	}

	logs, sub, err := _SystemConfig.contract.FilterLogs(opts, "ConfigUpdate", versionRule, updateTypeRule)
	if err != nil {
		return nil, err
	}
	return &SystemConfigConfigUpdateIterator{contract: _SystemConfig.contract, event: "ConfigUpdate", logs: logs, sub: sub}, nil
}

// WatchConfigUpdate is a free log subscription operation binding the contract event 0x1d2b0bda21d56b8bd12d4f94ebacffdfb35f5e226f84b461103bb8beab6353be.
//
// Solidity: event ConfigUpdate(uint256 indexed version, uint8 indexed updateType, bytes data)
func (_SystemConfig *SystemConfigFilterer) WatchConfigUpdate(opts *bind.WatchOpts, sink chan<- *SystemConfigConfigUpdate, version []*big.Int, updateType []uint8) (event.Subscription, error) {

	var versionRule []interface{}
	for _, versionItem := range version {
		versionRule = append(versionRule, versionItem)
	}
	var updateTypeRule []interface{}
	for _, updateTypeItem := range updateType {
		updateTypeRule = append(updateTypeRule, updateTypeItem)
	}

	logs, sub, err := _SystemConfig.contract.WatchLogs(opts, "ConfigUpdate", versionRule, updateTypeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SystemConfigConfigUpdate)
				if err := _SystemConfig.contract.UnpackLog(event, "ConfigUpdate", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseConfigUpdate is a log parse operation binding the contract event 0x1d2b0bda21d56b8bd12d4f94ebacffdfb35f5e226f84b461103bb8beab6353be.
//
// Solidity: event ConfigUpdate(uint256 indexed version, uint8 indexed updateType, bytes data)
func (_SystemConfig *SystemConfigFilterer) ParseConfigUpdate(log types.Log) (*SystemConfigConfigUpdate, error) {
	event := new(SystemConfigConfigUpdate)
	if err := _SystemConfig.contract.UnpackLog(event, "ConfigUpdate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SystemConfigInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the SystemConfig contract.
type SystemConfigInitializedIterator struct {
	Event *SystemConfigInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SystemConfigInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SystemConfigInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SystemConfigInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SystemConfigInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SystemConfigInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SystemConfigInitialized represents a Initialized event raised by the SystemConfig contract.
type SystemConfigInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_SystemConfig *SystemConfigFilterer) FilterInitialized(opts *bind.FilterOpts) (*SystemConfigInitializedIterator, error) {

	logs, sub, err := _SystemConfig.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &SystemConfigInitializedIterator{contract: _SystemConfig.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_SystemConfig *SystemConfigFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *SystemConfigInitialized) (event.Subscription, error) {

	logs, sub, err := _SystemConfig.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SystemConfigInitialized)
				if err := _SystemConfig.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_SystemConfig *SystemConfigFilterer) ParseInitialized(log types.Log) (*SystemConfigInitialized, error) {
	event := new(SystemConfigInitialized)
	if err := _SystemConfig.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SystemConfigOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SystemConfig contract.
type SystemConfigOwnershipTransferredIterator struct {
	Event *SystemConfigOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SystemConfigOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SystemConfigOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SystemConfigOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SystemConfigOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SystemConfigOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SystemConfigOwnershipTransferred represents a OwnershipTransferred event raised by the SystemConfig contract.
type SystemConfigOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SystemConfig *SystemConfigFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SystemConfigOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SystemConfig.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SystemConfigOwnershipTransferredIterator{contract: _SystemConfig.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SystemConfig *SystemConfigFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SystemConfigOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SystemConfig.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SystemConfigOwnershipTransferred)
				if err := _SystemConfig.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SystemConfig *SystemConfigFilterer) ParseOwnershipTransferred(log types.Log) (*SystemConfigOwnershipTransferred, error) {
	event := new(SystemConfigOwnershipTransferred)
	if err := _SystemConfig.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

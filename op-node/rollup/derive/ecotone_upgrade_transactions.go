package derive

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-service/solabi"
)

const UpgradeToFuncSignature = "upgradeTo(address)"

var (
	// known address w/ zero txns
	L1BlockDeployerAddress        = common.HexToAddress("0x4210000000000000000000000000000000000000")
	GasPriceOracleDeployerAddress = common.HexToAddress("0x4210000000000000000000000000000000000001")

	newL1BlockAddress        = crypto.CreateAddress(L1BlockDeployerAddress, 0)
	newGasPriceOracleAddress = crypto.CreateAddress(GasPriceOracleDeployerAddress, 0)

	deployL1BlockSource        = UpgradeDepositSource{Intent: "Ecotone: L1 Block Deployment"}
	deployGasPriceOracleSource = UpgradeDepositSource{Intent: "Ecotone: Gas Price Oracle Deployment"}
	updateL1BlockProxySource   = UpgradeDepositSource{Intent: "Ecotone: L1 Block Proxy Update"}
	updateGasPriceOracleSource = UpgradeDepositSource{Intent: "Ecotone: Gas Price Oracle Proxy Update"}
	enableEcotoneSource        = UpgradeDepositSource{Intent: "Ecotone: Gas Price Oracle Set Ecotone"}
	beaconRootsSource          = UpgradeDepositSource{Intent: "Ecotone: beacon block roots contract deployment"}

	enableEcotoneInput = crypto.Keccak256([]byte("setEcotone()"))[:4]

	EIP4788From         = common.HexToAddress("0x0B799C86a49DEeb90402691F1041aa3AF2d3C875")
	eip4788CreationData = common.Hex2Bytes("0x60618060095f395ff33373fffffffffffffffffffffffffffffffffffffffe14604d57602036146024575f5ffd5b5f35801560495762001fff810690815414603c575f5ffd5b62001fff01545f5260205ff35b5f5ffd5b62001fff42064281555f359062001fff015500")
	UpgradeToFuncBytes4 = crypto.Keccak256([]byte(UpgradeToFuncSignature))[:4]

	l1BlockDeploymentBytecode        = common.FromHex("0x608060405234801561001057600080fd5b506106a4806100206000396000f3fe608060405234801561001057600080fd5b506004361061011b5760003560e01c80638381f58a116100b2578063c598591811610081578063e591b28211610066578063e591b282146102a3578063e81b2c6d146102e3578063f8206140146102ec57600080fd5b8063c598591814610263578063dc59462c1461028357600080fd5b80638381f58a1461021d5780638b239f73146102315780639e8c49661461023a578063b80777ea1461024357600080fd5b80635cf24969116100ee5780635cf24969146101a257806361fba0ca146101ab57806364ca23ef146101d857806368d5dca6146101ec57600080fd5b806309bd5a6014610120578063440a5e201461013c57806354b7325c1461014657806354fd4d5014610159575b600080fd5b61012960025481565b6040519081526020015b60405180910390f35b6101446102f5565b005b610144610154366004610595565b610353565b6101956040518060400160405280600581526020017f312e322e3000000000000000000000000000000000000000000000000000000081525081565b60405161013391906105d8565b61012960015481565b6008546101bf9067ffffffffffffffff1681565b60405167ffffffffffffffff9091168152602001610133565b6003546101bf9067ffffffffffffffff1681565b6003546102089068010000000000000000900463ffffffff1681565b60405163ffffffff9091168152602001610133565b6000546101bf9067ffffffffffffffff1681565b61012960055481565b61012960065481565b6000546101bf9068010000000000000000900467ffffffffffffffff1681565b600354610208906c01000000000000000000000000900463ffffffff1681565b6008546101bf9068010000000000000000900467ffffffffffffffff1681565b6102be73deaddeaddeaddeaddeaddeaddeaddeaddead000181565b60405173ffffffffffffffffffffffffffffffffffffffff9091168152602001610133565b61012960045481565b61012960075481565b3373deaddeaddeaddeaddeaddeaddeaddeaddead00011461031e57633cc50b456000526004601cfd5b60043560801c60035560143560801c60005560243560015560443560075560643560025560843560045560a43560801c600855565b3373deaddeaddeaddeaddeaddeaddeaddeaddead0001146103fa576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603b60248201527f4c31426c6f636b3a206f6e6c7920746865206465706f7369746f72206163636f60448201527f756e742063616e20736574204c3120626c6f636b2076616c7565730000000000606482015260840160405180910390fd5b610407602082018261064b565b600080547fffffffffffffffffffffffffffffffffffffffffffffffff00000000000000001667ffffffffffffffff92909216919091179055610450604082016020830161064b565b6000805467ffffffffffffffff9290921668010000000000000000027fffffffffffffffffffffffffffffffff0000000000000000ffffffffffffffff909216919091179055604081013560015560608101356002556104b660a082016080830161064b565b600380547fffffffffffffffffffffffffffffffffffffffffffffffff00000000000000001667ffffffffffffffff9290921691909117905560a081013560045560c081013560055560e081013560065561051961012082016101008301610675565b610524576000610527565b60015b600880547fffffffffffffffffffffffffffffffffffffffffffffffff00000000000000001660ff9290921691909117905561056b6101408201610120830161064b565b6008806101000a81548167ffffffffffffffff021916908367ffffffffffffffff16021790555050565b6000602082840312156105a757600080fd5b813567ffffffffffffffff8111156105be57600080fd5b820161016081850312156105d157600080fd5b9392505050565b600060208083528351808285015260005b81811015610605578581018301518582016040015282016105e9565b81811115610617576000604083870101525b50601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016929092016040019392505050565b60006020828403121561065d57600080fd5b813567ffffffffffffffff811681146105d157600080fd5b60006020828403121561068757600080fd5b813580151581146105d157600080fdfea164736f6c634300080f000a")
	gasPriceOracleDeploymentBytecode = common.FromHex("0x608060405234801561001057600080fd5b50610fb5806100206000396000f3fe608060405234801561001057600080fd5b50600436106100f55760003560e01c806354fd4d5011610097578063de26c4a111610066578063de26c4a1146101da578063f45e65d8146101ed578063f8206140146101f5578063fe173b97146101cc57600080fd5b806354fd4d501461016657806368d5dca6146101af5780636ef25c3a146101cc578063c5985918146101d257600080fd5b8063313ce567116100d3578063313ce5671461012757806349948e0e1461012e5780634ef6e22414610141578063519b4bd31461015e57600080fd5b80630c18c162146100fa57806322b90ab3146101155780632e0f26251461011f575b600080fd5b6101026101fd565b6040519081526020015b60405180910390f35b61011d61031e565b005b610102600681565b6006610102565b61010261013c366004610b73565b610541565b60005461014e9060ff1681565b604051901515815260200161010c565b610102610565565b6101a26040518060400160405280600581526020017f312e322e3000000000000000000000000000000000000000000000000000000081525081565b60405161010c9190610c42565b6101b76105c6565b60405163ffffffff909116815260200161010c565b48610102565b6101b761064b565b6101026101e8366004610b73565b6106ac565b610102610760565b610102610853565b6000805460ff1615610296576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602860248201527f47617350726963654f7261636c653a206f76657268656164282920697320646560448201527f707265636174656400000000000000000000000000000000000000000000000060648201526084015b60405180910390fd5b73420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff16638b239f736040518163ffffffff1660e01b8152600401602060405180830381865afa1580156102f5573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906103199190610cb5565b905090565b73420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff1663e591b2826040518163ffffffff1660e01b8152600401602060405180830381865afa15801561037d573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906103a19190610cce565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610481576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604160248201527f47617350726963654f7261636c653a206f6e6c7920746865206465706f73697460448201527f6f72206163636f756e742063616e2073657420697345636f746f6e6520666c6160648201527f6700000000000000000000000000000000000000000000000000000000000000608482015260a40161028d565b60005460ff1615610514576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f47617350726963654f7261636c653a2045636f746f6e6520616c72656164792060448201527f6163746976650000000000000000000000000000000000000000000000000000606482015260840161028d565b600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00166001179055565b6000805460ff161561055c57610556826108b4565b92915050565b61055682610958565b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff16635cf249696040518163ffffffff1660e01b8152600401602060405180830381865afa1580156102f5573d6000803e3d6000fd5b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff166368d5dca66040518163ffffffff1660e01b8152600401602060405180830381865afa158015610627573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906103199190610d04565b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff1663c59859186040518163ffffffff1660e01b8152600401602060405180830381865afa158015610627573d6000803e3d6000fd5b6000806106b883610ab4565b60005490915060ff16156106cc5792915050565b73420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff16638b239f736040518163ffffffff1660e01b8152600401602060405180830381865afa15801561072b573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061074f9190610cb5565b6107599082610d59565b9392505050565b6000805460ff16156107f4576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f47617350726963654f7261636c653a207363616c61722829206973206465707260448201527f6563617465640000000000000000000000000000000000000000000000000000606482015260840161028d565b73420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff16639e8c49666040518163ffffffff1660e01b8152600401602060405180830381865afa1580156102f5573d6000803e3d6000fd5b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff1663f82061406040518163ffffffff1660e01b8152600401602060405180830381865afa1580156102f5573d6000803e3d6000fd5b6000806108c083610ab4565b905060006108cc610565565b6108d461064b565b6108df906010610d71565b63ffffffff166108ef9190610d9d565b905060006108fb610853565b6109036105c6565b63ffffffff166109139190610d9d565b905060006109218284610d59565b61092b9085610d9d565b90506109396006600a610efa565b610944906010610d9d565b61094e9082610f06565b9695505050505050565b60008061096483610ab4565b9050600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff16639e8c49666040518163ffffffff1660e01b8152600401602060405180830381865afa1580156109c7573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906109eb9190610cb5565b6109f3610565565b73420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff16638b239f736040518163ffffffff1660e01b8152600401602060405180830381865afa158015610a52573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610a769190610cb5565b610a809085610d59565b610a8a9190610d9d565b610a949190610d9d565b9050610aa26006600a610efa565b610aac9082610f06565b949350505050565b80516000908190815b81811015610b3757848181518110610ad757610ad7610f41565b01602001517fff0000000000000000000000000000000000000000000000000000000000000016600003610b1757610b10600484610d59565b9250610b25565b610b22601084610d59565b92505b80610b2f81610f70565b915050610abd565b50610aac82610440610d59565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b600060208284031215610b8557600080fd5b813567ffffffffffffffff80821115610b9d57600080fd5b818401915084601f830112610bb157600080fd5b813581811115610bc357610bc3610b44565b604051601f82017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0908116603f01168101908382118183101715610c0957610c09610b44565b81604052828152876020848701011115610c2257600080fd5b826020860160208301376000928101602001929092525095945050505050565b600060208083528351808285015260005b81811015610c6f57858101830151858201604001528201610c53565b81811115610c81576000604083870101525b50601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016929092016040019392505050565b600060208284031215610cc757600080fd5b5051919050565b600060208284031215610ce057600080fd5b815173ffffffffffffffffffffffffffffffffffffffff8116811461075957600080fd5b600060208284031215610d1657600080fd5b815163ffffffff8116811461075957600080fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60008219821115610d6c57610d6c610d2a565b500190565b600063ffffffff80831681851681830481118215151615610d9457610d94610d2a565b02949350505050565b6000817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0483118215151615610dd557610dd5610d2a565b500290565b600181815b80851115610e3357817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff04821115610e1957610e19610d2a565b80851615610e2657918102915b93841c9390800290610ddf565b509250929050565b600082610e4a57506001610556565b81610e5757506000610556565b8160018114610e6d5760028114610e7757610e93565b6001915050610556565b60ff841115610e8857610e88610d2a565b50506001821b610556565b5060208310610133831016604e8410600b8410161715610eb6575081810a610556565b610ec08383610dda565b807fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff04821115610ef257610ef2610d2a565b029392505050565b60006107598383610e3b565b600082610f3c577f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b500490565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8203610fa157610fa1610d2a565b506001019056fea164736f6c634300080f000a")
)

func EcotoneNetworkUpgradeTransactions() ([]hexutil.Bytes, error) {
	upgradeTxns := make([]hexutil.Bytes, 0, 5)

	deployL1BlockTransaction, err := types.NewTx(&types.DepositTx{
		SourceHash:          deployL1BlockSource.SourceHash(),
		From:                L1BlockDeployerAddress,
		To:                  nil,
		Mint:                big.NewInt(0),
		Value:               big.NewInt(0),
		Gas:                 375_000,
		IsSystemTransaction: false,
		Data:                l1BlockDeploymentBytecode,
	}).MarshalBinary()

	if err != nil {
		return nil, err
	}

	upgradeTxns = append(upgradeTxns, deployL1BlockTransaction)

	deployGasPriceOracle, err := types.NewTx(&types.DepositTx{
		SourceHash:          deployGasPriceOracleSource.SourceHash(),
		From:                GasPriceOracleDeployerAddress,
		To:                  nil,
		Mint:                big.NewInt(0),
		Value:               big.NewInt(0),
		Gas:                 1_000_000,
		IsSystemTransaction: false,
		Data:                gasPriceOracleDeploymentBytecode,
	}).MarshalBinary()

	if err != nil {
		return nil, err
	}

	upgradeTxns = append(upgradeTxns, deployGasPriceOracle)

	updateL1BlockProxy, err := types.NewTx(&types.DepositTx{
		SourceHash:          updateL1BlockProxySource.SourceHash(),
		From:                common.Address{},
		To:                  &predeploys.L1BlockAddr,
		Mint:                big.NewInt(0),
		Value:               big.NewInt(0),
		Gas:                 50_000,
		IsSystemTransaction: false,
		Data:                upgradeToCalldata(newL1BlockAddress),
	}).MarshalBinary()

	if err != nil {
		return nil, err
	}

	upgradeTxns = append(upgradeTxns, updateL1BlockProxy)

	updateGasPriceOracleProxy, err := types.NewTx(&types.DepositTx{
		SourceHash:          updateGasPriceOracleSource.SourceHash(),
		From:                common.Address{},
		To:                  &predeploys.GasPriceOracleAddr,
		Mint:                big.NewInt(0),
		Value:               big.NewInt(0),
		Gas:                 50_000,
		IsSystemTransaction: false,
		Data:                upgradeToCalldata(newGasPriceOracleAddress),
	}).MarshalBinary()

	if err != nil {
		return nil, err
	}

	upgradeTxns = append(upgradeTxns, updateGasPriceOracleProxy)

	enableEcotone, err := types.NewTx(&types.DepositTx{
		SourceHash:          enableEcotoneSource.SourceHash(),
		From:                L1InfoDepositerAddress,
		To:                  &predeploys.GasPriceOracleAddr,
		Mint:                big.NewInt(0),
		Value:               big.NewInt(0),
		Gas:                 80_000,
		IsSystemTransaction: false,
		Data:                enableEcotoneInput,
	}).MarshalBinary()
	if err != nil {
		return nil, err
	}
	upgradeTxns = append(upgradeTxns, enableEcotone)

	deployEIP4788, err := types.NewTx(&types.DepositTx{
		From:                EIP4788From,
		To:                  nil, // contract-deployment tx
		Mint:                big.NewInt(0),
		Value:               big.NewInt(0),
		Gas:                 0x3d090, // hex constant, as defined in EIP-4788
		Data:                eip4788CreationData,
		IsSystemTransaction: false,
		SourceHash:          beaconRootsSource.SourceHash(),
	}).MarshalBinary()

	if err != nil {
		return nil, err
	}

	upgradeTxns = append(upgradeTxns, deployEIP4788)

	return upgradeTxns, nil
}

func upgradeToCalldata(addr common.Address) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 4+20))
	if err := solabi.WriteSignature(buf, UpgradeToFuncBytes4); err != nil {
		panic(fmt.Errorf("failed to write upgradeTo signature data: %w", err))
	}
	if err := solabi.WriteAddress(buf, addr); err != nil {
		panic(fmt.Errorf("failed to write upgradeTo address data: %w", err))
	}
	return buf.Bytes()
}

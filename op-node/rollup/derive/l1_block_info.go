package derive

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-service/solabi"
)

const (
	L1InfoFuncSignature = "setL1BlockValues(uint64,uint64,uint256,bytes32,uint64,bytes32,uint256,uint256,bytes)"
	L1InfoArguments     = 8
)

var (
	L1InfoFuncBytes4          = crypto.Keccak256([]byte(L1InfoFuncSignature))[:4]
	L1InfoDepositerAddress    = common.HexToAddress("0xdeaddeaddeaddeaddeaddeaddeaddeaddead0001")
	L1InfoJustificationOffset = new(big.Int).SetUint64(288) // See Binary Format table below
	L1BlockAddress            = predeploys.L1BlockAddr
)

const (
	RegolithSystemTxGas = 1_000_000
)

// L1BlockInfo presents the information stored in a L1Block.setL1BlockValues call
type L1BlockInfo struct {
	Number    uint64
	Time      uint64
	BaseFee   *big.Int
	BlockHash common.Hash
	// Not strictly a piece of L1 information. Represents the number of L2 blocks since the start of the epoch,
	// i.e. when the actual L1 info was first introduced.
	SequenceNumber uint64
	// BatcherHash version 0 is just the address with 0 padding to the left.
	BatcherAddr   common.Address
	L1FeeOverhead eth.Bytes32
	L1FeeScalar   eth.Bytes32
	Justification *eth.L2BatchJustification `rlp:"nil"`
}

// Binary Format
// +---------+--------------------------+
// | Bytes   | Field                    |
// +---------+--------------------------+
// | 4       | Function signature       |
// | 32      | Number                   |
// | 32      | Time                     |
// | 32      | BaseFee                  |
// | 32      | BlockHash                |
// | 32      | SequenceNumber           |
// | 32      | BatcherAddr              |
// | 32      | L1FeeOverhead            |
// | 32      | L1FeeScalar              |
// | 32      | L1InfoJustificationOffset|
// | 		 | (this is how dynamic     |
// | 		 | types are ABI encoded)   |
// | variable| Justification            |
// +---------+--------------------------+

func (info *L1BlockInfo) MarshalBinary() ([]byte, error) {
	w := new(bytes.Buffer)
	if err := solabi.WriteSignature(w, L1InfoFuncBytes4); err != nil {
		return nil, err
	}
	if err := solabi.WriteUint64(w, info.Number); err != nil {
		return nil, err
	}
	if err := solabi.WriteUint64(w, info.Time); err != nil {
		return nil, err
	}
	if err := solabi.WriteUint256(w, info.BaseFee); err != nil {
		return nil, err
	}
	if err := solabi.WriteHash(w, info.BlockHash); err != nil {
		return nil, err
	}
	if err := solabi.WriteUint64(w, info.SequenceNumber); err != nil {
		return nil, err
	}
	if err := solabi.WriteAddress(w, info.BatcherAddr); err != nil {
		return nil, err
	}
	if err := solabi.WriteEthBytes32(w, info.L1FeeOverhead); err != nil {
		return nil, err
	}
	if err := solabi.WriteEthBytes32(w, info.L1FeeScalar); err != nil {
		return nil, err
	}

	// For simplicity, we don't ABI-encode the whole structure of the Justification. We RLP-encode
	// it and then ABI-encode the resulting byte string. This means the Justification can be
	// accessed by parsing calldata, but cannot (easily) by inspected on-chain.
	rlpBytes, err := rlp.EncodeToBytes(info.Justification)
	if err != nil {
		return nil, err
	}
	// The ABI-encoding of function parameters is that of a tuple, which requires that dynamic types
	// (such as `bytes`) are represented in the initial list of items as a uint256 with the offset
	// from the start of the encoding to the start of the payload of the dynamic type, which follows
	// the initial list of static types and dynamic type offsets. In this case, we only have one
	// item of dynamic type, and it is at the end of the list of items, so we will encode it by its
	// offset, which is just the length of the static section of the list, followed by the item
	// itself.
	if err := solabi.WriteUint256(w, L1InfoJustificationOffset); err != nil {
		return nil, err
	}
	if err := solabi.WriteBytes(w, rlpBytes); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (info *L1BlockInfo) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	var err error
	if _, err := solabi.ReadAndValidateSignature(reader, L1InfoFuncBytes4); err != nil {
		return err
	}
	if info.Number, err = solabi.ReadUint64(reader); err != nil {
		return err
	}
	if info.Time, err = solabi.ReadUint64(reader); err != nil {
		return err
	}
	if info.BaseFee, err = solabi.ReadUint256(reader); err != nil {
		return err
	}
	if info.BlockHash, err = solabi.ReadHash(reader); err != nil {
		return err
	}
	if info.SequenceNumber, err = solabi.ReadUint64(reader); err != nil {
		return err
	}
	if info.BatcherAddr, err = solabi.ReadAddress(reader); err != nil {
		return err
	}
	if info.L1FeeOverhead, err = solabi.ReadEthBytes32(reader); err != nil {
		return err
	}
	if info.L1FeeScalar, err = solabi.ReadEthBytes32(reader); err != nil {
		return err
	}

	// Read the offset of the Justification bytes followed by the bytes themselves.
	rlpOffset, err := solabi.ReadUint256(reader)
	if err != nil {
		return err
	}
	if rlpOffset.Cmp(L1InfoJustificationOffset) != 0 {
		return fmt.Errorf("invalid justification offset (%d, expected %d)", rlpOffset, L1InfoJustificationOffset)
	}
	rlpBytes, err := solabi.ReadBytes(reader)
	if err != nil {
		return err
	}
	// If the remaining bytes are the RLP encoding of an empty list (0xc, which represents a `nil`
	// pointer) skip the Justification. The RLP library automatically handles `nil` pointers as
	// struct fields with the `rlp:"nil"` attribute, but here it is not a nested field which might
	// be `nil` but the top-level object, and the RLP library does not allow that.
	if !(len(rlpBytes) == 1 && rlpBytes[0] == 0xc0) {
		if err := rlp.DecodeBytes(rlpBytes, &info.Justification); err != nil {
			return err
		}
	}

	if !solabi.EmptyReader(reader) {
		return errors.New("too many bytes")
	}
	return nil
}

// L1InfoDepositTxData is the inverse of L1InfoDeposit, to see where the L2 chain is derived from
func L1InfoDepositTxData(data []byte) (L1BlockInfo, error) {
	var info L1BlockInfo
	err := info.UnmarshalBinary(data)
	return info, err
}

// L1InfoDeposit creates a L1 Info deposit transaction based on the L1 block,
// and the L2 block-height difference with the start of the epoch.
func L1InfoDeposit(seqNumber uint64, block eth.BlockInfo, sysCfg eth.SystemConfig, justification *eth.L2BatchJustification, regolith bool) (*types.DepositTx, error) {
	infoDat := L1BlockInfo{
		Number:         block.NumberU64(),
		Time:           block.Time(),
		BaseFee:        block.BaseFee(),
		BlockHash:      block.Hash(),
		SequenceNumber: seqNumber,
		BatcherAddr:    sysCfg.BatcherAddr,
		L1FeeOverhead:  sysCfg.Overhead,
		L1FeeScalar:    sysCfg.Scalar,
		Justification:  justification,
	}
	data, err := infoDat.MarshalBinary()
	if err != nil {
		return nil, err
	}

	source := L1InfoDepositSource{
		L1BlockHash: block.Hash(),
		SeqNumber:   seqNumber,
	}
	// Set a very large gas limit with `IsSystemTransaction` to ensure
	// that the L1 Attributes Transaction does not run out of gas.
	out := &types.DepositTx{
		SourceHash:          source.SourceHash(),
		From:                L1InfoDepositerAddress,
		To:                  &L1BlockAddress,
		Mint:                nil,
		Value:               big.NewInt(0),
		Gas:                 150_000_000,
		IsSystemTransaction: true,
		Data:                data,
	}
	// With the regolith fork we disable the IsSystemTx functionality, and allocate real gas
	if regolith {
		out.IsSystemTransaction = false
		out.Gas = RegolithSystemTxGas
	}
	return out, nil
}

// L1InfoDepositBytes returns a serialized L1-info attributes transaction.
func L1InfoDepositBytes(seqNumber uint64, l1Info eth.BlockInfo, sysCfg eth.SystemConfig, justification *eth.L2BatchJustification, regolith bool) ([]byte, error) {
	dep, err := L1InfoDeposit(seqNumber, l1Info, sysCfg, justification, regolith)
	if err != nil {
		return nil, fmt.Errorf("failed to create L1 info tx: %w", err)
	}
	l1Tx := types.NewTx(dep)
	opaqueL1Tx, err := l1Tx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to encode L1 info tx: %w", err)
	}
	return opaqueL1Tx, nil
}

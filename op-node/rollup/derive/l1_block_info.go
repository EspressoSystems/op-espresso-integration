package derive

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/solabi"
)

const (
	L1InfoFuncBedrockSignature = "setL1BlockValues((uint64,uint64,uint256,bytes32,uint64,bytes32,uint256,uint256,bool,uint64,bytes))"
	L1InfoFuncEcotoneSignature = "setL1BlockValuesEcotone()"
	L1InfoArguments            = 8
)

var (
	L1InfoFuncBedrockBytes4   = crypto.Keccak256([]byte(L1InfoFuncBedrockSignature))[:4]
	L1InfoFuncEcotoneBytes4   = crypto.Keccak256([]byte(L1InfoFuncEcotoneSignature))[:4]
	L1InfoDepositerAddress    = common.HexToAddress("0xdeaddeaddeaddeaddeaddeaddeaddeaddead0001")
	L1InfoJustificationOffset = new(big.Int).SetUint64(352) // See Binary Format table below
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
	BatcherAddr common.Address

	// Whether Espresso mode is enabled.
	Espresso bool
	// When using Espresso, the configured confirmation depth for L1 origins.
	EspressoL1ConfDepth uint64

	Justification *eth.L2BatchJustification `rlp:"nil"`

	L1FeeOverhead eth.Bytes32 // ignored after Ecotone upgrade
	L1FeeScalar   eth.Bytes32 // ignored after Ecotone upgrade

	BlobBaseFee       *big.Int // added by Ecotone upgrade
	BaseFeeScalar     uint32   // added by Ecotone upgrade
	BlobBaseFeeScalar uint32   // added by Ecotone upgrade
}

// Bedrock Binary Format
//
// We marshal `L1BlockInfo` using the ABI encoding for a call to the `setL1BlockValues` method. This
// method has one argument, a struct of type `L1BlockValues`. This struct in turn contains all the
// fields of `L1BlockInfo` (we do this to avoid exceeding the stack depth, since we must pass many
// fields).
//
// As always, the parameters to a method are encoded as a tuple, in this case a 1-tuple containing
// only the struct. The struct itself is also encoded as a tuple, which is a dynamically sized type.
// Thus, the encoding of the struct consists of the offset of the segment of the encoding containing
// dynamic values, followed by each field of the struct in order. In this case, since there is only
// one argument in the outer tuple, the offset of the dynamic section is 32, since the offset itself
// takes up 32 bytes and the dynamic section follows immediately after.
//
// +---------+--------------------------+
// | Bytes   | Field                    |
// +---------+--------------------------+
// | 4       | Function signature       |
// | 32      | Struct fields offset (32)|
// | 32      | Number                   |
// | 32      | Time                     |
// | 32      | BaseFee                  |
// | 32      | BlockHash                |
// | 32      | SequenceNumber           |
// | 32      | BatcherHash              |
// | 32      | L1FeeOverhead            |
// | 32      | L1FeeScalar              |
// | 32      | Espresso                 |
// | 32      | EspressoL1ConfDepth      |
// | 32      | L1InfoJustificationOffset|
// | variable| Justification            |
// +---------+--------------------------+

func (info *L1BlockInfo) marshalBinaryBedrock() ([]byte, error) {
	w := new(bytes.Buffer)
	if err := solabi.WriteSignature(w, L1InfoFuncBedrockBytes4); err != nil {
		return nil, err
	}
	if err := solabi.WriteUint64(w, 32); err != nil {
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
	if err := solabi.WriteBool(w, info.Espresso); err != nil {
		return nil, err
	}
	if err := solabi.WriteUint64(w, info.EspressoL1ConfDepth); err != nil {
		return nil, err
	}

	// For simplicity, we don't ABI-encode the whole structure of the Justification. We RLP-encode
	// it and then ABI-encode the resulting byte string. This means the Justification can be
	// accessed by parsing calldata, but cannot (easily) by inspected on-chain.
	rlpBytes, err := rlp.EncodeToBytes(info.Justification)
	if err != nil {
		return nil, err
	}
	// The ABI-encoding of struct fields is that of a tuple, which requires that dynamic types (such
	// as `bytes`) are represented in the initial list of items as a uint256 with the offset from
	// the start of the encoding to the start of the payload of the dynamic type, which follows the
	// initial list of static types and dynamic type offsets. In this case, we only have one item of
	// dynamic type, and it is at the end of the list of items, so we will encode it by its offset,
	// which is just the length of the static section of the list, followed by the item itself.
	if err := solabi.WriteUint256(w, L1InfoJustificationOffset); err != nil {
		return nil, err
	}
	if err := solabi.WriteBytes(w, rlpBytes); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (info *L1BlockInfo) unmarshalBinaryBedrock(data []byte) error {
	reader := bytes.NewReader(data)

	var err error
	if _, err := solabi.ReadAndValidateSignature(reader, L1InfoFuncBedrockBytes4); err != nil {
		return err
	}
	if fieldsOffset, err := solabi.ReadUint64(reader); err != nil {
		return err
	} else if fieldsOffset != 32 {
		return fmt.Errorf("invalid struct fields offset (%d, expected 32)", fieldsOffset)
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
	if info.Espresso, err = solabi.ReadBool(reader); err != nil {
		return err
	}
	if info.EspressoL1ConfDepth, err = solabi.ReadUint64(reader); err != nil {
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

// Ecotone Binary Format
// +---------+--------------------------+
// | Bytes   | Field                    |
// +---------+--------------------------+
// | 4       | Function signature       |
// | 4       | BaseFeeScalar            |
// | 4       | BlobBaseFeeScalar        |
// | 8       | SequenceNumber           |
// | 8       | Timestamp                |
// | 8       | L1BlockNumber            |
// | 32      | BaseFee                  |
// | 32      | BlobBaseFee              |
// | 32      | BlockHash                |
// | 32      | BatcherHash              |
// | 8       | EspressoL1ConfDepth      |
// | 8       | Espresso                 |
// | variable| Justification            |
// +---------+--------------------------+

func (info *L1BlockInfo) marshalBinaryEcotone() ([]byte, error) {
	w := new(bytes.Buffer)
	if err := solabi.WriteSignature(w, L1InfoFuncEcotoneBytes4); err != nil {
		return nil, err
	}
	if err := binary.Write(w, binary.BigEndian, info.BaseFeeScalar); err != nil {
		return nil, err
	}
	if err := binary.Write(w, binary.BigEndian, info.BlobBaseFeeScalar); err != nil {
		return nil, err
	}
	if err := binary.Write(w, binary.BigEndian, info.SequenceNumber); err != nil {
		return nil, err
	}
	if err := binary.Write(w, binary.BigEndian, info.Time); err != nil {
		return nil, err
	}
	if err := binary.Write(w, binary.BigEndian, info.Number); err != nil {
		return nil, err
	}
	if err := solabi.WriteUint256(w, info.BaseFee); err != nil {
		return nil, err
	}
	blobBasefee := info.BlobBaseFee
	if blobBasefee == nil {
		blobBasefee = big.NewInt(1) // set to 1, to match the min blob basefee as defined in EIP-4844
	}
	if err := solabi.WriteUint256(w, blobBasefee); err != nil {
		return nil, err
	}
	if err := solabi.WriteHash(w, info.BlockHash); err != nil {
		return nil, err
	}
	// ABI encoding will perform the left-padding with zeroes to 32 bytes, matching the "batcherHash" SystemConfig format and version 0 byte.
	if err := solabi.WriteAddress(w, info.BatcherAddr); err != nil {
		return nil, err
	}
	if err := binary.Write(w, binary.BigEndian, info.EspressoL1ConfDepth); err != nil {
		return nil, err
	}
	if info.Espresso {
		if err := binary.Write(w, binary.BigEndian, uint64(1)); err != nil {
			return nil, err
		}
	} else {
		if err := binary.Write(w, binary.BigEndian, uint64(0)); err != nil {
			return nil, err
		}
	}
	rlpBytes, err := rlp.EncodeToBytes(info.Justification)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(rlpBytes); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (info *L1BlockInfo) unmarshalBinaryEcotone(data []byte) error {
	r := bytes.NewReader(data)

	var err error
	if _, err := solabi.ReadAndValidateSignature(r, L1InfoFuncEcotoneBytes4); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &info.BaseFeeScalar); err != nil {
		return fmt.Errorf("invalid ecotone l1 block info format")
	}
	if err := binary.Read(r, binary.BigEndian, &info.BlobBaseFeeScalar); err != nil {
		return fmt.Errorf("invalid ecotone l1 block info format")
	}
	if err := binary.Read(r, binary.BigEndian, &info.SequenceNumber); err != nil {
		return fmt.Errorf("invalid ecotone l1 block info format")
	}
	if err := binary.Read(r, binary.BigEndian, &info.Time); err != nil {
		return fmt.Errorf("invalid ecotone l1 block info format")
	}
	if err := binary.Read(r, binary.BigEndian, &info.Number); err != nil {
		return fmt.Errorf("invalid ecotone l1 block info format")
	}
	if info.BaseFee, err = solabi.ReadUint256(r); err != nil {
		return err
	}
	if info.BlobBaseFee, err = solabi.ReadUint256(r); err != nil {
		return err
	}
	if info.BlockHash, err = solabi.ReadHash(r); err != nil {
		return err
	}
	// The "batcherHash" will be correctly parsed as address, since the version 0 and left-padding matches the ABI encoding format.
	if info.BatcherAddr, err = solabi.ReadAddress(r); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &info.EspressoL1ConfDepth); err != nil {
		return fmt.Errorf("invalid ecotone l1 block info format: %w", err)
	}
	var espresso uint64
	if err := binary.Read(r, binary.BigEndian, &espresso); err != nil {
		return fmt.Errorf("invalid ecotone l1 block info format: %w", err)
	}
	info.Espresso = espresso != 0
	rlpBytes, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("invalid ecotone l1 block info format: %w", err)
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

	if !solabi.EmptyReader(r) {
		return errors.New("too many bytes")
	}
	return nil
}

// isEcotoneButNotFirstBlock returns whether the specified block is subject to the Ecotone upgrade,
// but is not the actiation block itself.
func isEcotoneButNotFirstBlock(rollupCfg *rollup.Config, l2BlockTime uint64) bool {
	return rollupCfg.IsEcotone(l2BlockTime) && !rollupCfg.IsEcotoneActivationBlock(l2BlockTime)
}

// L1BlockInfoFromBytes is the inverse of L1InfoDeposit, to see where the L2 chain is derived from
func L1BlockInfoFromBytes(rollupCfg *rollup.Config, l2BlockTime uint64, data []byte) (*L1BlockInfo, error) {
	var info L1BlockInfo
	if isEcotoneButNotFirstBlock(rollupCfg, l2BlockTime) {
		return &info, info.unmarshalBinaryEcotone(data)
	}
	return &info, info.unmarshalBinaryBedrock(data)
}

// L1InfoDeposit creates a L1 Info deposit transaction based on the L1 block,
// and the L2 block-height difference with the start of the epoch.
func L1InfoDeposit(rollupCfg *rollup.Config, sysCfg eth.SystemConfig, seqNumber uint64, block eth.BlockInfo, l2BlockTime uint64, justification *eth.L2BatchJustification) (*types.DepositTx, error) {
	l1BlockInfo := L1BlockInfo{
		Number:              block.NumberU64(),
		Time:                block.Time(),
		BaseFee:             block.BaseFee(),
		BlockHash:           block.Hash(),
		SequenceNumber:      seqNumber,
		BatcherAddr:         sysCfg.BatcherAddr,
		Espresso:            sysCfg.Espresso,
		EspressoL1ConfDepth: sysCfg.EspressoL1ConfDepth,
		Justification:       justification,
	}
	var data []byte
	if isEcotoneButNotFirstBlock(rollupCfg, l2BlockTime) {
		l1BlockInfo.BlobBaseFee = block.BlobBaseFee()
		if l1BlockInfo.BlobBaseFee == nil {
			// The L2 spec states to use the MIN_BLOB_GASPRICE from EIP-4844 if not yet active on L1.
			l1BlockInfo.BlobBaseFee = big.NewInt(1)
		}
		blobBaseFeeScalar, baseFeeScalar, err := sysCfg.EcotoneScalars()
		if err != nil {
			return nil, err
		}
		l1BlockInfo.BlobBaseFeeScalar = blobBaseFeeScalar
		l1BlockInfo.BaseFeeScalar = baseFeeScalar
		out, err := l1BlockInfo.marshalBinaryEcotone()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal Ecotone l1 block info: %w", err)
		}
		data = out
	} else {
		l1BlockInfo.L1FeeOverhead = sysCfg.Overhead
		l1BlockInfo.L1FeeScalar = sysCfg.Scalar
		out, err := l1BlockInfo.marshalBinaryBedrock()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal Bedrock l1 block info: %w", err)
		}
		data = out
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
	if rollupCfg.IsRegolith(l2BlockTime) {
		out.IsSystemTransaction = false
		out.Gas = RegolithSystemTxGas
	}
	return out, nil
}

// L1InfoDepositBytes returns a serialized L1-info attributes transaction.
func L1InfoDepositBytes(rollupCfg *rollup.Config, sysCfg eth.SystemConfig, seqNumber uint64, l1Info eth.BlockInfo, l2BlockTime uint64, justification *eth.L2BatchJustification) ([]byte, error) {
	dep, err := L1InfoDeposit(rollupCfg, sysCfg, seqNumber, l1Info, l2BlockTime, justification)
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

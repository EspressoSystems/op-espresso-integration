package derive

import (
	"bytes"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum-optimism/optimism/op-node/testutils"
	"github.com/ethereum-optimism/optimism/op-service/eth"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"
)

// basic implementation of the Compressor interface that does no compression
type nonCompressor struct {
	bytes.Buffer
}

func (s *nonCompressor) Flush() error {
	return nil
}

func (s *nonCompressor) Close() error {
	return nil
}

func (s *nonCompressor) FullErr() error {
	return nil
}

func TestChannelOutAddBlock(t *testing.T) {
	cout, err := NewChannelOut(&nonCompressor{})
	require.NoError(t, err)

	t.Run("returns err if first tx is not an l1info tx", func(t *testing.T) {
		header := &types.Header{Number: big.NewInt(1), Difficulty: big.NewInt(100)}
		block := types.NewBlockWithHeader(header).WithBody(
			[]*types.Transaction{
				types.NewTx(&types.DynamicFeeTx{}),
			},
			nil,
		)
		_, err := cout.AddBlock(block)
		require.Error(t, err)
		require.Equal(t, ErrNotDepositTx, err)
	})
}

// TestOutputFrameSmallMaxSize tests that calling [OutputFrame] with a small
// max size that is below the fixed frame size overhead of 23, will return
// an error.
func TestOutputFrameSmallMaxSize(t *testing.T) {
	cout, err := NewChannelOut(&nonCompressor{})
	require.NoError(t, err)

	// Call OutputFrame with the range of small max size values that err
	var w bytes.Buffer
	for i := 0; i < 23; i++ {
		fid, err := cout.OutputFrame(&w, uint64(i))
		require.ErrorIs(t, err, ErrMaxFrameSizeTooSmall)
		require.Zero(t, fid)
	}
}

// TestRLPByteLimit ensures that stream encoder is properly limiting the length.
// It will decode the input if `len(input) <= inputLimit`.
func TestRLPByteLimit(t *testing.T) {
	// Should succeed if `len(input) == inputLimit`
	enc := []byte("\x8bhello world") // RLP encoding of the string "hello world"
	in := bytes.NewBuffer(enc)
	var out string
	stream := rlp.NewStream(in, 12)
	err := stream.Decode(&out)
	require.Nil(t, err)
	require.Equal(t, out, "hello world")

	// Should fail if the `inputLimit = len(input) - 1`
	enc = []byte("\x8bhello world") // RLP encoding of the string "hello world"
	in = bytes.NewBuffer(enc)
	var out2 string
	stream = rlp.NewStream(in, 11)
	err = stream.Decode(&out2)
	require.Equal(t, err, rlp.ErrValueTooLarge)
	require.Equal(t, out2, "")
}

func TestForceCloseTxData(t *testing.T) {
	id := [16]byte{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef}
	tests := []struct {
		frames []Frame
		errors bool
		output string
	}{
		{
			frames: []Frame{},
			errors: true,
			output: "",
		},
		{
			frames: []Frame{Frame{FrameNumber: 0, IsLast: false}, Frame{ID: id, FrameNumber: 1, IsLast: true}},
			errors: true,
			output: "",
		},
		{
			frames: []Frame{Frame{ID: id, FrameNumber: 0, IsLast: false}},
			errors: false,
			output: "00deadbeefdeadbeefdeadbeefdeadbeef00000000000001",
		},
		{
			frames: []Frame{Frame{ID: id, FrameNumber: 0, IsLast: true}},
			errors: false,
			output: "00",
		},
		{
			frames: []Frame{Frame{ID: id, FrameNumber: 1, IsLast: false}},
			errors: false,
			output: "00deadbeefdeadbeefdeadbeefdeadbeef00000000000001",
		},
		{
			frames: []Frame{Frame{ID: id, FrameNumber: 1, IsLast: true}},
			errors: false,
			output: "00deadbeefdeadbeefdeadbeefdeadbeef00000000000000",
		},
		{
			frames: []Frame{Frame{ID: id, FrameNumber: 2, IsLast: true}},
			errors: false,
			output: "00deadbeefdeadbeefdeadbeefdeadbeef00000000000000deadbeefdeadbeefdeadbeefdeadbeef00010000000000",
		},
		{
			frames: []Frame{Frame{ID: id, FrameNumber: 1, IsLast: false}, Frame{ID: id, FrameNumber: 3, IsLast: true}},
			errors: false,
			output: "00deadbeefdeadbeefdeadbeefdeadbeef00000000000000deadbeefdeadbeefdeadbeefdeadbeef00020000000000",
		},
		{
			frames: []Frame{Frame{ID: id, FrameNumber: 1, IsLast: false}, Frame{ID: id, FrameNumber: 3, IsLast: true}, Frame{ID: id, FrameNumber: 5, IsLast: true}},
			errors: false,
			output: "00deadbeefdeadbeefdeadbeefdeadbeef00000000000000deadbeefdeadbeefdeadbeefdeadbeef00020000000000",
		},
	}

	for i, test := range tests {
		out, err := ForceCloseTxData(test.frames)
		if test.errors {
			require.NotNil(t, err, "Should error on tc %v", i)
			require.Nil(t, out, "Should return no value in tc %v", i)
		} else {
			require.NoError(t, err, "Should not error on tc %v", i)
			require.Equal(t, common.FromHex(test.output), out, "Should match output tc %v", i)
		}
	}
}

func TestBlockToBatchValidity(t *testing.T) {
	block := new(types.Block)
	_, _, err := BlockToBatch(block)
	require.ErrorContains(t, err, "has no transactions")
}

func TestBlockToBatchAllTransactionTypes(t *testing.T) {
	rng := rand.New(rand.NewSource(12345))

	// Build a block from transactions in the following order:
	// * L1 info deposit
	// * Regular deposit
	// * Regular transaction
	// * Rejected transaction
	// * Rejected transaction
	// * Regular transaction
	rejected1 := types.RejectedTransaction{
		Pos:  3,
		Data: []byte{1, 2, 3, 4, 5},
	}
	rejected2 := types.RejectedTransaction{
		Pos:  3,
		Data: []byte{6, 7, 8, 9},
	}
	accepted1 := types.NewTx(&types.LegacyTx{
		Nonce: 0,
	})
	accepted2 := types.NewTx(&types.LegacyTx{
		Nonce: 1,
	})
	deposit := types.NewTx(&types.DepositTx{})
	l1Info, err := L1InfoDeposit(0, eth.HeaderBlockInfo(testutils.RandomHeader(rng)), eth.SystemConfig{}, nil, false)
	require.Nil(t, err, "failed to build L1 deposit info")

	block := types.NewBlockWithHeader(testutils.RandomHeader(rng)).
		WithBody([]*types.Transaction{types.NewTx(l1Info), deposit, accepted1, accepted2}, nil).
		WithRejected([]types.RejectedTransaction{rejected1, rejected2})

	batch, _, err := BlockToBatch(block)
	require.Nil(t, err, "BlockToBatch failed")

	accepted1Bytes, err := accepted1.MarshalBinary()
	require.Nil(t, err, "MarshalBinary failed")
	accepted2Bytes, err := accepted2.MarshalBinary()
	require.Nil(t, err, "MarshalBinary failed")

	require.Equal(t, batch.Transactions, []hexutil.Bytes{accepted1Bytes, rejected1.Data, rejected2.Data, accepted2Bytes})
}

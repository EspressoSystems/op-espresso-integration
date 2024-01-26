package derive

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	espresso "github.com/EspressoSystems/espresso-sequencer-go/types"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type EspressoValidBatchTestCase struct {
	Name       string
	L1Blocks   []eth.L1BlockRef
	L2SafeHead eth.L2BlockRef
	Batch      BatchWithL1InclusionBlock
	Expected   BatchValidity
	Headers    []espresso.Header
}

type mockL1Provider struct {
	L1Blocks []eth.L1BlockRef
	Headers  []espresso.Header
}

func (m *mockL1Provider) L1BlockRefByNumber(ctx context.Context, num uint64) (eth.L1BlockRef, error) {
	if num >= uint64(len(m.L1Blocks)) {
		return eth.L1BlockRef{}, fmt.Errorf("L1 block number %d not available", num)
	}
	return m.L1Blocks[num], nil
}

func (m *mockL1Provider) FetchReceipts(ctx context.Context, blockHash common.Hash) (eth.BlockInfo, types.Receipts, error) {
	return nil, nil, fmt.Errorf("not implemented: FetchReceipts")
}

func (m *mockL1Provider) VerifyCommitments(firstBlockHeight uint64, comms []espresso.Commitment) (bool, error) {
	if int(firstBlockHeight)+len(comms) > len(m.Headers) {
		return false, NewCriticalError(errors.New("Headers unavailable"))
	}

	for i, comm := range comms {
		if !comm.Equals(m.Headers[int(firstBlockHeight)+i].Commit()) {
			return false, nil
		}
	}
	return true, nil
}

func (m *mockL1Provider) setBlocks(blocks []eth.L1BlockRef) {
	m.L1Blocks = blocks
}

func (m *mockL1Provider) setHeaders(headers []espresso.Header) {
	m.Headers = headers
}

func TestValidBatchEspresso(t *testing.T) {
	sysCfg := eth.SystemConfig{
		Espresso:            true,
		EspressoL1ConfDepth: 0,
	}
	conf := rollup.Config{
		Genesis: rollup.Genesis{
			L2Time: 31, // a genesis time that itself does not align to make it more interesting
		},
		BlockTime:         2,
		SeqWindowSize:     4,
		MaxSequencerDrift: 6,
		L2ChainID:         big.NewInt(901),
		// other config fields are ignored and can be left empty.
	}

	rng := rand.New(rand.NewSource(1234))
	l1A := eth.L1BlockRef{
		Hash:       testutils.RandomHash(rng),
		Number:     0,
		ParentHash: testutils.RandomHash(rng),
		Time:       rng.Uint64(),
	}
	l1B := eth.L1BlockRef{
		Hash:       testutils.RandomHash(rng),
		Number:     l1A.Number + 1,
		ParentHash: l1A.Hash,
		Time:       l1A.Time + 7,
	}
	l1C := eth.L1BlockRef{
		Hash:       testutils.RandomHash(rng),
		Number:     l1B.Number + 1,
		ParentHash: l1B.Hash,
		Time:       l1B.Time + 7,
	}

	l2Parent := eth.L2BlockRef{
		Hash:           testutils.RandomHash(rng),
		Number:         99,
		ParentHash:     testutils.RandomHash(rng),
		Time:           l1A.Time - conf.BlockTime,
		L1Origin:       l1A.ID(),
		SequenceNumber: 0,
	}

	l2A0 := eth.L2BlockRef{
		Hash:           testutils.RandomHash(rng),
		Number:         l2Parent.Number + 1,
		ParentHash:     l2Parent.Hash,
		Time:           l2Parent.Time + conf.BlockTime,
		L1Origin:       l1A.ID(),
		SequenceNumber: 0,
	}

	l2A1 := eth.L2BlockRef{
		Hash:           testutils.RandomHash(rng),
		Number:         l2A0.Number + 1,
		ParentHash:     l2A0.Hash,
		Time:           l2A0.Time + conf.BlockTime,
		L1Origin:       l1A.ID(),
		SequenceNumber: 1,
	}

	l2A2 := eth.L2BlockRef{
		Hash:           testutils.RandomHash(rng),
		Number:         l2A1.Number + 1,
		ParentHash:     l2A1.Hash,
		Time:           l2A1.Time + conf.BlockTime,
		L1Origin:       l1A.ID(),
		SequenceNumber: 2,
	}

	l2A3 := eth.L2BlockRef{
		Hash:           testutils.RandomHash(rng),
		Number:         l2A2.Number + 1,
		ParentHash:     l2A2.Hash,
		Time:           l2A2.Time + conf.BlockTime,
		L1Origin:       l1A.ID(),
		SequenceNumber: 3,
	}

	l2B0 := eth.L2BlockRef{
		Hash:           testutils.RandomHash(rng),
		Number:         l2A3.Number + 1,
		ParentHash:     l2A3.Hash,
		Time:           l2A3.Time + conf.BlockTime, // 8 seconds larger than l1A0, 1 larger than origin
		L1Origin:       l1B.ID(),
		SequenceNumber: 0,
	}

	// Three valid windows, with varying numbers of HotShot blocks in the window.
	hotshotHeaders := []espresso.Header{
		{
			Height:    0,
			Timestamp: l2A1.Time - 1,
		},
		{
			Height:    1,
			Timestamp: l2A1.Time,
		},
		{
			Height:    2,
			Timestamp: l2A2.Time,
		},
		{
			Height:    3,
			Timestamp: l2A2.Time + 1,
		},
		{
			Height:    4,
			Timestamp: l2A3.Time,
		},
		{
			Height:    5,
			Timestamp: l2A3.Time + 1,
		},
		{
			Height:    6,
			Timestamp: l2A3.Time + 1,
		},
		{
			Height:    7,
			Timestamp: l2A3.Time + conf.BlockTime,
		},
	}

	// Hotshot skipped an L1 block
	hotshotSkippedHeaders := []espresso.Header{
		{
			Height:    0,
			Timestamp: l2B0.Time - 1,
		},
		{
			Height:    1,
			Timestamp: l2B0.Time,
			L1Head:    l2A3.L1Origin.Number + 2,
		},
		{
			Height:    2,
			Timestamp: l2B0.Time + conf.BlockTime,
			L1Head:    l2A3.L1Origin.Number + 2,
		},
	}

	// Case where Hotshot window is genuinely empty
	emptyHotshotWindowHeaders :=
		[]espresso.Header{
			{
				Height:    0,
				Timestamp: l2A1.Time - 1,
			},
			{
				Height:    1,
				Timestamp: l2A1.Time + 1000,
			},
		}

	// Case where Espresso tries to fool validator by providing a previous batch last block
	// That is greater than the window range.
	hotshotDishonestHeaders :=
		[]espresso.Header{
			{
				Height:    0,
				Timestamp: l2B0.Time - 1,
			},
			{
				Height:    1,
				Timestamp: l2B0.Time + 1000,
			},
			{
				Height:    2,
				Timestamp: l2B0.Time + 1001,
			},
		}

	testCases := []EspressoValidBatchTestCase{
		{
			Name:       "valid batch where one hotshot block falls within the window",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A0,
			Headers:    hotshotHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1A,
				Batch: &SingularBatch{
					ParentHash: l2A1.ParentHash,
					EpochNum:   rollup.Epoch(l2A1.L1Origin.Number),
					EpochHash:  l2A1.L1Origin.Hash,
					Timestamp:  l2A1.Time,
					Justification: &eth.L2BatchJustification{
						Prev: &hotshotHeaders[0],
						Blocks: []eth.EspressoBlockJustification{
							{
								Header: hotshotHeaders[1],
								Proof:  nil,
							},
						},
						Next: &hotshotHeaders[2],
					},
				},
			},
			Expected: BatchAccept,
		},
		{
			Name:       "valid batch where two hotshot blocks fall within the window",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A1,
			Headers:    hotshotHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1A,
				Batch: &SingularBatch{
					ParentHash: l2A2.ParentHash,
					EpochNum:   rollup.Epoch(l2A2.L1Origin.Number),
					EpochHash:  l2A2.L1Origin.Hash,
					Timestamp:  l2A2.Time,
					Justification: &eth.L2BatchJustification{
						Prev: &hotshotHeaders[1],
						Blocks: []eth.EspressoBlockJustification{
							{
								Header: hotshotHeaders[2],
								Proof:  nil,
							},
							{
								Header: hotshotHeaders[3],
								Proof:  nil,
							},
						},
						Next: &hotshotHeaders[4],
					},
				},
			},
			Expected: BatchAccept,
		},
		{
			Name:       "valid batch where three hotshot blocks fall within the window",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A2,
			Headers:    hotshotHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1A,
				Batch: &SingularBatch{
					ParentHash: l2A3.ParentHash,
					EpochNum:   rollup.Epoch(l2A3.L1Origin.Number),
					EpochHash:  l2A3.L1Origin.Hash,
					Timestamp:  l2A3.Time,
					Justification: &eth.L2BatchJustification{
						Prev: &hotshotHeaders[3],
						Blocks: []eth.EspressoBlockJustification{
							{
								Header: hotshotHeaders[4],
								Proof:  nil,
							},
							{
								Header: hotshotHeaders[5],
								Proof:  nil,
							},
							{
								Header: hotshotHeaders[6],
								Proof:  nil,
							},
						},
						Next: &hotshotHeaders[7],
					},
				},
			},
			Expected: BatchAccept,
		},
		{
			Name:       "empty batch due to empty hotshot window",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A0,
			Headers:    emptyHotshotWindowHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1A,
				Batch: &SingularBatch{
					ParentHash: l2A1.ParentHash,
					EpochNum:   rollup.Epoch(l2A1.L1Origin.Number),
					EpochHash:  l2A1.L1Origin.Hash,
					Timestamp:  l2A1.Time,
					Justification: &eth.L2BatchJustification{
						Prev: &emptyHotshotWindowHeaders[0],
						Next: &emptyHotshotWindowHeaders[1],
					},
				},
			},
			Expected: BatchAccept,
		},
		{
			Name:       "valid batch where HotShot skips an L1 block",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A3,
			Headers:    hotshotSkippedHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1B,
				Batch: &SingularBatch{
					ParentHash:   l2B0.ParentHash,
					EpochNum:     rollup.Epoch(l2B0.L1Origin.Number),
					EpochHash:    l2B0.L1Origin.Hash,
					Timestamp:    l2B0.Time,
					Transactions: []hexutil.Bytes{},
					Justification: &eth.L2BatchJustification{
						Prev: &hotshotSkippedHeaders[0],
						Blocks: []eth.EspressoBlockJustification{
							{
								Header: hotshotSkippedHeaders[1],
								Proof:  nil,
							},
						},
						Next: &hotshotSkippedHeaders[2],
					},
				},
			},
			Expected: BatchAccept,
		},
		{
			Name:       "invalid batch due to invalid headers",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A3,
			Headers:    hotshotSkippedHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1B,
				Batch: &SingularBatch{
					ParentHash:   l2B0.ParentHash,
					EpochNum:     rollup.Epoch(l2B0.L1Origin.Number),
					EpochHash:    l2B0.L1Origin.Hash,
					Timestamp:    l2B0.Time,
					Transactions: []hexutil.Bytes{},
					Justification: &eth.L2BatchJustification{
						// Switch the blocks
						Prev: &hotshotSkippedHeaders[1],
						Next: &hotshotSkippedHeaders[0],
					},
				},
			},
			Expected: BatchDrop,
		},
		{
			Name:       "invalid batch due to espresso providing a previous batch header outside of the window range",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A3,
			Headers:    hotshotDishonestHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1B,
				Batch: &SingularBatch{
					ParentHash:   l2B0.ParentHash,
					EpochNum:     rollup.Epoch(l2B0.L1Origin.Number),
					EpochHash:    l2B0.L1Origin.Hash,
					Timestamp:    l2B0.Time,
					Transactions: []hexutil.Bytes{},
					Justification: &eth.L2BatchJustification{
						Prev: &hotshotDishonestHeaders[0],
						Blocks: []eth.EspressoBlockJustification{
							{
								Header: hotshotDishonestHeaders[1],
								Proof:  nil,
							},
						},
						Next: &hotshotDishonestHeaders[2],
					},
				},
			},
			Expected: BatchDrop,
		},
		{
			Name:     "invalid batch when hotshot skips an L1 block",
			L1Blocks: []eth.L1BlockRef{l1A, l1B, l1C},
			// In this case, the L1 origin wont increment
			L2SafeHead: l2A3,
			Headers:    hotshotSkippedHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1B,
				Batch: &SingularBatch{
					ParentHash:   l2B0.ParentHash,
					EpochNum:     rollup.Epoch(l2A3.L1Origin.Number),
					EpochHash:    l2A3.L1Origin.Hash,
					Timestamp:    l2B0.Time,
					Transactions: []hexutil.Bytes{},
					Justification: &eth.L2BatchJustification{
						Prev: &hotshotSkippedHeaders[0],
						Blocks: []eth.EspressoBlockJustification{
							{
								Header: hotshotSkippedHeaders[1],
								Proof:  nil,
							},
						},
						Next: &hotshotSkippedHeaders[2],
					},
				},
			},
			Expected: BatchDrop,
		},
		{
			Name:       "invalid batch due to a HotShot block falling outside of the transaction window",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A0,
			Headers:    hotshotHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1A,
				Batch: &SingularBatch{
					ParentHash: l2A1.ParentHash,
					EpochNum:   rollup.Epoch(l2A1.L1Origin.Number),
					EpochHash:  l2A1.L1Origin.Hash,
					Timestamp:  l2A1.Time,
					Justification: &eth.L2BatchJustification{
						Prev: &hotshotHeaders[0],
						Blocks: []eth.EspressoBlockJustification{
							{
								Header: hotshotHeaders[1],
								Proof:  nil,
							},
							// Include an extra block that is outside the window
							{
								Header: hotshotHeaders[2],
								Proof:  nil,
							},
						},
						// Increment Next from the valid test case by one
						Next: &hotshotHeaders[3],
					},
				},
			},
			Expected: BatchDrop,
		},
		{
			Name:     "invalid batch due to lack of justification",
			L1Blocks: []eth.L1BlockRef{l1A, l1B, l1C},
			// In this case, the L1 origin wont increment
			L2SafeHead: l2B0,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1B,
				Batch: &SingularBatch{
					ParentHash:   l2B0.ParentHash,
					EpochNum:     rollup.Epoch(l2B0.L1Origin.Number),
					EpochHash:    l2B0.L1Origin.Hash,
					Timestamp:    l2B0.Time,
					Transactions: []hexutil.Bytes{},
				},
			},
			Expected: BatchDrop,
		},
		{
			Name:       "undecided batch if headers are not available",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A3,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1A,
				Batch: &SingularBatch{
					ParentHash: l2B0.ParentHash,
					EpochNum:   rollup.Epoch(l2A0.L1Origin.Number),
					EpochHash:  l2A0.L1Origin.Hash,
					Timestamp:  l2B0.Time,
					Justification: &eth.L2BatchJustification{
						Prev: &emptyHotshotWindowHeaders[0],
						Next: &emptyHotshotWindowHeaders[1],
					},
				},
			},
			Expected: BatchUndecided,
		},
		{
			Name:       "valid batch genesis",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2Parent,
			Headers:    hotshotHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1A,
				Batch: &SingularBatch{
					ParentHash: l2A0.ParentHash,
					EpochNum:   rollup.Epoch(l2A0.L1Origin.Number),
					EpochHash:  l2A0.L1Origin.Hash,
					Timestamp:  l2A0.Time,
					Justification: &eth.L2BatchJustification{
						Prev: nil,
						Blocks: []eth.EspressoBlockJustification{
							{
								Header: hotshotHeaders[0],
								Proof:  nil,
							},
						},
						Next: &hotshotHeaders[1],
					},
				},
			},
			Expected: BatchAccept,
		},
		{
			Name:       "invalid batch missing prev not genesis",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A0,
			Headers:    hotshotHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1A,
				Batch: &SingularBatch{
					ParentHash: l2A1.ParentHash,
					EpochNum:   rollup.Epoch(l2A1.L1Origin.Number),
					EpochHash:  l2A1.L1Origin.Hash,
					Timestamp:  l2A1.Time,
					Justification: &eth.L2BatchJustification{
						Prev: nil,
						Blocks: []eth.EspressoBlockJustification{
							{
								Header: hotshotHeaders[1],
								Proof:  nil,
							},
						},
						Next: &hotshotHeaders[2],
					},
				},
			},
			Expected: BatchDrop,
		},
		{
			Name:       "invalid batch missing prev genesis before window start",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A0,
			Headers:    hotshotHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1A,
				Batch: &SingularBatch{
					ParentHash: l2A1.ParentHash,
					EpochNum:   rollup.Epoch(l2A1.L1Origin.Number),
					EpochHash:  l2A1.L1Origin.Hash,
					Timestamp:  l2A1.Time,
					Justification: &eth.L2BatchJustification{
						Prev: nil,
						Blocks: []eth.EspressoBlockJustification{
							{
								Header: hotshotHeaders[0],
								Proof:  nil,
							},
							{
								Header: hotshotHeaders[1],
								Proof:  nil,
							},
						},
						Next: &hotshotHeaders[2],
					},
				},
			},
			Expected: BatchDrop,
		},
	}

	// Log level can be increased for debugging purposes
	logger := testlog.Logger(t, log.LvlWarn)

	var l1 = &mockL1Provider{}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			l1.setBlocks(testCase.L1Blocks)
			l1.setHeaders(testCase.Headers)
			ctx := context.Background()
			validity := CheckBatch(ctx, &sysCfg, &conf, logger, testCase.L1Blocks, testCase.L2SafeHead, &testCase.Batch, l1, nil)
			require.Equal(t, testCase.Expected, validity, "batch check must return expected validity level")
		})
	}
}

func TestL1OriginLag(t *testing.T) {
	sysCfg := eth.SystemConfig{
		Espresso:            true,
		EspressoL1ConfDepth: 2,
	}
	conf := rollup.Config{
		Genesis: rollup.Genesis{
			L2Time: 31, // a genesis time that itself does not align to make it more interesting
		},
		BlockTime:         2,
		SeqWindowSize:     4,
		MaxSequencerDrift: 17,
		L2ChainID:         big.NewInt(901),
		// other config fields are ignored and can be left empty.
	}

	rng := rand.New(rand.NewSource(1234))
	l1A := eth.L1BlockRef{
		Hash:       testutils.RandomHash(rng),
		Number:     0,
		ParentHash: testutils.RandomHash(rng),
		Time:       rng.Uint64(),
	}
	l1B := eth.L1BlockRef{
		Hash:       testutils.RandomHash(rng),
		Number:     l1A.Number + 1,
		ParentHash: l1A.Hash,
		Time:       l1A.Time + 7,
	}
	l1C := eth.L1BlockRef{
		Hash:       testutils.RandomHash(rng),
		Number:     l1B.Number + 1,
		ParentHash: l1B.Hash,
		Time:       l1B.Time + 7,
	}

	headers := []espresso.Header{
		{
			Height:    0,
			Timestamp: l1C.Time,
			L1Head:    l1C.Number,
		},
		{
			Height:    1,
			Timestamp: l1C.Time + 2*conf.BlockTime + 1,
			L1Head:    l1C.Number,
		},
	}

	l2SafeHead := eth.L2BlockRef{
		Hash:           testutils.RandomHash(rng),
		Number:         100,
		ParentHash:     testutils.RandomHash(rng),
		Time:           l1C.Time,
		L1Origin:       l1A.ID(),
		SequenceNumber: 0,
	}

	testCases := []EspressoValidBatchTestCase{
		{
			Name:       "valid origin lag",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2SafeHead,
			Headers:    headers,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1C,
				Batch: &SingularBatch{
					ParentHash: l2SafeHead.Hash,
					EpochNum:   rollup.Epoch(l1A.Number),
					EpochHash:  l1A.Hash,
					Timestamp:  l2SafeHead.Time + conf.BlockTime,
					Justification: &eth.L2BatchJustification{
						Prev: &headers[0],
						Next: &headers[1],
					},
				},
			},
			Expected: BatchAccept,
		},
		{
			Name:       "missing lag",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2SafeHead,
			Headers:    headers,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1C,
				Batch: &SingularBatch{
					ParentHash: l2SafeHead.Hash,
					EpochNum:   rollup.Epoch(l1B.Number),
					EpochHash:  l1B.Hash,
					Timestamp:  l2SafeHead.Time + conf.BlockTime,
					Justification: &eth.L2BatchJustification{
						Prev: &headers[0],
						Next: &headers[1],
					},
				},
			},
			Expected: BatchDrop,
		},
	}

	// Log level can be increased for debugging purposes
	logger := testlog.Logger(t, log.LvlWarn)

	var l1 = &mockL1Provider{}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			l1.setBlocks(testCase.L1Blocks)
			l1.setHeaders(testCase.Headers)
			ctx := context.Background()
			validity := CheckBatch(ctx, &sysCfg, &conf, logger, testCase.L1Blocks, testCase.L2SafeHead, &testCase.Batch, l1, nil)
			require.Equal(t, testCase.Expected, validity, "batch check must return expected validity level")
		})
	}
}

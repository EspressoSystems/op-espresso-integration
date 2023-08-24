package derive

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/testlog"
	"github.com/ethereum-optimism/optimism/op-node/testutils"
	"github.com/ethereum-optimism/optimism/op-service/espresso"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
)

type mockHotShotProvider struct {
}

func (m *mockHotShotProvider) verifyHeaders(headers []espresso.Header, height uint64) error {
	return nil
}

func (m *mockHotShotProvider) getHeadersFromHeight(firstBlockHeight uint64, numHeaders uint64) ([]espresso.Header, error) {
	return []espresso.Header{}, nil
}

func TestValidBatchEspresso(t *testing.T) {
	conf := rollup.Config{
		Genesis: rollup.Genesis{
			L2Time: 31, // a genesis time that itself does not align to make it more interesting
		},
		BlockTime:         2,
		SeqWindowSize:     4,
		MaxSequencerDrift: 6,
		// other config fields are ignored and can be left empty.
	}

	rng := rand.New(rand.NewSource(1234))
	l1A := testutils.RandomBlockRef(rng)
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

	l2A0 := eth.L2BlockRef{
		Hash:           testutils.RandomHash(rng),
		Number:         100,
		ParentHash:     testutils.RandomHash(rng),
		Time:           l1A.Time,
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

	testCases := []ValidBatchTestCase{
		{
			Name:       "valid batch where hotshot transactions fall within the window",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A3,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1A,
				Batch: &BatchData{BatchV2{
					BatchV1: BatchV1{
						ParentHash: l2B0.ParentHash,
						EpochNum:   rollup.Epoch(l2B0.L1Origin.Number),
						EpochHash:  l2B0.L1Origin.Hash,
						Timestamp:  l2B0.Time,
						Transactions: []hexutil.Bytes{
							[]byte{0x02, 0x42, 0x13, 0x37},
							[]byte{0x02, 0xde, 0xad, 0xbe, 0xef},
						},
					},
					Justification: &eth.L2BatchJustification{
						FirstBlock: espresso.Header{
							Timestamp: l2B0.Time,
						},
						PrevBatchLastBlock: espresso.Header{
							Timestamp: l2B0.Time - 1,
						},
						FirstBlockNumber: 1,
						Payload: &eth.L2BatchPayloadJustification{
							LastBlock: espresso.Header{
								Timestamp: l2B0.Time + conf.BlockTime - 1,
							},
							NextBatchFirstBlock: espresso.Header{
								Timestamp: l2B0.Time + conf.BlockTime,
							},
							NmtProofs: []espresso.Bytes{
								[]byte{0x02, 0x42, 0x13, 0x37},
								[]byte{0x02, 0xde, 0xad, 0xbe, 0xef},
							},
						},
					},
				}},
			},
			Expected: BatchAccept,
		},
		{
			Name:       "valid batch where hotshot transactions fall within the window and there is a block in between first and last block",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A3,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1A,
				Batch: &BatchData{BatchV2{
					BatchV1: BatchV1{
						ParentHash: l2B0.ParentHash,
						EpochNum:   rollup.Epoch(l2B0.L1Origin.Number),
						EpochHash:  l2B0.L1Origin.Hash,
						Timestamp:  l2B0.Time,
						Transactions: []hexutil.Bytes{
							[]byte{0x02, 0x42, 0x13, 0x37},
							[]byte{0x02, 0xde, 0xad, 0xbe, 0xef},
						},
					},
					Justification: &eth.L2BatchJustification{
						FirstBlock: espresso.Header{
							Timestamp: l2B0.Time,
						},
						PrevBatchLastBlock: espresso.Header{
							Timestamp: l2B0.Time - 1,
						},
						FirstBlockNumber: 1,
						Payload: &eth.L2BatchPayloadJustification{
							LastBlock: espresso.Header{
								Timestamp: l2B0.Time + conf.BlockTime - 1,
							},
							NextBatchFirstBlock: espresso.Header{
								Timestamp: l2B0.Time + conf.BlockTime,
							},
							NmtProofs: []espresso.Bytes{
								[]byte{0x02, 0x42, 0x13, 0x37},
								[]byte{0x02, 0x42, 0x13, 0x37},
								[]byte{0x02, 0xde, 0xad, 0xbe, 0xef},
							},
						},
					},
				}},
			},
			Expected: BatchAccept,
		},
		{
			Name:       "empty batch due to empty hotshot window",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A3,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1A,
				Batch: &BatchData{BatchV2{
					BatchV1: BatchV1{
						ParentHash: l2B0.ParentHash,
						EpochNum:   rollup.Epoch(l2A0.L1Origin.Number),
						EpochHash:  l2A0.L1Origin.Hash,
						Timestamp:  l2B0.Time,
					},
					Justification: &eth.L2BatchJustification{
						FirstBlock: espresso.Header{
							Timestamp: l2B0.Time + 1000,
						},
						PrevBatchLastBlock: espresso.Header{
							Timestamp: l2B0.Time - 1,
						},
						FirstBlockNumber: 1,
					},
				}},
			},
			Expected: BatchAccept,
		},
		{
			Name:       "empty batch due to hotshot skipping an L1 block",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A3,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1B,
				Batch: &BatchData{BatchV2{
					BatchV1: BatchV1{
						ParentHash: l2B0.ParentHash,
						EpochNum:   rollup.Epoch(l2B0.L1Origin.Number),
						EpochHash:  l2B0.L1Origin.Hash,
						Timestamp:  l2B0.Time,
						Transactions: []hexutil.Bytes{
							[]byte{0x02, 0x42, 0x13, 0x37},
							[]byte{0x02, 0xde, 0xad, 0xbe, 0xef},
						},
					},
					Justification: &eth.L2BatchJustification{
						FirstBlock: espresso.Header{
							Timestamp: l2B0.Time,
							L1Block: espresso.L1BlockInfo{
								Number: l2A3.L1Origin.Number + 2,
							},
						},
						PrevBatchLastBlock: espresso.Header{
							Timestamp: l2B0.Time - 1,
						},
						FirstBlockNumber: 1,
					},
				}},
			},
			Expected: BatchAccept,
		},
	}

	// Log level can be increased for debugging purposes
	logger := testlog.Logger(t, log.LvlWarn)

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			validity := CheckBatch(&conf, logger, testCase.L1Blocks, testCase.L2SafeHead, &testCase.Batch, true, &mockHotShotProvider{})
			require.Equal(t, testCase.Expected, validity, "batch check must return expected validity level")
		})
	}
}

func TestDefaultBatchCasesEspresso(t *testing.T) {
	ValidBatch(t, true)
}

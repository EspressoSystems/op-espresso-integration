package derive

import (
	"errors"
	"fmt"
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

type EspressoValidBatchTestCase struct {
	Name       string
	L1Blocks   []eth.L1BlockRef
	L2SafeHead eth.L2BlockRef
	Batch      BatchWithL1InclusionBlock
	Expected   BatchValidity
	Headers    []espresso.Header
}

type mockHotShotProvider struct {
	Headers []espresso.Header
}

func (m *mockHotShotProvider) verifyHeaders(headers []espresso.Header, height uint64) (bool, error) {
	if height+uint64(len(headers)) > uint64(len(m.Headers)) {
		fmt.Println("Headers unavailable")
		return false, NewCriticalError(errors.New("Headers unavailable"))
	}
	// For testing purposes, use the timestamp to check equality
	for i, header := range headers {
		if header.Timestamp != m.Headers[uint64(i)+height].Timestamp {
			fmt.Println("Invalid header")
			return false, nil

		}
	}
	return true, nil
}

func (m *mockHotShotProvider) getHeadersFromHeight(firstBlockHeight uint64, numHeaders uint64) ([]espresso.Header, error) {
	if firstBlockHeight+numHeaders > uint64(len(m.Headers)) {
		fmt.Println("Headers unavailable")
		return nil, NewCriticalError(errors.New("Headers unavailable"))
	}
	return m.Headers[firstBlockHeight : firstBlockHeight+numHeaders], nil
}

func (m *mockHotShotProvider) setHeaders(headers []espresso.Header) {
	m.Headers = headers
}

func makeHeader(timestamp uint64) espresso.Header {
	return espresso.Header{
		Timestamp: timestamp,
	}
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

	// Two valid windows, one with a hotshot block in between the first and last blocks
	hotshotHeaders := []espresso.Header{
		makeHeader(l2A1.Time - 1),
		makeHeader(l2A1.Time),
		makeHeader(l2A2.Time - 1),
		makeHeader(l2A2.Time),
		makeHeader(l2A2.Time + 1),
		makeHeader(l2A3.Time - 1),
		makeHeader(l2A3.Time),
	}

	// Hotshot skipped an L1 block
	hotshotSkippedHeaders := []espresso.Header{
		makeHeader(
			l2B0.Time - 1,
		),
		{
			Timestamp: l2B0.Time,
			L1Block: espresso.L1BlockInfo{
				Number: l2A3.L1Origin.Number + 2,
			},
		},
	}

	// Case where Hotshot window is genuinely empty
	emptyHotshotWindowHeaders :=
		[]espresso.Header{
			makeHeader(l2B0.Time - 1),
			makeHeader(l2B0.Time + 1000),
		}

	// Case where Espresso tries to fool validator by providing a previous batch last block
	// That is greater than the window range.
	hotshotDishonestHeaders :=
		[]espresso.Header{
			makeHeader(l2B0.Time - 1),
			makeHeader(l2B0.Time + 1000),
		}

	testCases := []EspressoValidBatchTestCase{
		{
			Name:       "valid batch where hotshot transactions fall within the window",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A0,
			Headers:    hotshotHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1A,
				Batch: &BatchData{BatchV2{
					BatchV1: BatchV1{
						ParentHash: l2A1.ParentHash,
						EpochNum:   rollup.Epoch(l2A1.L1Origin.Number),
						EpochHash:  l2A1.L1Origin.Hash,
						Timestamp:  l2A1.Time,
					},
					Justification: &eth.L2BatchJustification{
						PrevBatchLastBlock: hotshotHeaders[0],
						FirstBlock:         hotshotHeaders[1],
						FirstBlockNumber:   1,
						Payload: &eth.L2BatchPayloadJustification{
							LastBlock:           hotshotHeaders[2],
							NextBatchFirstBlock: hotshotHeaders[3],
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
			L2SafeHead: l2A1,
			Headers:    hotshotHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1A,
				Batch: &BatchData{BatchV2{
					BatchV1: BatchV1{
						ParentHash: l2A2.ParentHash,
						EpochNum:   rollup.Epoch(l2A2.L1Origin.Number),
						EpochHash:  l2A2.L1Origin.Hash,
						Timestamp:  l2A2.Time,
					},
					Justification: &eth.L2BatchJustification{
						PrevBatchLastBlock: hotshotHeaders[2],
						FirstBlock:         hotshotHeaders[3],
						FirstBlockNumber:   3,
						Payload: &eth.L2BatchPayloadJustification{
							LastBlock:           hotshotHeaders[5],
							NextBatchFirstBlock: hotshotHeaders[6],
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
			Headers:    emptyHotshotWindowHeaders,
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
						PrevBatchLastBlock: emptyHotshotWindowHeaders[0],
						FirstBlock:         emptyHotshotWindowHeaders[1],
						FirstBlockNumber:   1,
					},
				}},
			},
			Expected: BatchAccept,
		},
		{
			Name:       "empty batch due to hotshot skipping an L1 block",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A3,
			Headers:    hotshotSkippedHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1B,
				Batch: &BatchData{BatchV2{
					BatchV1: BatchV1{
						ParentHash:   l2B0.ParentHash,
						EpochNum:     rollup.Epoch(l2B0.L1Origin.Number),
						EpochHash:    l2B0.L1Origin.Hash,
						Timestamp:    l2B0.Time,
						Transactions: []hexutil.Bytes{},
					},
					Justification: &eth.L2BatchJustification{
						PrevBatchLastBlock: hotshotSkippedHeaders[0],
						FirstBlock:         hotshotSkippedHeaders[1],
						FirstBlockNumber:   1,
					},
				}},
			},
			Expected: BatchAccept,
		},
		{
			Name:       "invalid batch due to empty headers",
			L1Blocks:   []eth.L1BlockRef{l1A, l1B, l1C},
			L2SafeHead: l2A3,
			Headers:    hotshotSkippedHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1B,
				Batch: &BatchData{BatchV2{
					BatchV1: BatchV1{
						ParentHash:   l2B0.ParentHash,
						EpochNum:     rollup.Epoch(l2B0.L1Origin.Number),
						EpochHash:    l2B0.L1Origin.Hash,
						Timestamp:    l2B0.Time,
						Transactions: []hexutil.Bytes{},
					},
					Justification: &eth.L2BatchJustification{
						// Switch the blocks
						PrevBatchLastBlock: hotshotSkippedHeaders[1],
						FirstBlock:         hotshotSkippedHeaders[0],
						FirstBlockNumber:   1,
					},
				}},
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
				Batch: &BatchData{BatchV2{
					BatchV1: BatchV1{
						ParentHash:   l2B0.ParentHash,
						EpochNum:     rollup.Epoch(l2B0.L1Origin.Number),
						EpochHash:    l2B0.L1Origin.Hash,
						Timestamp:    l2B0.Time,
						Transactions: []hexutil.Bytes{},
					},
					Justification: &eth.L2BatchJustification{
						PrevBatchLastBlock: hotshotDishonestHeaders[0],
						FirstBlock:         hotshotDishonestHeaders[1],
						FirstBlockNumber:   1,
					},
				}},
			},
			Expected: BatchDrop,
		},
		{
			Name:     "invalid batch when hotshot skips an L1 block",
			L1Blocks: []eth.L1BlockRef{l1A, l1B, l1C},
			// In this case, the L1 origin wont increment
			L2SafeHead: l2B0,
			Headers:    hotshotSkippedHeaders,
			Batch: BatchWithL1InclusionBlock{
				L1InclusionBlock: l1B,
				Batch: &BatchData{BatchV2{
					BatchV1: BatchV1{
						ParentHash:   l2B0.ParentHash,
						EpochNum:     rollup.Epoch(l2B0.L1Origin.Number),
						EpochHash:    l2B0.L1Origin.Hash,
						Timestamp:    l2B0.Time,
						Transactions: []hexutil.Bytes{},
					},
					Justification: &eth.L2BatchJustification{
						PrevBatchLastBlock: hotshotSkippedHeaders[0],
						FirstBlock:         hotshotSkippedHeaders[1],
						FirstBlockNumber:   1,
					},
				}},
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
				Batch: &BatchData{BatchV2{
					BatchV1: BatchV1{
						ParentHash: l2A1.ParentHash,
						EpochNum:   rollup.Epoch(l2A1.L1Origin.Number),
						EpochHash:  l2A1.L1Origin.Hash,
						Timestamp:  l2A1.Time,
					},
					Justification: &eth.L2BatchJustification{
						PrevBatchLastBlock: hotshotHeaders[0],
						FirstBlock:         hotshotHeaders[1],
						FirstBlockNumber:   1,
						Payload: &eth.L2BatchPayloadJustification{
							// Increment LastBlock and NextBatchFirstBlock from the valid test case by one
							LastBlock:           hotshotHeaders[3],
							NextBatchFirstBlock: hotshotHeaders[4],
							NmtProofs: []espresso.Bytes{
								[]byte{0x02, 0x42, 0x13, 0x37},
								[]byte{0x02, 0x42, 0x13, 0x37},
								[]byte{0x02, 0xde, 0xad, 0xbe, 0xef},
							},
						},
					},
				}},
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
				Batch: &BatchData{BatchV2{
					BatchV1: BatchV1{
						ParentHash:   l2B0.ParentHash,
						EpochNum:     rollup.Epoch(l2B0.L1Origin.Number),
						EpochHash:    l2B0.L1Origin.Hash,
						Timestamp:    l2B0.Time,
						Transactions: []hexutil.Bytes{},
					},
				}},
			},
			Expected: BatchDrop,
		},
		{
			Name:       "future batch if headers are not available",
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
						PrevBatchLastBlock: emptyHotshotWindowHeaders[0],
						FirstBlock:         emptyHotshotWindowHeaders[1],
						FirstBlockNumber:   1,
					},
				}},
			},
			Expected: BatchFuture,
		},
	}

	// Log level can be increased for debugging purposes
	logger := testlog.Logger(t, log.LvlWarn)

	var hotshot = &mockHotShotProvider{}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			hotshot.setHeaders(testCase.Headers)
			validity := CheckBatch(&conf, logger, testCase.L1Blocks, testCase.L2SafeHead, &testCase.Batch, true, hotshot)
			require.Equal(t, testCase.Expected, validity, "batch check must return expected validity level")
		})
	}
}

func TestDefaultBatchCasesEspresso(t *testing.T) {
	ValidBatch(t, true)
}

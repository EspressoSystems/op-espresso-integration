package derive

import (
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/espresso"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type BatchWithL1InclusionBlock struct {
	L1InclusionBlock eth.L1BlockRef
	Batch            *BatchData
}

type BatchValidity uint8

const (
	// BatchDrop indicates that the batch is invalid, and will always be in the future, unless we reorg
	BatchDrop = iota
	// BatchAccept indicates that the batch is valid and should be processed
	BatchAccept
	// BatchUndecided indicates we are lacking L1 information until we can proceed batch filtering
	BatchUndecided
	// BatchFuture indicates that the batch may be valid, but cannot be processed yet and should be checked again later
	BatchFuture
)

func CheckBatchEspresso(cfg *rollup.Config, log log.Logger, l1Blocks []eth.L1BlockRef, l2SafeHead eth.L2BlockRef, batch *BatchWithL1InclusionBlock, hotshot HotShotContractProvider) BatchValidity {
	// First, validate the headers that represent the beginning of the L2 block range of this batch.
	jst := batch.Batch.Justification
	if jst == nil {
		// Espresso blocks must have a justification
		log.Warn("No justification provided, dropping the batch.")
		return BatchDrop
	}
	// Case where the start of the window is the genesis block
	genesisStart := jst.FirstBlockNumber == 0
	var startHeaders []espresso.Header
	var firstBlockHeight uint64

	if genesisStart {
		startHeaders =
			[]espresso.Header{jst.FirstBlock}
		firstBlockHeight = 0
	} else {
		startHeaders =
			[]espresso.Header{jst.PrevBatchLastBlock, jst.FirstBlock}
		firstBlockHeight = jst.FirstBlockNumber - 1
	}

	validHeaders, err := hotshot.VerifyHeaders(startHeaders, firstBlockHeight)

	// In the case that the headers aren't available yet (perhaps the validator's L1 client is behind), return BatchFuture so that we can try again later
	// If the headers are available but invalid, drop the batch
	if err != nil {
		log.Warn("Headers unavailable, returning BatchUndecided.")
		return BatchUndecided
	} else if !validHeaders {
		log.Warn("Headers invalid, dropping the batch.")
		return BatchDrop
	}

	// First, check for cases where it is valid to have any empty batch.
	windowStart := l2SafeHead.Time + cfg.BlockTime
	windowEnd := windowStart + cfg.BlockTime
	payload := jst.Payload
	prevL1Origin := l2SafeHead.L1Origin

	// Ensure that the sequencer did not cheat by choosing a previous block that is within the sequencing window
	// Slightly redundant to the validRange check, but we need this check first to ensure that we don't get incorrect empty batches in the nil cases below
	if jst.FirstBlockNumber != 0 && jst.PrevBatchLastBlock.Timestamp >= windowStart {
		log.Warn("Dropping batch. The previous batch last block cannot be past the start of the sequencing window")
		return BatchDrop
	}

	// If Espresso did not produce any blocks in this window, an empty batch is valid.
	// In this case, the L1 origin must be the same as the previous block.
	if payload == nil && jst.FirstBlock.Timestamp >= windowEnd {
		if uint64(batch.Batch.EpochNum) != prevL1Origin.Number {
			log.Warn("Dropping batch. When HotShot has not seqeuenced anything in the batch window, the L1 origin must be the same as the prior block")
			return BatchDrop
		}
		return BatchAccept
	}

	// An empty batch can also be valid if the L1 origin is too far behind (either due to lag, or because HotShot skipped a block)
	// In this case, the L1 origin must increase by one
	skippedL1Block := jst.FirstBlock.L1Head > prevL1Origin.Number
	if payload == nil && skippedL1Block {
		if uint64(batch.Batch.EpochNum) != prevL1Origin.Number+1 {
			log.Warn("Dropping batch because the L1 origin did not increase by one")
			return BatchDrop
		}
		return BatchAccept
	}

	// At this point, there is no valid reason to have an empty payload
	if payload == nil {
		log.Warn("Dropping batch due to empty payload")
		return BatchDrop
	}

	// Now validate the transactions in the batch
	numBlocks := len(payload.NmtProofs)

	// Validate the headers representing the end of the batch window
	endHeaders :=
		[]espresso.Header{payload.LastBlock, payload.NextBatchFirstBlock}
	validHeaders, err = hotshot.VerifyHeaders(endHeaders, jst.FirstBlockNumber+uint64(numBlocks)-1)

	// In the case that the headers aren't available yet (perhaps the validator's L1 client is behind), return BatchFuture so that we can try again later
	// If the headers are available but invalid, drop the batch
	if err != nil {
		log.Warn("Headers unavailable, returning BatchFuture.")
		return BatchFuture
	} else if !validHeaders {
		log.Warn("Headers invalid, dropping the batch.")
		return BatchDrop
	}

	// Check that the range of HotShot blocks fall within the window
	validRange :=
		jst.PrevBatchLastBlock.Timestamp < windowStart &&
			jst.FirstBlock.Timestamp >= windowStart &&
			payload.LastBlock.Timestamp < windowEnd &&
			payload.NextBatchFirstBlock.Timestamp >= windowEnd

	if !validRange {
		log.Warn("Header range invalid, dropping the batch.")
		return BatchDrop
	}

	// Sanity check that the number of NMT proofs is at least 1 if the first and last block in the range are the same,
	// and at least 2 otherwise. This check would redundant to the ValidateBatchTransactions function below if
	// that function weren't mocked.
	var minimumNMTProofs = 1
	if jst.FirstBlock.Timestamp != jst.Payload.LastBlock.Timestamp {
		minimumNMTProofs = 2
	}

	if len(payload.NmtProofs) < minimumNMTProofs {
		log.Warn("Dropping batch, minimum NMT proofs required: ", minimumNMTProofs)
		return BatchDrop
	}

	// Validate the transactions against the NMT proofs
	comms, err := hotshot.GetCommitmentsFromHeight(firstBlockHeight, uint64(len(payload.NmtProofs)))
	if err != nil {
		// If we couldn't fetch headers, try again later
		log.Warn("Headers unavailable, returning BatchUndecided.")
		return BatchUndecided
	}
	err = espresso.ValidateBatchTransactions(batch.Batch.Transactions, payload.NmtProofs, comms)
	if err != nil {
		log.Warn("Error validating batch transactions, dropping the batch.")
		return BatchDrop
	}

	return BatchAccept

}

// CheckBatch checks if the given batch can be applied on top of the given l2SafeHead, given the contextual L1 blocks the batch was included in.
// The first entry of the l1Blocks should match the origin of the l2SafeHead. One or more consecutive l1Blocks should be provided.
// In case of only a single L1 block, the decision whether a batch is valid may have to stay undecided.
func CheckBatch(cfg *rollup.Config, log log.Logger, l1Blocks []eth.L1BlockRef, l2SafeHead eth.L2BlockRef, batch *BatchWithL1InclusionBlock, usingEspresso bool, hotshot HotShotContractProvider) BatchValidity {
	// add details to the log
	log = log.New(
		"batch_timestamp", batch.Batch.Timestamp,
		"parent_hash", batch.Batch.ParentHash,
		"batch_epoch", batch.Batch.Epoch(),
		"txs", len(batch.Batch.Transactions),
	)
	// sanity check we have consistent inputs
	if len(l1Blocks) == 0 {
		log.Warn("missing L1 block input, cannot proceed with batch checking")
		return BatchUndecided
	}
	epoch := l1Blocks[0]

	nextTimestamp := l2SafeHead.Time + cfg.BlockTime
	if batch.Batch.Timestamp > nextTimestamp {
		log.Trace("received out-of-order batch for future processing after next batch", "next_timestamp", nextTimestamp)
		return BatchFuture
	}
	if batch.Batch.Timestamp < nextTimestamp {
		log.Warn("dropping batch with old timestamp", "min_timestamp", nextTimestamp)
		return BatchDrop
	}

	// dependent on above timestamp check. If the timestamp is correct, then it must build on top of the safe head.
	if batch.Batch.ParentHash != l2SafeHead.Hash {
		log.Warn("ignoring batch with mismatching parent hash", "current_safe_head", l2SafeHead.Hash)
		return BatchDrop
	}

	// Filter out batches that were included too late.
	if uint64(batch.Batch.EpochNum)+cfg.SeqWindowSize < batch.L1InclusionBlock.Number {
		log.Warn("batch was included too late, sequence window expired")
		return BatchDrop
	}

	// Check the L1 origin of the batch
	batchOrigin := epoch
	if uint64(batch.Batch.EpochNum) < epoch.Number {
		log.Warn("dropped batch, epoch is too old", "minimum", epoch.ID())
		// batch epoch too old
		return BatchDrop
	} else if uint64(batch.Batch.EpochNum) == epoch.Number {
		// Batch is sticking to the current epoch, continue.
	} else if uint64(batch.Batch.EpochNum) == epoch.Number+1 {
		// With only 1 l1Block we cannot look at the next L1 Origin.
		// Note: This means that we are unable to determine validity of a batch
		// without more information. In this case we should bail out until we have
		// more information otherwise the eager algorithm may diverge from a non-eager
		// algorithm.
		if len(l1Blocks) < 2 {
			log.Info("eager batch wants to advance epoch, but could not without more L1 blocks", "current_epoch", epoch.ID())
			return BatchUndecided
		}
		batchOrigin = l1Blocks[1]
	} else {
		log.Warn("batch is for future epoch too far ahead, while it has the next timestamp, so it must be invalid", "current_epoch", epoch.ID())
		return BatchDrop
	}

	if batch.Batch.EpochHash != batchOrigin.Hash {
		log.Warn("batch is for different L1 chain, epoch hash does not match", "expected", batchOrigin.ID())
		return BatchDrop
	}

	if batch.Batch.Timestamp < batchOrigin.Time {
		log.Warn("batch timestamp is less than L1 origin timestamp", "l2_timestamp", batch.Batch.Timestamp, "l1_timestamp", batchOrigin.Time, "origin", batchOrigin.ID())
		return BatchDrop
	}

	// Check if we ran out of sequencer time drift
	if max := batchOrigin.Time + cfg.MaxSequencerDrift; batch.Batch.Timestamp > max {
		if len(batch.Batch.Transactions) == 0 {
			// If the sequencer is co-operating by producing an empty batch,
			// then allow the batch if it was the right thing to do to maintain the L2 time >= L1 time invariant.
			// We only check batches that do not advance the epoch, to ensure epoch advancement regardless of time drift is allowed.
			if epoch.Number == batchOrigin.Number {
				if len(l1Blocks) < 2 {
					log.Info("without the next L1 origin we cannot determine yet if this empty batch that exceeds the time drift is still valid")
					return BatchUndecided
				}
				nextOrigin := l1Blocks[1]
				// If Espresso is sequencing, the sequencer cannot adopt the next origin in the case
				// that HotShot failed to sequence any blocks
				if !usingEspresso && batch.Batch.Timestamp >= nextOrigin.Time { // check if the next L1 origin could have been adopted
					log.Info("batch exceeded sequencer time drift without adopting next origin, and next L1 origin would have been valid")
					return BatchDrop
				} else {
					log.Info("continuing with empty batch before late L1 block to preserve L2 time invariant")
				}
			}
		} else {
			// If the sequencer is ignoring the time drift rule, then drop the batch and force an empty batch instead,
			// as the sequencer is not allowed to include anything past this point without moving to the next epoch.
			log.Warn("batch exceeded sequencer time drift, sequencer must adopt new L1 origin to include transactions again", "max_time", max)
			return BatchDrop
		}
	}

	// We can do this check earlier, but it's a more intensive one, so we do this last.
	for i, txBytes := range batch.Batch.Transactions {
		if len(txBytes) == 0 {
			log.Warn("transaction data must not be empty, but found empty tx", "tx_index", i)
			return BatchDrop
		}
		if txBytes[0] == types.DepositTxType {
			log.Warn("sequencers may not embed any deposits into batch data, but found tx that has one", "tx_index", i)
			return BatchDrop
		}
	}
	if usingEspresso {
		return CheckBatchEspresso(cfg, log, l1Blocks, l2SafeHead, batch, hotshot)
	} else {
		return BatchAccept
	}
}

package derive

import (
	"bytes"
	"context"
	"fmt"

	"github.com/EspressoSystems/espresso-sequencer-go/nmt"
	espresso "github.com/EspressoSystems/espresso-sequencer-go/types"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type BatchWithL1InclusionBlock struct {
	L1InclusionBlock eth.L1BlockRef
	Batch            Batch
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

// Find the L1 origin which is required of an L2 block built on `parent` when running in Espresso
// mode. `suggested` is the L1 origin "suggested" by the Espresso Sequencer; namely, the L1 head
// referenced by the first Espresso block after the end of the sequencing window for this L2 block.
//
// First, `suggested` will be adjusted by the configured L1 confirmation depth, so that we will only
// use an L1 block if it has a certain number of confirmations. If the result is a valid L1 origin
// according to the rules of the derivation pipeline (e.g. it is not too old for the L2 batch, it
// did not skip an L1 block from `parent.L1Origin`, etc.) its number will be returned. Otherwise, a
// different L1 origin will be selected _deterministically_ to conform with the constraints of the
// derivation pipeline. The resulting L1 origin will always be the same as parent's or one block
// after parent's, will always conform to the derivation constraints, and is deterministic given
// `parent` and `suggested.`
func EspressoL1Origin(ctx context.Context, cfg *rollup.Config, sysCfg *eth.SystemConfig,
	parent eth.L2BlockRef, suggested uint64, l1 L1BlockRefByNumberFetcher, l log.Logger) (eth.L1BlockRef, error) {
	// The Espresso Sequencer always suggests the latest L1 block as the L1 origin. Using this
	// suggestion as-is makes us highly sensitive to L1 reorgs, since we are using a block with no
	// confirmations. `EspressoL1ConfDepth` allows the pipeline to lag behind the L1 origins
	// suggested by the Espresso Sequencer, thus always using an L1 block with at least a certain
	// number of confirmations, while the derivation remains deterministic.
	if suggested > sysCfg.EspressoL1ConfDepth {
		suggested -= sysCfg.EspressoL1ConfDepth
	} else {
		suggested = 0
	}

	prev := parent.L1Origin
	windowStart := parent.Time + cfg.BlockTime

	// Constraint 1: the L1 origin must not skip an L1 block.
	if suggested > prev.Number+1 {
		nextL1Block, err := l1.L1BlockRefByNumber(ctx, prev.Number+1)
		if err != nil {
			return eth.L1BlockRef{}, fmt.Errorf("failed to fetch next possible L1 origin %d: %w", nextL1Block, err)
		}
		nextL1BlockEligible := nextL1Block.Time <= windowStart
		// If we did skip an L1 block, that is Espresso telling us that multiple new L1 blocks have
		// already been produced. In this case, we will not block when fetching the next L1 origin,
		// so advance as far as the derivation pipeline allows: one block.
		if nextL1BlockEligible {
			l.Info("We skipped an L1 block and the next L1 block is eligible as an origin, advancing by one")
			return nextL1Block, nil
		} else {
			l.Info("We skipped an L1 block and the next L1 block is not eligible as an origin, using the old origin")
			return l1.L1BlockRefByNumber(ctx, prev.Number)
		}
	}
	// Constraint 2: the L1 origin number decreased.
	//
	// While Espresso _should_ guarantee that L1 origin numbers are monotonically increasing, a
	// limitation in the current design means that on rare occasions the L1 origin number can
	// decrease.
	if suggested < prev.Number {
		// In this case, we have no indication that new L1 blocks are ready. We don't want to
		// advance the L1 origin number and force the derivation pipeline to block waiting for a new
		// L1 block to be produced, so just reuse the previous L1 origin.
		l.Info("L1 origin decreased, using the old origin")
		return l1.L1BlockRefByNumber(ctx, prev.Number)
	}

	// Fetch information about the suggested L1 block needed to evaluate the rest of the constraints.
	l1Block, err := l1.L1BlockRefByNumber(ctx, suggested)
	if err != nil {
		return eth.L1BlockRef{}, fmt.Errorf("failed to fetch suggested L1 origin %d: %w", suggested, err)
	}

	// Constraint 3: the L1 origin is too old.
	if l1Block.Time+cfg.MaxSequencerDrift < windowStart {
		// Again, we have no explicit indication that new L1 blocks are ready, but here we are
		// forced to advance the L1 origin. At worst, the derivation pipeline may block until the
		// next L1 origin is available, but if the chosen L1 origin is this old, it is likely that a
		// new L1 block is available and Espresso just hasn't seen it yet for some reason.
		l.Info("L1 origin is too old, advancing by one",
			"suggested", l1Block, "suggested_time", l1Block.Time)
		return l1.L1BlockRefByNumber(ctx, prev.Number+1)
	}
	// Constraint 4: the L1 origin must not be newer than the L2 batch.
	if l1Block.Time > windowStart {
		// In this case `suggested` must be `prev.Number + 1`, since `prev.Number` would have a
		// timestamp earlier than `prev`, and thus earlier than the current batch. Espresso must be
		// running ahead of the L2, which is fine, we'll just wait to advance the L1 origin until
		// the L2 chain catches up.
		l.Info("L1 origin is newer than the L2 batch, use the previous origin")
		return l1.L1BlockRefByNumber(ctx, prev.Number)
	}

	// In all other cases, the suggested L1 origin is valid.
	return l1Block, nil
}

func EspressoBatchMustBeEmpty(cfg *rollup.Config, l1Origin eth.L1BlockRef, timestamp uint64) bool {
	// The constraints of the derivation pipeline require that if the L2 has fallen behind the L1
	// and is catching up, it must produce empty batches.
	return l1Origin.Time+cfg.MaxSequencerDrift < timestamp
}

func CheckBatchEspresso(cfg *rollup.Config, sysCfg *eth.SystemConfig, log log.Logger,
	l2SafeHead eth.L2BlockRef, batch *SingularBatch, l1 EspressoL1Provider) BatchValidity {
	jst := batch.Justification
	if jst == nil {
		log.Warn("dropping batch because it has no justification")
		return BatchDrop
	}

	// First, check that the headers provided by the justification match those in the sequencer
	// contract. Compute their commitments which we can compare to the sequencer contract.
	var comms []espresso.Commitment
	if jst.Prev != nil {
		comms = append(comms, jst.Prev.Commit())
	}
	for _, b := range jst.Blocks {
		comms = append(comms, b.Header.Commit())
	}
	comms = append(comms, jst.Next.Commit())
	// Compare to the authenticated commitments from the contract.
	validComms, err := l1.VerifyCommitments(jst.First().Height, comms)
	if err != nil {
		// If we can't read the expected commitments for some reason (maybe they haven't been sent
		// to the sequencer contract yet, or maybe our connection to the L1 is down) try again
		// later.
		log.Warn("error reading expected commitments", "err", err, "first", jst.First(), "count", len(comms))
		return BatchUndecided
	}
	if !validComms {
		log.Warn("dropping batch because headers do not match contract", "first", jst.First(), "count", len(comms))
		return BatchDrop
	}

	// The headers claimed by the justification are all legitimate, now check that they correctly
	// define the start and end of the time window.
	windowStart := l2SafeHead.Time + cfg.BlockTime
	windowEnd := windowStart + cfg.BlockTime
	if !checkBookends(log, windowStart, jst, WindowStart) {
		return BatchDrop
	}
	if !checkBookends(log, windowEnd, jst, WindowEnd) {
		return BatchDrop
	}

	// The Espresso data in the justification is good. Check that the L2 batch is correctly derived
	// from the Espresso blocks. First, the L1 origin:
	l1Origin, err := EspressoL1Origin(context.Background(), cfg, sysCfg, l2SafeHead,
		jst.Next.L1Head, l1, log)
	if err != nil {
		log.Warn("error finding Espresso L1 origin", "err", err, "suggested", jst.Next.L1Head)
		return BatchUndecided
	}
	if l1Origin.Number != uint64(batch.EpochNum) {
		log.Warn("dropping batch because L1 origin was not set correctly",
			"suggested", jst.Next.L1Head, "expected", l1Origin, "actual", batch.EpochNum)
		return BatchDrop
	}
	// Finally, the transactions:
	if EspressoBatchMustBeEmpty(cfg, l1Origin, batch.Timestamp) {
		if len(batch.Transactions) != 0 {
			log.Warn("dropping batch because it must be empty but isn't")
			return BatchDrop
		}
	} else {
		roots := make([]*espresso.NmtRoot, len(jst.Blocks))
		proofs := make([]*espresso.NmtProof, len(jst.Blocks))
		for i, block := range jst.Blocks {
			roots[i] = &block.Header.TransactionsRoot
			proofs[i] = &block.Proof
		}
		txs := make([]espresso.Bytes, len(batch.Transactions))
		for i, tx := range batch.Transactions {
			txs[i] = []byte(tx)
		}
		err = nmt.ValidateBatchTransactions(cfg.L2ChainID.Uint64(), roots, proofs, txs)
		if err != nil {
			log.Warn("dropping batch because of invalid NMT proofs", "err", err)
			return BatchDrop
		}
	}

	return BatchAccept
}

// CheckBatch checks if the given batch can be applied on top of the given l2SafeHead, given the contextual L1 blocks the batch was included in.
// The first entry of the l1Blocks should match the origin of the l2SafeHead. One or more consecutive l1Blocks should be provided.
// In case of only a single L1 block, the decision whether a batch is valid may have to stay undecided.
func CheckBatch(ctx context.Context, cfg *rollup.Config, sysCfg *eth.SystemConfig, log log.Logger, l1Blocks []eth.L1BlockRef,
	l2SafeHead eth.L2BlockRef, batch *BatchWithL1InclusionBlock, l1 EspressoL1Provider, l2Fetcher SafeBlockFetcher) BatchValidity {
	switch batch.Batch.GetBatchType() {
	case SingularBatchType:
		singularBatch, ok := batch.Batch.(*SingularBatch)
		if !ok {
			log.Error("failed type assertion to SingularBatch")
			return BatchDrop
		}
		return checkSingularBatch(cfg, sysCfg, log, l1Blocks, l2SafeHead, singularBatch, batch.L1InclusionBlock, l1)
	case SpanBatchType:
		spanBatch, ok := batch.Batch.(*SpanBatch)
		if !ok {
			log.Error("failed type assertion to SpanBatch")
			return BatchDrop
		}
		if !cfg.IsSpanBatch(batch.Batch.GetTimestamp()) {
			log.Warn("received SpanBatch before SpanBatch hard fork")
			return BatchDrop
		}
		return checkSpanBatch(ctx, cfg, log, l1Blocks, l2SafeHead, spanBatch, batch.L1InclusionBlock, l2Fetcher)
	default:
		log.Warn("Unrecognized batch type: %d", batch.Batch.GetBatchType())
		return BatchDrop
	}
}

// checkSingularBatch implements SingularBatch validation rule.
func checkSingularBatch(cfg *rollup.Config, sysCfg *eth.SystemConfig, log log.Logger, l1Blocks []eth.L1BlockRef, l2SafeHead eth.L2BlockRef,
	batch *SingularBatch, l1InclusionBlock eth.L1BlockRef, l1 EspressoL1Provider) BatchValidity {
	// add details to the log
	log = batch.LogContext(log)

	// sanity check we have consistent inputs
	if len(l1Blocks) == 0 {
		log.Warn("missing L1 block input, cannot proceed with batch checking")
		return BatchUndecided
	}
	epoch := l1Blocks[0]

	nextTimestamp := l2SafeHead.Time + cfg.BlockTime
	if batch.Timestamp > nextTimestamp {
		log.Trace("received out-of-order batch for future processing after next batch", "next_timestamp", nextTimestamp)
		return BatchFuture
	}
	if batch.Timestamp < nextTimestamp {
		log.Warn("dropping batch with old timestamp", "min_timestamp", nextTimestamp)
		return BatchDrop
	}

	// dependent on above timestamp check. If the timestamp is correct, then it must build on top of the safe head.
	if batch.ParentHash != l2SafeHead.Hash {
		log.Warn("ignoring batch with mismatching parent hash", "current_safe_head", l2SafeHead.Hash)
		return BatchDrop
	}

	// Filter out batches that were included too late.
	if uint64(batch.EpochNum)+cfg.SeqWindowSize < l1InclusionBlock.Number {
		log.Warn("batch was included too late, sequence window expired")
		return BatchDrop
	}

	// Check the L1 origin of the batch
	batchOrigin := epoch
	if uint64(batch.EpochNum) < epoch.Number {
		log.Warn("dropped batch, epoch is too old", "minimum", epoch.ID())
		// batch epoch too old
		return BatchDrop
	} else if uint64(batch.EpochNum) == epoch.Number {
		// Batch is sticking to the current epoch, continue.
	} else if uint64(batch.EpochNum) == epoch.Number+1 {
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

	if batch.EpochHash != batchOrigin.Hash {
		log.Warn("batch is for different L1 chain, epoch hash does not match", "expected", batchOrigin.ID())
		return BatchDrop
	}

	if batch.Timestamp < batchOrigin.Time {
		log.Warn("batch timestamp is less than L1 origin timestamp", "l2_timestamp", batch.Timestamp, "l1_timestamp", batchOrigin.Time, "origin", batchOrigin.ID())
		return BatchDrop
	}

	// Check if we ran out of sequencer time drift
	if max := batchOrigin.Time + cfg.MaxSequencerDrift; batch.Timestamp > max {
		if len(batch.Transactions) == 0 {
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
				if !sysCfg.Espresso && batch.Timestamp >= nextOrigin.Time { // check if the next L1 origin could have been adopted
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
	for i, txBytes := range batch.Transactions {
		if len(txBytes) == 0 {
			log.Warn("transaction data must not be empty, but found empty tx", "tx_index", i)
			return BatchDrop
		}
		if txBytes[0] == types.DepositTxType {
			log.Warn("sequencers may not embed any deposits into batch data, but found tx that has one", "tx_index", i)
			return BatchDrop
		}
	}
	if sysCfg.Espresso {
		return CheckBatchEspresso(cfg, sysCfg, log, l2SafeHead, batch, l1)
	} else {
		return BatchAccept
	}
}

// Check that the starting or ending bookend blocks of an Espresso block range surround the given
// starting or ending timestamp.
func checkBookends(log log.Logger, timestamp uint64, jst *eth.L2BatchJustification, endpoint windowEndpoint) bool {
	prev, next := endpoint.Bookends(jst)
	if prev == nil {
		// It is allowed that there is no Espresso block just before the endpoint only in the case
		// where the Espresso genesis block falls after the endpoint.
		if jst.First().Height != 0 {
			log.Warn("dropping batch because prev header is missing, but first block is not genesis",
				"endpoint", endpoint.String(), "first", jst.First(), "next", next, "timestamp", timestamp)
			return false
		}
		if jst.First().Timestamp < timestamp {
			log.Warn("dropping batch because prev header is missing, but genesis block is before endpoint",
				"endpoint", endpoint.String(), "first", jst.First(), "next", next, "timestamp", timestamp)
			return false
		}
	} else {
		if prev.Timestamp >= timestamp {
			log.Warn("dropping batch because prev header is from after the endpoint",
				"endpoint", endpoint.String(), "prev", prev, "timestamp", timestamp)
			return false
		}
	}
	if next.Timestamp < timestamp {
		log.Warn("dropping batch because next header is from before the endpoint",
			"endpoint", endpoint.String(), "next", next, "timestamp", timestamp)
		return false
	}

	return true
}

type windowEndpoint int

const (
	WindowStart windowEndpoint = iota
	WindowEnd
)

func (e windowEndpoint) String() string {
	return [...]string{"WindowStart", "WindowEnd"}[e]
}

func (e windowEndpoint) Bookends(jst *eth.L2BatchJustification) (prev *espresso.Header, next espresso.Header) {
	switch e {
	case WindowStart:
		// The bookend just before the start of the window is always `jst.Prev`. If it doesn't
		// exist, it's because the genesis falls in or after the window.
		prev = jst.Prev
		if len(jst.Blocks) != 0 {
			// If the window is not empty, the first block in the window defines the start of the
			// window.
			next = jst.Blocks[0].Header
		} else {
			// Otherwise, the window is empty, and the place where its starting point would be is
			// defined by the first block after the end of the window.
			next = *jst.Next
		}
	case WindowEnd:
		if len(jst.Blocks) != 0 {
			// If the window is not empty, the last block defines its end.
			prev = &jst.Blocks[len(jst.Blocks)-1].Header
		} else {
			// Otherwise, the first block before where the window would be defines the end of the
			// window. If it doesn't exist, it's because the genesis falls after the window.
			prev = jst.Prev
		}
		// The end of the window is always defined by the first block after the time range.
		next = *jst.Next
	}
	return
}

// checkSpanBatch implements SpanBatch validation rule.
func checkSpanBatch(ctx context.Context, cfg *rollup.Config, log log.Logger, l1Blocks []eth.L1BlockRef, l2SafeHead eth.L2BlockRef,
	batch *SpanBatch, l1InclusionBlock eth.L1BlockRef, l2Fetcher SafeBlockFetcher) BatchValidity {
	// add details to the log
	log = batch.LogContext(log)

	// sanity check we have consistent inputs
	if len(l1Blocks) == 0 {
		log.Warn("missing L1 block input, cannot proceed with batch checking")
		return BatchUndecided
	}
	epoch := l1Blocks[0]

	nextTimestamp := l2SafeHead.Time + cfg.BlockTime

	if batch.GetTimestamp() > nextTimestamp {
		log.Trace("received out-of-order batch for future processing after next batch", "next_timestamp", nextTimestamp)
		return BatchFuture
	}
	if batch.GetBlockTimestamp(batch.GetBlockCount()-1) < nextTimestamp {
		log.Warn("span batch has no new blocks after safe head")
		return BatchDrop
	}

	// finding parent block of the span batch.
	// if the span batch does not overlap the current safe chain, parentBLock should be l2SafeHead.
	parentNum := l2SafeHead.Number
	parentBlock := l2SafeHead
	if batch.GetTimestamp() < nextTimestamp {
		if batch.GetTimestamp() > l2SafeHead.Time {
			// batch timestamp cannot be between safe head and next timestamp
			log.Warn("batch has misaligned timestamp")
			return BatchDrop
		}
		if (l2SafeHead.Time-batch.GetTimestamp())%cfg.BlockTime != 0 {
			log.Warn("batch has misaligned timestamp")
			return BatchDrop
		}
		parentNum = l2SafeHead.Number - (l2SafeHead.Time-batch.GetTimestamp())/cfg.BlockTime - 1
		var err error
		parentBlock, err = l2Fetcher.L2BlockRefByNumber(ctx, parentNum)
		if err != nil {
			log.Error("failed to fetch L2 block", "number", parentNum, "err", err)
			// unable to validate the batch for now. retry later.
			return BatchUndecided
		}
	}
	if !batch.CheckParentHash(parentBlock.Hash) {
		log.Warn("ignoring batch with mismatching parent hash", "parent_block", parentBlock.Hash)
		return BatchDrop
	}

	startEpochNum := uint64(batch.GetStartEpochNum())

	// Filter out batches that were included too late.
	if startEpochNum+cfg.SeqWindowSize < l1InclusionBlock.Number {
		log.Warn("batch was included too late, sequence window expired")
		return BatchDrop
	}

	// Check the L1 origin of the batch
	if startEpochNum > parentBlock.L1Origin.Number+1 {
		log.Warn("batch is for future epoch too far ahead, while it has the next timestamp, so it must be invalid", "current_epoch", epoch.ID())
		return BatchDrop
	}

	endEpochNum := batch.GetBlockEpochNum(batch.GetBlockCount() - 1)
	originChecked := false
	// l1Blocks is supplied from batch queue and its length is limited to SequencerWindowSize.
	for _, l1Block := range l1Blocks {
		if l1Block.Number == endEpochNum {
			if !batch.CheckOriginHash(l1Block.Hash) {
				log.Warn("batch is for different L1 chain, epoch hash does not match", "expected", l1Block.Hash)
				return BatchDrop
			}
			originChecked = true
			break
		}
	}
	if !originChecked {
		log.Info("need more l1 blocks to check entire origins of span batch")
		return BatchUndecided
	}

	if startEpochNum < parentBlock.L1Origin.Number {
		log.Warn("dropped batch, epoch is too old", "minimum", parentBlock.ID())
		return BatchDrop
	}

	originIdx := 0
	originAdvanced := false
	if startEpochNum == parentBlock.L1Origin.Number+1 {
		originAdvanced = true
	}

	for i := 0; i < batch.GetBlockCount(); i++ {
		if batch.GetBlockTimestamp(i) <= l2SafeHead.Time {
			continue
		}
		var l1Origin eth.L1BlockRef
		for j := originIdx; j < len(l1Blocks); j++ {
			if batch.GetBlockEpochNum(i) == l1Blocks[j].Number {
				l1Origin = l1Blocks[j]
				originIdx = j
				break
			}

		}
		if i > 0 {
			originAdvanced = false
			if batch.GetBlockEpochNum(i) > batch.GetBlockEpochNum(i-1) {
				originAdvanced = true
			}
		}
		blockTimestamp := batch.GetBlockTimestamp(i)
		if blockTimestamp < l1Origin.Time {
			log.Warn("block timestamp is less than L1 origin timestamp", "l2_timestamp", blockTimestamp, "l1_timestamp", l1Origin.Time, "origin", l1Origin.ID())
			return BatchDrop
		}

		// Check if we ran out of sequencer time drift
		if max := l1Origin.Time + cfg.MaxSequencerDrift; blockTimestamp > max {
			if len(batch.GetBlockTransactions(i)) == 0 {
				// If the sequencer is co-operating by producing an empty batch,
				// then allow the batch if it was the right thing to do to maintain the L2 time >= L1 time invariant.
				// We only check batches that do not advance the epoch, to ensure epoch advancement regardless of time drift is allowed.
				if !originAdvanced {
					if originIdx+1 >= len(l1Blocks) {
						log.Info("without the next L1 origin we cannot determine yet if this empty batch that exceeds the time drift is still valid")
						return BatchUndecided
					}
					if blockTimestamp >= l1Blocks[originIdx+1].Time { // check if the next L1 origin could have been adopted
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

		for i, txBytes := range batch.GetBlockTransactions(i) {
			if len(txBytes) == 0 {
				log.Warn("transaction data must not be empty, but found empty tx", "tx_index", i)
				return BatchDrop
			}
			if txBytes[0] == types.DepositTxType {
				log.Warn("sequencers may not embed any deposits into batch data, but found tx that has one", "tx_index", i)
				return BatchDrop
			}
		}
	}

	// Check overlapped blocks
	if batch.GetTimestamp() < nextTimestamp {
		for i := uint64(0); i < l2SafeHead.Number-parentNum; i++ {
			safeBlockNum := parentNum + i + 1
			safeBlockPayload, err := l2Fetcher.PayloadByNumber(ctx, safeBlockNum)
			if err != nil {
				log.Error("failed to fetch L2 block payload", "number", parentNum, "err", err)
				// unable to validate the batch for now. retry later.
				return BatchUndecided
			}
			safeBlockTxs := safeBlockPayload.Transactions
			batchTxs := batch.GetBlockTransactions(int(i))
			// execution payload has deposit TXs, but batch does not.
			depositCount := 0
			for _, tx := range safeBlockTxs {
				if tx[0] == types.DepositTxType {
					depositCount++
				}
			}
			if len(safeBlockTxs)-depositCount != len(batchTxs) {
				log.Warn("overlapped block's tx count does not match", "safeBlockTxs", len(safeBlockTxs), "batchTxs", len(batchTxs))
				return BatchDrop
			}
			for j := 0; j < len(batchTxs); j++ {
				if !bytes.Equal(safeBlockTxs[j+depositCount], batchTxs[j]) {
					log.Warn("overlapped block's transaction does not match")
					return BatchDrop
				}
			}
			safeBlockRef, err := PayloadToBlockRef(safeBlockPayload, &cfg.Genesis)
			if err != nil {
				log.Error("failed to extract L2BlockRef from execution payload", "hash", safeBlockPayload.BlockHash, "err", err)
				return BatchDrop
			}
			if safeBlockRef.L1Origin.Number != batch.GetBlockEpochNum(int(i)) {
				log.Warn("overlapped block's L1 origin number does not match")
				return BatchDrop
			}
		}
	}

	return BatchAccept
}

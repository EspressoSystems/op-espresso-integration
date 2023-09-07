package derive

import (
	"context"

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

// Find the L1 origin which is required of an L2 block built on `parent` when running in Espresso
// mode. `suggested` is the L1 origin "suggested" by the Espresso Sequencer; namely, the L1 head
// referenced by the first Espresso block after the end of the sequencing window for this L2 block.
// If `suggested` is a valid L1 origin according to the rules of the derivation pipeline (e.g. it is
// not too old for the L2 batch, it did not skip an L1 block from `parent.L1Origin`, etc.) its
// number will be returned. Otherwise, a different L1 origin will be selected _deterministically_ to
// conform with the constraints of the derivation pipeline. The resulting L1 origin will always be
// the same as parent's or one block after parent's, will always conform to the derivation
// constraints, and is deterministic given `parent` and `suggested.`
func EspressoL1Origin(cfg *rollup.Config, parent eth.L2BlockRef, suggested eth.L1BlockRef) uint64 {
	prev := parent.L1Origin
	windowStart := parent.Time + cfg.BlockTime

	// Constraint 1: the L1 origin must not skip an L1 block.
	if suggested.Number > prev.Number+1 {
		// If we did skip an L1 block, that is Espresso telling us that multiple new L1 blocks have
		// already been produced. In this case, we will not block when fetching the next L1 origin,
		// so advance as far as the derivation pipeline allows: one block.
		return prev.Number + 1
	}
	// Constraint 2: the L1 origin number decreased.
	//
	// While Espresso _should_ guarantee that L1 origin numbers are monotonically increasing, a
	// limitation in the current design means that on rare occasions the L1 origin number can
	// decrease.
	if suggested.Number < prev.Number {
		// In this case, we have no indication that new L1 blocks are ready. We don't want to
		// advance the L1 origin number and force the derivation pipeline to block waiting for a new
		// L1 block to be produced, so just reuse the previous L1 origin.
		return prev.Number
	}
	// Constraint 3: the L1 origin is too old.
	if suggested.Time+cfg.MaxSequencerDrift < windowStart {
		// Again, we have no explicit indication that new L1 blocks are ready, but here we are
		// forced to advance the L1 origin. At worst, the derivation pipeline may block until the
		// next L1 origin is available, but if the chosen L1 origin is this old, it is likely that a
		// new L1 block is available and Espresso just hasn't seen it yet for some reason.
		return prev.Number + 1
	}
	// Constraint 4: the L1 origin must not be newer than the L2 batch.
	if suggested.Time > windowStart {
		// In this case `suggested` must be `prev.Number + 1`, since `prev.Number` would have a
		// timestamp earlier than `prev`, and thus earlier than the current batch. Espresso must be
		// running ahead of the L2, which is fine, we'll just wait to advance the L1 origin until
		// the L2 chain catches up.
		return prev.Number
	}

	// In all other cases, the suggested L1 origin is valid.
	return suggested.Number
}

func EspressoBatchMustBeEmpty(cfg *rollup.Config, l1Origin eth.L1BlockRef, timestamp uint64) bool {
	// The constraints of the derivation pipeline require that if the L2 has fallen behind the L1
	// and is catching up, it must produce empty batches.
	return l1Origin.Time+cfg.MaxSequencerDrift < timestamp
}

func CheckBatchEspresso(cfg *rollup.Config, log log.Logger, l2SafeHead eth.L2BlockRef, batch *BatchWithL1InclusionBlock, l1 EspressoL1Provider) BatchValidity {
	jst := batch.Batch.Justification
	if jst == nil {
		log.Warn("dropping batch because it has no justification")
		return BatchDrop
	}

	// First, check that the headers provided by the justification match those in the sequencer
	// contract. Compute their commitments which we can compare to the sequencer contract.
	firstComm := jst.From
	var comms []espresso.Commitment
	if jst.Prev != nil {
		firstComm -= 1
		comms = append(comms, jst.Prev.Commit())
	}
	for _, b := range jst.Blocks {
		comms = append(comms, b.Header.Commit())
	}
	comms = append(comms, jst.Next.Commit())
	// Compare to the authenticated commitments from the contract.
	validComms, err := l1.VerifyCommitments(firstComm, comms)
	if err != nil {
		// If we can't read the expected commitments for some reason (maybe they haven't been sent
		// to the sequencer contract yet, or maybe our connection to the L1 is down) try again
		// later.
		log.Warn("error reading expected commitments", "err", err, "first", firstComm, "count", len(comms))
		return BatchUndecided
	}
	if !validComms {
		log.Warn("dropping batch because headers do not match contract", "first", firstComm, "count", len(comms))
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
	suggestedL1Origin, err := l1.L1BlockRefByNumber(context.Background(), jst.Next.L1Head)
	if err != nil {
		// If we can't read the suggested L1 origin for some reason (maybe our L1 client is lagging
		// behind Espresso's view of the L1) try again later.
		log.Warn("error reading suggested L1 origin", "err", err, "l1 head", jst.Next.L1Head)
		return BatchUndecided
	}
	expectedL1Origin := EspressoL1Origin(cfg, l2SafeHead, suggestedL1Origin)
	actualL1Origin := uint64(batch.Batch.EpochNum)
	if expectedL1Origin != actualL1Origin {
		log.Warn("dropping batch because L1 origin was not set correctly",
			"suggested", jst.Next.L1Head, "expected", expectedL1Origin, "actual", actualL1Origin)
		return BatchDrop
	}
	// Fetch details for the actual L1 origin.
	var l1Origin eth.L1BlockRef
	if actualL1Origin == suggestedL1Origin.Number {
		l1Origin = suggestedL1Origin
	} else {
		l1Origin, err = l1.L1BlockRefByNumber(context.Background(), actualL1Origin)
		if err != nil {
			log.Warn("error reading actual L1 origin", "err", err, "origin", actualL1Origin)
			return BatchUndecided
		}
	}
	// Finally, the transactions:
	if EspressoBatchMustBeEmpty(cfg, l1Origin, batch.Batch.Timestamp) {
		if len(batch.Batch.Transactions) != 0 {
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
		txs := make([]espresso.Bytes, len(batch.Batch.Transactions))
		for i, tx := range batch.Batch.Transactions {
			txs[i] = []byte(tx)
		}
		err = espresso.ValidateBatchTransactions(cfg.L2ChainID.Uint64(), roots, proofs, txs)
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
func CheckBatch(cfg *rollup.Config, log log.Logger, l1Blocks []eth.L1BlockRef, l2SafeHead eth.L2BlockRef, batch *BatchWithL1InclusionBlock, usingEspresso bool, l1 EspressoL1Provider) BatchValidity {
	// add details to the log
	log = log.New(
		"batch_timestamp", batch.Batch.Timestamp,
		"parent_number", l2SafeHead.Number,
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
					log.Warn("batch exceeded sequencer time drift without adopting next origin, and next L1 origin would have been valid")
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
		return CheckBatchEspresso(cfg, log, l2SafeHead, batch, l1)
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
		if jst.From != 0 || next.Timestamp < timestamp {
			log.Warn("dropping batch because prev header is missing, but genesis is not after endpoint",
				"endpoint", endpoint.String(), "from", jst.From, "next", next, "timestamp", timestamp)
			return false
		}
	} else {
		if prev.Timestamp >= timestamp {
			log.Warn("dropping batch because prev header is from after the endpoint",
				"endpoint", endpoint.String(), "prev", prev, "timestamp", timestamp)
			return false
		}
		if next.Timestamp < timestamp {
			log.Warn("dropping batch because next header is from before the endpoint",
				"endpoint", endpoint.String(), "next", next, "timestamp", timestamp)
			return false
		}
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

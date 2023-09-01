package driver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/espresso"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type SequencerMode uint64

const (
	Espresso SequencerMode = iota
	Legacy
	Unknown
)

type Downloader interface {
	InfoByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, error)
	FetchReceipts(ctx context.Context, blockHash common.Hash) (eth.BlockInfo, types.Receipts, error)
}

type L1OriginSelectorIface interface {
	FindL1Origin(ctx context.Context, l2Head eth.L2BlockRef) (eth.L1BlockRef, error)
	FindL1OriginByNumber(ctx context.Context, number uint64) (eth.L1BlockRef, error)
}

type SequencerMetrics interface {
	RecordSequencerInconsistentL1Origin(from eth.BlockID, to eth.BlockID)
	RecordSequencerReset()
}

type InProgressBatch struct {
	onto         eth.L2BlockRef
	jst          eth.L2BatchJustification
	transactions []eth.Data
	windowStart  uint64
	windowEnd    uint64
}

func (b *InProgressBatch) complete() bool {
	// A batch is complete when we have one block past the end of its time window, proving that we
	// have not omitted any blocks that should have fallen within the window.
	return b.jst.Next != nil
}

// Sequencer implements the sequencing interface of the driver: it starts and completes block building jobs.
type Sequencer struct {
	log    log.Logger
	config *rollup.Config
	mode   SequencerMode

	engine derive.ResettableEngineControl

	attrBuilder      derive.AttributesBuilder
	l1OriginSelector L1OriginSelectorIface
	espresso         espresso.QueryService

	metrics SequencerMetrics

	// timeNow enables sequencer testing to mock the time
	timeNow func() time.Time

	nextAction time.Time

	// The current Espresso block we are building, if applicable.
	espressoBatch *InProgressBatch
}

func NewSequencer(log log.Logger, cfg *rollup.Config, engine derive.ResettableEngineControl, attributesBuilder derive.AttributesBuilder, l1OriginSelector L1OriginSelectorIface, espresso espresso.QueryService, metrics SequencerMetrics) *Sequencer {
	return &Sequencer{
		log:              log,
		config:           cfg,
		mode:             Unknown,
		engine:           engine,
		timeNow:          time.Now,
		attrBuilder:      attributesBuilder,
		l1OriginSelector: l1OriginSelector,
		espresso:         espresso,
		metrics:          metrics,
		espressoBatch:    nil,
	}
}

// startBuildingEspressoBatch initiates an Espresso block building job on top of the given L2 head,
// safe and finalized blocks. After this function succeeds, `d.espressoBatch` is guaranteed to be
// non-nil.
func (d *Sequencer) startBuildingEspressoBatch(ctx context.Context, l2Head eth.L2BlockRef) error {
	windowStart := l2Head.Time + d.config.BlockTime
	windowEnd := windowStart + d.config.BlockTime

	// Fetch the available HotShot blocks from this sequencing window.
	blocks, err := d.espresso.FetchHeadersForWindow(ctx, windowStart, windowEnd)
	if err != nil {
		return err
	}

	d.espressoBatch = &InProgressBatch{
		onto:        l2Head,
		windowStart: windowStart,
		windowEnd:   windowEnd,
		jst: eth.L2BatchJustification{
			From: blocks.From,
			Prev: blocks.Prev,
		},
	}
	return d.updateEspressoBatch(ctx, blocks.Window, blocks.Next)
}

// updateEspressoBatch appends the transactions contained in the Espresso blocks denoted by
// `newHeaders` to the current in-progress batch. If `end`, the first block after the window of this
// batch, is available, it will be saved in the `Next` field of the batch justification.
func (d *Sequencer) updateEspressoBatch(ctx context.Context, newHeaders []espresso.Header, end *espresso.Header) error {
	batch := d.espressoBatch
	for _, header := range newHeaders {
		blocks := batch.jst.Blocks
		numBlocks := len(blocks)

		// Validate that the given header is in the window and in the right order.
		if header.Timestamp >= batch.windowEnd {
			return derive.NewCriticalError(fmt.Errorf("inconsistent data from Espresso query service: header %v in window has timestamp after window end %d", header, batch.windowEnd))
		}
		if header.Timestamp < batch.windowStart {
			// Eventually, we should return an error here. However due to a limitation in the
			// current implementation of HotShot/Espresso, block timestamps will sometimes decrease.
			d.log.Error("inconsistent data from Espresso query service: header is before window start", "header", header, "start", batch.windowStart)
		}
		prev := batch.jst.Prev
		if numBlocks != 0 {
			prev = &blocks[numBlocks-1].Header
		}
		if prev != nil && header.Timestamp < prev.Timestamp {
			// Similarly, this should eventually be an error, but can happen with the current
			// version of Espresso.
			d.log.Error("inconsistent data from Espresso query service: header is before its predecessor", "header", header, "prev", prev)
		}

		blockNum := batch.jst.From + uint64(numBlocks)
		txs, err := d.espresso.FetchTransactionsInBlock(ctx, blockNum, &header, d.config.L2ChainID.Uint64())
		if err != nil {
			return err
		}
		d.log.Info("adding new transactions from Espresso", "block", blockNum, "count", len(txs.Transactions))
		batch.jst.Blocks = append(blocks, eth.EspressoBlockJustification{
			Header: header,
			Proof:  &txs.Proof,
		})
		for _, tx := range txs.Transactions {
			batch.transactions = append(batch.transactions, []byte(tx))
		}
	}

	batch.jst.Next = end
	return nil
}

// tryToSealEspressoBatch polls for new transactions from the Espresso Sequencer to append to the
// current Espresso Block. If the resulting block is complete (Espresso has sequenced at least one
// block with a timestamp beyond the end of the current sequencing window) it will submit the block
// to the engine and return the resulting execution payload. If the block cannot be sealed yet
// because Espresso hasn't sequenced enough blocks, returns nil.
func (d *Sequencer) tryToSealEspressoBatch(ctx context.Context) (*eth.ExecutionPayload, error) {
	batch := d.espressoBatch
	if !batch.complete() {
		blocks, err := d.espresso.FetchRemainingHeadersForWindow(ctx, batch.jst.From+uint64(len(batch.jst.Blocks)), batch.windowEnd)
		if err != nil {
			return nil, err
		}
		if err := d.updateEspressoBatch(ctx, blocks.Window, blocks.Next); err != nil {
			return nil, err
		}
	}
	if batch.complete() {
		return d.sealEspressoBatch(ctx)
	} else {
		return nil, nil
	}
}

// sealEspressoBatch submits the current Espresso batch to the engine and return the resulting
// execution payload.
func (d *Sequencer) sealEspressoBatch(ctx context.Context) (*eth.ExecutionPayload, error) {
	batch := d.espressoBatch

	// Deterministically choose an L1 origin for this L2 batch, based on the latest L1 block that
	// Espresso has told us exists, but adjusting as needed to meet the constraints of the
	// derivation pipeline.
	suggestedL1Origin, err := d.l1OriginSelector.FindL1OriginByNumber(ctx, batch.jst.Next.L1Head)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch suggested L1 origin %d: %w", batch.jst.Next.L1Head, err)
	}
	l1OriginNumber := derive.EspressoL1Origin(d.config, batch.onto, suggestedL1Origin)
	l1Origin := suggestedL1Origin
	if l1Origin.Number != l1OriginNumber {
		l1Origin, err = d.l1OriginSelector.FindL1OriginByNumber(ctx, l1OriginNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch L1 origin %d: %w", l1OriginNumber, err)
		}
		d.log.Info("using adjusted L1 origin",
			"suggested", suggestedL1Origin, "adjusted", l1Origin, "parent", batch.onto, "parentOrigin", batch.onto.L1Origin)
	}

	// In certain edge cases, like when the L2 has fallen too far behind the L1, we are required to
	// submit empty batches until we catch up.
	if derive.EspressoBatchMustBeEmpty(d.config, l1Origin, batch.windowStart) {
		batch.transactions = nil

		// We don't need all the NMT proofs in this case, so save space in the batch by deleting
		// them.
		for i := range batch.jst.Blocks {
			batch.jst.Blocks[i].Proof = nil
		}
	}

	attrs, err := d.attrBuilder.PreparePayloadAttributes(ctx, batch.onto, l1Origin.ID(), &batch.jst)
	if err != nil {
		return nil, err
	}
	attrs.NoTxPool = true
	attrs.Transactions = append(attrs.Transactions, batch.transactions...)

	d.log.Debug("prepared attributes for new Espresso block",
		"num", batch.onto.Number+1, "time", uint64(attrs.Timestamp), "origin", l1Origin)

	// Start a payload building process.
	errTyp, err := d.engine.StartPayload(ctx, batch.onto, attrs, false)
	if err != nil {
		return nil, fmt.Errorf("failed to start building on top of L2 chain %s, error (%d): %w", batch.onto, errTyp, err)
	}
	// Immediately seal the block in the engine.
	payload, errTyp, err := d.engine.ConfirmPayload(ctx)
	if err != nil {
		_ = d.engine.CancelPayload(ctx, true)
		return nil, fmt.Errorf("failed to complete building block: error (%d): %w", errTyp, err)
	}
	d.espressoBatch = nil
	return payload, nil
}

// startBuildingLegacyBlock initiates a legacy block building job on top of the given L2 head, safe and finalized blocks, and using the provided l1Origin.
func (d *Sequencer) startBuildingLegacyBlock(ctx context.Context) error {
	l2Head := d.engine.UnsafeL2Head()

	// Figure out which L1 origin block we're going to be building on top of.
	l1Origin, err := d.l1OriginSelector.FindL1Origin(ctx, l2Head)
	if err != nil {
		d.log.Error("error finding next L1 Origin", "err", err)
		return err
	}

	if !(l2Head.L1Origin.Hash == l1Origin.ParentHash || l2Head.L1Origin.Hash == l1Origin.Hash) {
		d.metrics.RecordSequencerInconsistentL1Origin(l2Head.L1Origin, l1Origin.ID())
		return derive.NewResetError(fmt.Errorf("cannot build new L2 block with L1 origin %s (parent L1 %s) on current L2 head %s with L1 origin %s", l1Origin, l1Origin.ParentHash, l2Head, l2Head.L1Origin))
	}

	d.log.Info("creating new block", "parent", l2Head, "l1Origin", l1Origin)

	fetchCtx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	attrs, err := d.attrBuilder.PreparePayloadAttributes(fetchCtx, l2Head, l1Origin.ID(), nil)
	if err != nil {
		return err
	}

	// If our next L2 block timestamp is beyond the Sequencer drift threshold, then we must produce
	// empty blocks (other than the L1 info deposit and any user deposits). We handle this by
	// setting NoTxPool to true, which will cause the Sequencer to not include any transactions
	// from the transaction pool.
	attrs.NoTxPool = uint64(attrs.Timestamp) > l1Origin.Time+d.config.MaxSequencerDrift

	d.log.Debug("prepared attributes for new block",
		"num", l2Head.Number+1, "time", uint64(attrs.Timestamp),
		"origin", l1Origin, "origin_time", l1Origin.Time, "noTxPool", attrs.NoTxPool)

	// Start a payload building process.
	errTyp, err := d.engine.StartPayload(ctx, l2Head, attrs, false)
	if err != nil {
		return fmt.Errorf("failed to start building on top of L2 chain %s, error (%d): %w", l2Head, errTyp, err)
	}
	return nil
}

// completeBuildingLegacyBlock takes the current legacy block that is being built, and asks the engine to complete the building, seal the block, and persist it as canonical.
// Warning: the safe and finalized L2 blocks as viewed during the initiation of the block building are reused for completion of the block building.
// The Execution engine should not change the safe and finalized blocks between start and completion of block building.
func (d *Sequencer) completeBuildingLegacyBlock(ctx context.Context) (*eth.ExecutionPayload, error) {
	payload, errTyp, err := d.engine.ConfirmPayload(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to complete building block: error (%d): %w", errTyp, err)
	}
	return payload, nil
}

// CancelBuildingBlock cancels the current open block building job.
// This sequencer only maintains one block building job at a time.
func (d *Sequencer) cancelBuildingLegacyBlock(ctx context.Context) {
	// force-cancel, we can always continue block building, and any error is logged by the engine state
	_ = d.engine.CancelPayload(ctx, true)
}

// PlanNextSequencerAction returns a desired delay till the RunNextSequencerAction call.
func (d *Sequencer) PlanNextSequencerAction() time.Duration {
	// Regardless of what mode we are in (Espresso or Legacy) our first priority is to not bother
	// the engine if it is busy building safe blocks (and thus changing the head that we would sync
	// on top of). Give it time to sync up.
	if onto, _, safe := d.engine.BuildingPayload(); safe {
		d.log.Warn("delaying sequencing to not interrupt safe-head changes", "onto", onto, "onto_time", onto.Time)
		// approximates the worst-case time it takes to build a block, to reattempt sequencing after.
		return time.Second * time.Duration(d.config.BlockTime)
	}

	switch d.mode {
	case Espresso:
		return d.planNextEspressoSequencerAction()
	case Legacy:
		return d.planNextLegacySequencerAction()
	default:
		// If we don't yet know what mode we are in, our first action is going to be discovering our
		// mode based on the L2 system config. We should start this immediately, since it will
		// impact our scheduling decisions for all future actions.
		return 0
	}
}

func (d *Sequencer) planNextEspressoSequencerAction() time.Duration {
	head := d.engine.UnsafeL2Head()
	now := d.timeNow()

	// We may have to wait till the next sequencing action, e.g. upon an error.
	// However, we ignore this delay if we are building a block and the L2 head has changed, in
	// which case we need to respond immediately.
	delay := d.nextAction.Sub(now)
	reorg := d.espressoBatch != nil && d.espressoBatch.onto.Hash != head.Hash
	if delay > 0 && !reorg {
		return delay
	}

	// In case there has been a reorg or the previous action did not set a delay, run the next
	// action immediately.
	return 0
}

func (d *Sequencer) planNextLegacySequencerAction() time.Duration {
	head := d.engine.UnsafeL2Head()
	now := d.timeNow()

	buildingOnto, buildingID, _ := d.engine.BuildingPayload()

	// We may have to wait till the next sequencing action, e.g. upon an error.
	// If the head changed we need to respond and will not delay the sequencing.
	if delay := d.nextAction.Sub(now); delay > 0 && buildingOnto.Hash == head.Hash {
		return delay
	}

	blockTime := time.Duration(d.config.BlockTime) * time.Second
	payloadTime := time.Unix(int64(head.Time+d.config.BlockTime), 0)
	remainingTime := payloadTime.Sub(now)

	// If we started building a block already, and if that work is still consistent,
	// then we would like to finish it by sealing the block.
	if buildingID != (eth.PayloadID{}) && buildingOnto.Hash == head.Hash {
		// if we started building already, then we will schedule the sealing.
		if remainingTime < sealingDuration {
			return 0 // if there's not enough time for sealing, don't wait.
		} else {
			// finish with margin of sealing duration before payloadTime
			return remainingTime - sealingDuration
		}
	} else {
		// if we did not yet start building, then we will schedule the start.
		if remainingTime > blockTime {
			// if we have too much time, then wait before starting the build
			return remainingTime - blockTime
		} else {
			// otherwise start instantly
			return 0
		}
	}
}

// BuildingOnto returns the L2 head reference that the latest block is or was being built on top of.
func (d *Sequencer) BuildingOnto() eth.L2BlockRef {
	if d.espressoBatch != nil {
		return d.espressoBatch.onto
	} else {
		ref, _, _ := d.engine.BuildingPayload()
		return ref
	}
}

func (d *Sequencer) StartBuildingBlock(ctx context.Context) error {
	switch d.mode {
	case Espresso:
		return d.startBuildingEspressoBatch(ctx, d.engine.UnsafeL2Head())
	case Legacy:
		return d.startBuildingLegacyBlock(ctx)
	default:
		// Detect mode, then try again.
		if err := d.detectMode(ctx); err != nil {
			return err
		}
		// If that succeeded, `d.mode` is now either Espresso or Legacy.
		return d.StartBuildingBlock(ctx)
	}
}

func (d *Sequencer) CompleteBuildingBlock(ctx context.Context) (*eth.ExecutionPayload, error) {
	switch d.mode {
	case Espresso:
		return d.tryToSealEspressoBatch(ctx)
	case Legacy:
		return d.completeBuildingLegacyBlock(ctx)
	default:
		return nil, fmt.Errorf("not building a block")
	}
}

// RunNextSequencerAction starts new block building work, or seals existing work,
// and is best timed by first awaiting the delay returned by PlanNextSequencerAction.
// If a new block is successfully sealed, it will be returned for publishing, nil otherwise.
//
// Only critical errors are bubbled up, other errors are handled internally.
// Internally starting or sealing of a block may fail with a derivation-like error:
//   - If it is a critical error, the error is bubbled up to the caller.
//   - If it is a reset error, the ResettableEngineControl used to build blocks is requested to reset, and a backoff applies.
//     No attempt is made at completing the block building.
//   - If it is a temporary error, a backoff is applied to reattempt building later.
//   - If it is any other error, a backoff is applied and building is cancelled.
//
// Upon L1 reorgs that are deep enough to affect the L1 origin selection, a reset-error may occur,
// to direct the engine to follow the new L1 chain before continuing to sequence blocks.
// It is up to the EngineControl implementation to handle conflicting build jobs of the derivation
// process (as verifier) and sequencing process.
// Generally it is expected that the latest call interrupts any ongoing work,
// and the derivation process does not interrupt in the happy case,
// since it can consolidate previously sequenced blocks by comparing sequenced inputs with derived inputs.
// If the derivation pipeline does force a conflicting block, then an ongoing sequencer task might still finish,
// but the derivation can continue to reset until the chain is correct.
// If the engine is currently building safe blocks, then that building is not interrupted, and sequencing is delayed.
func (d *Sequencer) RunNextSequencerAction(ctx context.Context) (*eth.ExecutionPayload, error) {
	// Regardless of what mode we are in (Espresso or Legacy) our first priority is to not bother
	// the engine if it is busy building safe blocks (and thus changing the head that we would sync
	// on top of). Give it time to sync up.
	onto, buildingID, safe := d.engine.BuildingPayload()
	if buildingID != (eth.PayloadID{}) && safe {
		d.log.Warn("avoiding sequencing to not interrupt safe-head changes", "onto", onto, "onto_time", onto.Time)
		// approximates the worst-case time it takes to build a block, to reattempt sequencing after.
		d.nextAction = d.timeNow().Add(time.Second * time.Duration(d.config.BlockTime))
		return nil, nil
	}

	switch d.mode {
	case Espresso:
		return d.buildEspressoBatch(ctx)
	case Legacy:
		return d.buildLegacyBlock(ctx, buildingID != eth.PayloadID{})
	default:
		// If we don't know what mode we are in, figure it out and then schedule another action
		// immediately.
		if err := d.detectMode(ctx); err != nil {
			return nil, d.handleNonEngineError("to determine mode", err)
		}
		// Now that we know what mode we're in, return to the scheduler to plan the next action.
		return nil, nil
	}
}

func (d *Sequencer) buildEspressoBatch(ctx context.Context) (*eth.ExecutionPayload, error) {
	// First, check if there has been a reorg. If so, drop the current block and restart.
	head := d.engine.UnsafeL2Head()
	if d.espressoBatch != nil && d.espressoBatch.onto.Hash != head.Hash {
		d.log.Warn("reorg detected", "head", head, "onto", d.espressoBatch.onto)
		d.espressoBatch = nil
	}

	// Begin a new block if necessary.
	if d.espressoBatch == nil {
		d.log.Info("building new Espresso batch", "onto", head)
		if err := d.startBuildingEspressoBatch(ctx, head); err != nil {
			return nil, d.handleNonEngineError("starting Espresso block", err)
		}
	}

	// Poll for transactions from the Espresso Sequencer and see if we can submit the block.
	block, err := d.tryToSealEspressoBatch(ctx)
	if err != nil {
		return nil, d.handlePossibleEngineError("trying to seal Espresso block", err)
	}
	if block == nil {
		// If we didn't seal the block, it means we reached the end of the Espresso block stream.
		// Wait a reasonable amount of time before checking for more transactions.
		d.log.Debug("Espresso batch was not ready to seal, will retry in 1 second")
		d.nextAction = d.timeNow().Add(time.Second)
		return nil, nil
	} else {
		// If we did seal the block, return it and do not set a delay, so that the scheduler will
		// start the next action (starting the next block) immediately.
		d.log.Info("sealed Espresso batch", "payload", block)
		return block, nil
	}
}

func (d *Sequencer) buildLegacyBlock(ctx context.Context, building bool) (*eth.ExecutionPayload, error) {
	if building {
		payload, err := d.completeBuildingLegacyBlock(ctx)
		if err != nil {
			if errors.Is(err, derive.ErrCritical) {
				return nil, err // bubble up critical errors.
			} else if errors.Is(err, derive.ErrReset) {
				d.log.Error("sequencer failed to seal new block, requiring derivation reset", "err", err)
				d.metrics.RecordSequencerReset()
				d.nextAction = d.timeNow().Add(time.Second * time.Duration(d.config.BlockTime)) // hold off from sequencing for a full block
				d.cancelBuildingLegacyBlock(ctx)
				d.engine.Reset()
			} else if errors.Is(err, derive.ErrTemporary) {
				d.log.Error("sequencer failed temporarily to seal new block", "err", err)
				d.nextAction = d.timeNow().Add(time.Second)
				// We don't explicitly cancel block building jobs upon temporary errors: we may still finish the block.
				// Any unfinished block building work eventually times out, and will be cleaned up that way.
			} else {
				d.log.Error("sequencer failed to seal block with unclassified error", "err", err)
				d.nextAction = d.timeNow().Add(time.Second)
				d.cancelBuildingLegacyBlock(ctx)
			}
			return nil, nil
		} else {
			d.log.Info("sequencer successfully built a new block", "block", payload.ID(), "time", uint64(payload.Timestamp), "txs", len(payload.Transactions))
			return payload, nil
		}
	} else {
		err := d.startBuildingLegacyBlock(ctx)
		if err != nil {
			return nil, d.handlePossibleEngineError("to start building new block", err)
		} else {
			parent, buildingID, _ := d.engine.BuildingPayload() // we should have a new payload ID now that we're building a block
			d.log.Info("sequencer started building new block", "payload_id", buildingID, "l2_parent_block", parent, "l2_parent_block_time", parent.Time)
			return nil, nil
		}
	}
}

func (d *Sequencer) detectMode(ctx context.Context) error {
	head := d.engine.UnsafeL2Head()
	espressoBatch, err := d.attrBuilder.ChildNeedsJustification(ctx, head)
	if err != nil {
		return err
	}
	if espressoBatch {
		d.log.Info("OP sequencer running in Espresso mode")
		d.mode = Espresso
	} else {
		d.log.Info("OP sequencer running in legacy mode")
		d.mode = Legacy
	}
	return nil
}

func (d *Sequencer) handlePossibleEngineError(action string, err error) error {
	if errors.Is(err, derive.ErrCritical) {
		return err
	} else if errors.Is(err, derive.ErrReset) {
		d.log.Error("sequencer failed, requiring derivation reset", "action", action, "err", err)
		d.metrics.RecordSequencerReset()
		d.nextAction = d.timeNow().Add(time.Second * time.Duration(d.config.BlockTime)) // hold off from sequencing for a full block
		d.engine.Reset()
		return nil
	} else {
		return d.handleNonEngineError(action, err)
	}
}

func (d *Sequencer) handleNonEngineError(action string, err error) error {
	if errors.Is(err, derive.ErrCritical) {
		return err
	} else if errors.Is(err, derive.ErrTemporary) {
		d.log.Error("sequencer encountered temporary error", "action", action, "err", err)
		d.nextAction = d.timeNow().Add(time.Second)
	} else {
		d.log.Error("sequencer encountered unclassified error", "action", action, "err", err)
		d.nextAction = d.timeNow().Add(time.Second)
	}
	return nil
}

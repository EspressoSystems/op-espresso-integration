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
	onto     eth.L2BlockRef
	l1Origin eth.L1BlockRef
	jst      eth.L2BatchJustification
	blocks   [][]espresso.Bytes

	windowStart uint64
	windowEnd   uint64
}

func (b *InProgressBatch) complete() bool {
	// A batch with a nil `Payload` justification is always complete: it means that the batch is
	// not eligible to include Espresso transactions, and should be sealed empty.
	return b.jst.Payload == nil || b.jst.Payload.NextBatchFirstBlock.Timestamp >= b.windowEnd
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

	// Fetch the available HotShot blocks from this sequencing window. The first block in the window
	// tells us what L1 origin we're going to be building for.
	blocks, err := d.espresso.FetchHeadersForWindow(ctx, windowStart, windowEnd)
	if err != nil {
		return err
	}

	batch := &InProgressBatch{
		onto:        l2Head,
		windowStart: windowStart,
		windowEnd:   windowEnd,
	}

	// Usually, `blocks` includes one block before the start of the L2 batch window. However, a
	// special case is when the Espresso chain starts inside or after the window. In this case,
	// `PrevBatchLastBlock` is meaningless.
	if blocks.Prev != nil {
		batch.jst.PrevBatchLastBlock = *blocks.Prev
	} else {
		// Sanity check that we are in the Espresso genesis case.
		if blocks.From != 0 {
			return derive.NewCriticalError(fmt.Errorf("inconsistent data from Espresso query service: Prev is nil but From is not 0 (%d)", blocks.From))
		}
	}

	// Find the first block in the window, or just past it if the window is empty.
	if len(blocks.Window) > 0 {
		batch.jst.FirstBlock = blocks.Window[0]
	} else if blocks.Next != nil {
		batch.jst.FirstBlock = *blocks.Next
	} else {
		// Neither the first block in the window nor the end of the window is currently available,
		// so we can't even determine metadata like the L1 origin to use for the batch. Return a
		// temporary error so we try again shortly.
		return derive.NewTemporaryError(fmt.Errorf("not enough blocks available to determine L1 origin of next L2 batch"))
	}
	batch.jst.FirstBlockNumber = blocks.From

	// Before fetching the L1 origin determined by `jst.FirstBlock`, check for cases where Espresso
	// did not provide an eligible L1 origin.
	// 1) Espresso did not produce any blocks in the window
	if batch.jst.FirstBlock.Timestamp >= windowEnd {
		// Produce an empty batch but keep the same L1 origin as the previous block. This origin may
		// be old, but this is allowed since the batch is empty. We don't want to advance the L1
		// origin because the next L1 block may not be available yet, which would force the
		// derivation pipeline to block.
		l1OriginNumber := l2Head.L1Origin.Number
		batch.l1Origin, err = d.l1OriginSelector.FindL1OriginByNumber(ctx, l1OriginNumber)
		if err != nil {
			d.log.Error("error finding L1 origin", "number", l1OriginNumber, "err", err)
			return err
		}
		d.espressoBatch = batch
		d.log.Info("building empty Espresso batch because Espresso produced no blocks",
			"windowStart", windowStart, "windowEnd", windowEnd,
			"PrevBatchLastBlock", batch.jst.PrevBatchLastBlock, "firstBlock", batch.jst.FirstBlock)
		return nil
	}
	// 2) Espresso skipped an L1 block.
	if batch.jst.FirstBlock.L1Head > l2Head.L1Origin.Number+1 {
		// Produce an empty batch that advances the L1 origin by 1, so we can catch up to Espresso.
		l1OriginNumber := l2Head.L1Origin.Number + 1
		batch.l1Origin, err = d.l1OriginSelector.FindL1OriginByNumber(ctx, l1OriginNumber)
		if err != nil {
			d.log.Error("error finding L1 origin", "number", l1OriginNumber, "err", err)
			return err
		}
		d.espressoBatch = batch
		d.log.Info("building empty Espresso batch because Espresso skipped an L1 block",
			"firstBlockL1", batch.jst.FirstBlock.L1Head, "prevL1", l2Head.L1Origin.Number)
		return nil
	}

	// Fetch the L1 origin determined by the first Espresso block.
	l1OriginNumber := batch.jst.FirstBlock.L1Head
	// While Espresso _should_ guarantee that L1 origin numbers are monotonically increasing, a
	// limitation in the current design means that on rare occasions the L1 origin number can
	// decrease. As a temporary work-around, we detect this case and handle it as if Espresso had
	// simply chosen the same L1 origin as the previous block.
	if l1OriginNumber < l2Head.L1Origin.Number {
		l1OriginNumber = l2Head.L1Origin.Number
	}
	batch.l1Origin, err = d.l1OriginSelector.FindL1OriginByNumber(ctx, l1OriginNumber)
	if err != nil {
		d.log.Error("error finding L1 origin", "number", l1OriginNumber, "err", err)
		return err
	}

	// Check for one more case where the L1 origin is ineligible: if it is too old, we produce an
	// empty batch that advances the L1 origin by 1.
	if batch.l1Origin.Time+d.config.MaxSequencerDrift < windowStart {
		oldL1Origin := batch.l1Origin
		l1OriginNumber = l2Head.L1Origin.Number + 1
		batch.l1Origin, err = d.l1OriginSelector.FindL1OriginByNumber(ctx, l1OriginNumber)
		if err != nil {
			d.log.Error("error finding L1 origin", "number", l1OriginNumber, "err", err)
			return err
		}
		d.espressoBatch = batch
		d.log.Info("building empty Espresso batch because L1 origin is too far behind",
			"windowStart", windowStart, "MaxSequencerDrift", d.config.MaxSequencerDrift,
			"suggestedOrigin", oldL1Origin)
		return nil
	}

	// If we didn't hit any of the edge cases above, we are eligible to include transactions
	// produced by Espresso in this batch.
	batch.jst.Payload = &eth.L2BatchPayloadJustification{
		// We have not included any blocks yet, so the "last block" (the block to start fetching new
		// blocks after) is one _before_ the start of the window. `updateEspressoBatch` will update
		// this field as the blocks we have fetched get inserted into the batch.
		LastBlock: batch.jst.PrevBatchLastBlock,
		// `NextBatchFirstBlock` must always be one block after `LastBlock`. `updateEspressoBatch`
		// will also keep this field in sync.
		NextBatchFirstBlock: batch.jst.FirstBlock,
		// We haven't added any proofs yet.
		NmtProofs: nil,
	}
	d.espressoBatch = batch
	return d.updateEspressoBatch(ctx, blocks.Window, blocks.Next)
}

// updateEspressoBatch appends the transactions contained in the Espresso blocks denoted by
// `newHeaders` to the current in-progress batch. If `end`, the first block after the window of this
// batch, is available, it will be saved in the `NextBatchFirstBlock` field of the batch
// justification.
func (d *Sequencer) updateEspressoBatch(ctx context.Context, newHeaders []espresso.Header, end *espresso.Header) error {
	batch := d.espressoBatch
	for i := range newHeaders {
		header := &newHeaders[i]

		// Validate that the given header is in the window and in the right order.
		if header.Timestamp >= batch.windowEnd {
			return derive.NewCriticalError(fmt.Errorf("inconsistent data from Espresso query service: header %v in window has timestamp after window end %d", header, batch.windowEnd))
		}
		if header.Timestamp < batch.windowStart {
			// Eventually, we should return an error here. However due to a limitation in the
			// current implementation of HotShot/Espresso, block timestamps will sometimes decrease.
			d.log.Error("inconsistent data from Espresso query service: header is before window start", "header", header, "start", batch.windowStart)
		}
		if header.Timestamp < batch.jst.Payload.LastBlock.Timestamp {
			// Similarly, this should eventually be an error, but can happen with the current
			// version of Espresso.
			d.log.Error("inconsistent data from Espresso query service: header is before its predecessor", "header", header, "prev", batch.jst.Payload.LastBlock)
		}

		blockNum := batch.jst.FirstBlockNumber + uint64(len(batch.blocks))
		txs, err := d.espresso.FetchTransactionsInBlock(ctx, blockNum, header, d.config.L2ChainID.Uint64())
		if err != nil {
			return err
		}
		d.log.Info("adding new transactions from Espresso", "block", blockNum, "count", len(txs.Transactions))
		batch.jst.Payload.NmtProofs = append(batch.jst.Payload.NmtProofs, txs.Proof)
		batch.blocks = append(batch.blocks, txs.Transactions)
		batch.jst.Payload.LastBlock = *header
	}

	if end != nil {
		batch.jst.Payload.NextBatchFirstBlock = *end
	}

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
		blocks, err := d.espresso.FetchRemainingHeadersForWindow(ctx, batch.jst.FirstBlockNumber+uint64(len(batch.blocks)), batch.windowEnd)
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
	attrs, err := d.attrBuilder.PreparePayloadAttributes(ctx, batch.onto, batch.l1Origin.ID(), &batch.jst)
	if err != nil {
		return nil, err
	}
	attrs.NoTxPool = true
	for i := range batch.blocks {
		block := batch.blocks[i]
		for j := range block {
			txn := block[j]
			attrs.Transactions = append(attrs.Transactions, []byte(txn))
		}
	}

	d.log.Debug("prepared attributes for new Espresso block",
		"num", batch.onto.Number+1, "time", uint64(attrs.Timestamp),
		"origin", batch.l1Origin)

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

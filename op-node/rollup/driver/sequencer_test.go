package driver

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-node/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/espresso"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
)

var mockResetErr = fmt.Errorf("mock reset err: %w", derive.ErrReset)

type FakeEngineControl struct {
	finalized eth.L2BlockRef
	safe      eth.L2BlockRef
	unsafe    eth.L2BlockRef

	buildingOnto eth.L2BlockRef
	buildingID   eth.PayloadID
	buildingSafe bool

	buildingAttrs *eth.PayloadAttributes
	buildingStart time.Time

	cfg *rollup.Config

	timeNow func() time.Time

	makePayload func(onto eth.L2BlockRef, attrs *eth.PayloadAttributes) *eth.ExecutionPayload

	errTyp derive.BlockInsertionErrType
	err    error

	totalBuildingTime time.Duration
	totalBuiltBlocks  int
	totalTxs          int

	l2Batches []*eth.ExecutionPayload
}

func (m *FakeEngineControl) avgBuildingTime() time.Duration {
	return m.totalBuildingTime / time.Duration(m.totalBuiltBlocks)
}

func (m *FakeEngineControl) avgTxsPerBlock() float64 {
	return float64(m.totalTxs) / float64(m.totalBuiltBlocks)
}

func (m *FakeEngineControl) StartPayload(ctx context.Context, parent eth.L2BlockRef, attrs *eth.PayloadAttributes, updateSafe bool) (errType derive.BlockInsertionErrType, err error) {
	if m.err != nil {
		return m.errTyp, m.err
	}
	m.buildingID = eth.PayloadID{}
	_, _ = crand.Read(m.buildingID[:])
	m.buildingOnto = parent
	m.buildingSafe = updateSafe
	m.buildingAttrs = attrs
	m.buildingStart = m.timeNow()
	return derive.BlockInsertOK, nil
}

func (m *FakeEngineControl) ConfirmPayload(ctx context.Context) (out *eth.ExecutionPayload, errTyp derive.BlockInsertionErrType, err error) {
	if m.err != nil {
		return nil, m.errTyp, m.err
	}
	buildTime := m.timeNow().Sub(m.buildingStart)
	m.totalBuildingTime += buildTime
	m.totalBuiltBlocks += 1
	payload := m.makePayload(m.buildingOnto, m.buildingAttrs)
	ref, err := derive.PayloadToBlockRef(payload, &m.cfg.Genesis)
	if err != nil {
		panic(err)
	}
	m.unsafe = ref
	if m.buildingSafe {
		m.safe = ref
	}

	m.resetBuildingState()
	m.totalTxs += len(payload.Transactions)
	m.l2Batches = append(m.l2Batches, payload)
	return payload, derive.BlockInsertOK, nil
}

func (m *FakeEngineControl) CancelPayload(ctx context.Context, force bool) error {
	if force {
		m.resetBuildingState()
	}
	return m.err
}

func (m *FakeEngineControl) BuildingPayload() (onto eth.L2BlockRef, id eth.PayloadID, safe bool) {
	return m.buildingOnto, m.buildingID, m.buildingSafe
}

func (m *FakeEngineControl) Finalized() eth.L2BlockRef {
	return m.finalized
}

func (m *FakeEngineControl) UnsafeL2Head() eth.L2BlockRef {
	return m.unsafe
}

func (m *FakeEngineControl) SafeL2Head() eth.L2BlockRef {
	return m.safe
}

func (m *FakeEngineControl) resetBuildingState() {
	m.buildingID = eth.PayloadID{}
	m.buildingOnto = eth.L2BlockRef{}
	m.buildingSafe = false
	m.buildingAttrs = nil
}

func (m *FakeEngineControl) Reset() {
	m.err = nil
}

var _ derive.ResettableEngineControl = (*FakeEngineControl)(nil)

type FakeEspressoClient struct {
	Blocks          []FakeEspressoBlock
	AdvanceL1Origin bool
}

type FakeEspressoBlock struct {
	Header       espresso.Header
	Transactions []espresso.Bytes
}

type TestSequencer struct {
	t   *testing.T
	rng *rand.Rand

	cfg        rollup.Config
	seq        *Sequencer
	engControl FakeEngineControl
	espresso   *FakeEspressoClient

	clockTime time.Time
	clockFn   func() time.Time
	l1Times   map[eth.BlockID]uint64

	attrsErr    error
	originErr   error
	espressoErr error
}

// Implement AttributeBuilder interface for TestSequencer.

// We keep attribute building simple, we don't talk to a real execution engine in this test.
// Sometimes we fake an error in the attributes preparation.
func (s *TestSequencer) PreparePayloadAttributes(ctx context.Context, l2Parent eth.L2BlockRef, epoch eth.BlockID, justification *eth.L2BatchJustification) (attrs *eth.PayloadAttributes, err error) {
	if s.attrsErr != nil {
		return nil, s.attrsErr
	}
	seqNr := l2Parent.SequenceNumber + 1
	if epoch != l2Parent.L1Origin {
		seqNr = 0
	}
	l1Info := &testutils.MockBlockInfo{
		InfoHash:        epoch.Hash,
		InfoParentHash:  mockL1Hash(epoch.Number - 1),
		InfoCoinbase:    common.Address{},
		InfoRoot:        common.Hash{},
		InfoNum:         epoch.Number,
		InfoTime:        s.l1Times[epoch],
		InfoMixDigest:   [32]byte{},
		InfoBaseFee:     big.NewInt(1234),
		InfoReceiptRoot: common.Hash{},
	}
	infoDep, err := derive.L1InfoDepositBytes(seqNr, l1Info, s.cfg.Genesis.SystemConfig, justification, false)
	require.NoError(s.t, err)

	testGasLimit := eth.Uint64Quantity(10_000_000)
	return &eth.PayloadAttributes{
		Timestamp:             eth.Uint64Quantity(l2Parent.Time + s.cfg.BlockTime),
		PrevRandao:            eth.Bytes32{},
		SuggestedFeeRecipient: common.Address{},
		Transactions:          []eth.Data{infoDep},
		NoTxPool:              false,
		GasLimit:              &testGasLimit,
	}, nil
}

// The system config never changes in this test, so we just read whether Espresso is enabled or not from the genesis config.
// Sometimes we fake an error.
func (s *TestSequencer) ChildNeedsJustification(ctx context.Context, parent eth.L2BlockRef) (bool, error) {
	if s.attrsErr != nil {
		return false, s.attrsErr
	}
	return s.cfg.Genesis.SystemConfig.Espresso, nil
}

var _ derive.AttributesBuilder = (*TestSequencer)(nil)

// Implement L1OriginSelector interface for TestSequencer.

// The origin selector just generates random L1 blocks based on RNG
func (s *TestSequencer) FindL1Origin(ctx context.Context, l2Head eth.L2BlockRef) (eth.L1BlockRef, error) {
	if s.originErr != nil {
		return eth.L1BlockRef{}, s.originErr
	}

	origin := eth.L1BlockRef{
		Hash:       mockL1Hash(l2Head.L1Origin.Number),
		Number:     l2Head.L1Origin.Number,
		ParentHash: mockL1Hash(l2Head.L1Origin.Number),
		Time:       s.l1Times[l2Head.L1Origin],
	}
	// randomly make a L1 origin appear, if we can even select it
	nextL2Time := l2Head.Time + s.cfg.BlockTime
	if nextL2Time <= origin.Time {
		return origin, nil
	}
	if s.rng.Intn(10) == 0 {
		return s.nextOrigin(origin, l2Head.Time, false), nil
	} else {
		return origin, nil
	}
}

func (s *TestSequencer) FindL1OriginByNumber(ctx context.Context, number uint64) (eth.L1BlockRef, error) {
	if s.originErr != nil {
		return eth.L1BlockRef{}, s.originErr
	}
	return s.l1BlockByNumber(number), nil
}

// Infallible version of `FindL1OriginByNumber`. This is used internally by other mock functions
// which do their own error injection, to avoid doubling up mock errors.
func (s *TestSequencer) l1BlockByNumber(number uint64) eth.L1BlockRef {
	if number <= s.cfg.Genesis.L1.Number {
		return eth.L1BlockRef{
			Hash:       s.cfg.Genesis.L1.Hash,
			Number:     s.cfg.Genesis.L1.Number,
			ParentHash: s.cfg.Genesis.L1.Hash,
			Time:       s.l1Times[s.cfg.Genesis.L1],
		}
	}

	parent := s.l1BlockByNumber(number - 1)
	id := eth.BlockID{
		Number: number,
		Hash:   mockL1Hash(number),
	}
	if s.l1Times[id] == 0 {
		return s.nextOrigin(parent, s.engControl.UnsafeL2Head().Time, true)
	} else {
		return eth.L1BlockRef{
			Hash:       id.Hash,
			Number:     id.Number,
			ParentHash: parent.Hash,
			Time:       s.l1Times[id],
		}
	}
}

func (s *TestSequencer) nextOrigin(prevOrigin eth.L1BlockRef, prevL2Time uint64, allowFuture bool) eth.L1BlockRef {
	// Find the first L2 block time which is greater than the previous origin time.
	nextL2Time := prevL2Time + s.cfg.BlockTime
	if nextL2Time <= prevOrigin.Time {
		// Get the amount of time we need to increment by to reach `prevOrigin.Time`, rounded up to
		// a multiple of the L2 block time
		delta := ((prevOrigin.Time-nextL2Time)/s.cfg.BlockTime + 1) * s.cfg.BlockTime
		nextL2Time += delta
	}

	maxL1BlockTimeGap := uint64(100)
	maxTimeIncrement := nextL2Time - prevOrigin.Time
	if maxTimeIncrement > maxL1BlockTimeGap {
		maxTimeIncrement = maxL1BlockTimeGap
	}
	nextOrigin := eth.L1BlockRef{
		Hash:       mockL1Hash(prevOrigin.Number + 1),
		Number:     prevOrigin.Number + 1,
		ParentHash: prevOrigin.Hash,
		Time:       prevOrigin.Time + 1 + uint64(s.rng.Int63n(int64(maxTimeIncrement))),
	}
	if allowFuture && s.rng.Int63n(20) == 0 {
		nextOrigin.Time = nextL2Time + 1
		s.t.Logf("using L1 origin in the future prevOrigin=%v nextOrigin=%v prevL2Time=%d nextL2Time=%d", prevOrigin, nextOrigin, prevL2Time, nextL2Time)
	}
	s.l1Times[nextOrigin.ID()] = nextOrigin.Time
	return nextOrigin
}

var _ L1OriginSelectorIface = (*TestSequencer)(nil)

// Implement EspressoL1Provider interface for TestSequencer.

func (s *TestSequencer) L1BlockRefByNumber(ctx context.Context, number uint64) (eth.L1BlockRef, error) {
	return s.FindL1OriginByNumber(ctx, number)
}

func (s *TestSequencer) FetchReceipts(ctx context.Context, blockHash common.Hash) (eth.BlockInfo, types.Receipts, error) {
	return nil, nil, fmt.Errorf("not implemented: FetchReceipts")
}

func (s *TestSequencer) VerifyCommitments(firstBlockHeight uint64, comms []espresso.Commitment) (bool, error) {
	for i, comm := range comms {
		if !comm.Equals(s.espressoBlock(firstBlockHeight + uint64(i)).Commit()) {
			return false, nil
		}
	}
	return true, nil
}

var _ derive.EspressoL1Provider = (*TestSequencer)(nil)

// Implement Espresso QueryService interface for TestSequencer.

func (s *TestSequencer) FetchHeadersForWindow(ctx context.Context, start uint64, end uint64) (espresso.WindowStart, error) {
	// Find the start of the range.
	for i := uint64(0); ; i += 1 {
		header := s.espressoBlock(i)
		if header == nil {
			// New headers not available.
			return espresso.WindowStart{}, nil
		}
		if header.Timestamp >= start {
			res, err := s.FetchRemainingHeadersForWindow(ctx, i, end)
			if err != nil {
				return espresso.WindowStart{}, err
			} else {
				var prev *espresso.Header
				if i > 0 {
					prev = s.espressoBlock(i - 1)
				}
				return espresso.WindowStart{
					From:   i,
					Window: res.Window,
					Prev:   prev,
					Next:   res.Next,
				}, nil
			}
		}
	}
}

func (s *TestSequencer) FetchRemainingHeadersForWindow(ctx context.Context, from uint64, end uint64) (espresso.WindowMore, error) {
	// Inject errors.
	if s.espressoErr != nil {
		return espresso.WindowMore{}, s.espressoErr
	}

	headers := make([]espresso.Header, 0)
	for i := from; ; i += 1 {
		header := s.espressoBlock(i)
		if header == nil {
			// New headers not available.
			return espresso.WindowMore{
				Window: headers,
				Next:   nil,
			}, nil
		}
		if header.Timestamp >= end {
			return espresso.WindowMore{
				Window: headers,
				Next:   header,
			}, nil
		}
		headers = append(headers, *header)
	}
}

func (s *TestSequencer) FetchTransactionsInBlock(ctx context.Context, block uint64, header *espresso.Header, namespace uint64) (espresso.TransactionsInBlock, error) {
	// Inject errors.
	if s.espressoErr != nil {
		return espresso.TransactionsInBlock{}, s.espressoErr
	}

	// The sequencer should only ever ask for one namespace, that of the OP-chain.
	require.Equal(s.t, namespace, s.cfg.L2ChainID.Uint64())

	if int(block) >= len(s.espresso.Blocks) {
		return espresso.TransactionsInBlock{}, fmt.Errorf("invalid block number %d total blocks %d", block, len(s.espresso.Blocks))
	}
	if s.espresso.Blocks[block].Header.Commit() != header.Commit() {
		return espresso.TransactionsInBlock{}, fmt.Errorf("wrong header for block %d header %v expected %v", block, header, s.espresso.Blocks[block].Header)
	}
	txs := s.espresso.Blocks[block].Transactions

	// Fake an NMT proof.
	proof := espresso.NmtProof{}
	return espresso.TransactionsInBlock{
		Transactions: txs,
		Proof:        proof,
	}, nil
}

func (s *TestSequencer) espressoBlock(i uint64) *espresso.Header {
	// Insert blocks as necessary.
	for uint64(len(s.espresso.Blocks)) <= i {
		if s.nextEspressoBlock() == nil {
			return nil
		}
	}
	return &s.espresso.Blocks[i].Header
}

func (s *TestSequencer) nextEspressoBlock() *espresso.Header {
	var prev espresso.Header
	if len(s.espresso.Blocks) > 0 {
		prev = s.espresso.Blocks[len(s.espresso.Blocks)-1].Header
	} else {
		// Set a timestamp for the genesis that is near the L2 genesis.
		prev.Timestamp = s.cfg.Genesis.L2Time
	}

	// Advance the timestamp by a random amount between 0 and 150% of the L2 block time. This
	// should lead to some L2 batches having multiple Espresso blocks, some having only 1, and some
	// being empty entirely.
	timestamp := prev.Timestamp + uint64(s.rng.Intn(int(s.cfg.BlockTime+s.cfg.BlockTime/2+1)))
	// Don't produce blocks in the future.
	now := uint64(s.clockTime.Unix())
	if timestamp > now {
		return nil
	}
	// Don't produce blocks too far in the past, as this can cause the L2 timestamp to drift from
	// the wall clock time in a way that wouldn't really happen in real life: real Espresso blocks
	// will have a timestamp that is pretty close to wall clock time.
	if now-timestamp > 6 {
		// If we have drifted too far, catch up all at once, so we can start slowly drifting again
		// rather than constantly drifting a little and catching up a little.
		timestamp = now
	}

	// Fake an NMT root, but ensure it is unique.
	root := espresso.NmtRoot{
		Root: make([]byte, 8),
	}
	binary.LittleEndian.PutUint64(root.Root, uint64(len(s.espresso.Blocks)))

	var l1OriginNumber uint64
	if s.espresso.AdvanceL1Origin {
		l1OriginNumber = prev.L1Head + 1
		s.espresso.AdvanceL1Origin = false
	} else {
		l1OriginNumber = prev.L1Head
		switch s.rng.Intn(20) {
		case 0:
			// 5%: move the L1 origin _backwards_. Espresso is supposed to enforce that the L1
			// origin is monotonically increasing, but due to limitations in the current version of
			// the HotShot interfaces, the current version does not, and the L1 block number will,
			// rarely, decrease.
			if l1OriginNumber > 0 {
				// Correct the mistake in the next block. Otherwise, since the probability of
				// advancing the L1 origin is fairly low (in order to simulate the case where the L1
				// origin is old), once we decrease the L1 origin once, it can remain behind for
				// many blocks.
				s.espresso.AdvanceL1Origin = true
				l1OriginNumber -= 1
			}
		case 1, 2:
			// 10%: advance L1 origin
			l1OriginNumber += 1
		case 3:
			// 5%: skip ahead to the latest possible L1 origin
			for {
				l1Origin := s.l1BlockByNumber(l1OriginNumber + 1)
				if l1Origin.Time >= timestamp {
					break
				}
				l1OriginNumber += 1
			}
		default:
			// 80%: use old L1 origin
			break
		}
	}

	l1Origin := s.l1BlockByNumber(l1OriginNumber)

	// 5% of the time, mess with the timestamp. Again, Espresso should ensure that the timestamps
	// are monotonically increasing, but for now, it doesn't.
	if prev.Timestamp > 0 && s.rng.Intn(20) == 0 {
		timestamp = prev.Timestamp - 1
	}

	header := espresso.Header{
		TransactionsRoot: root,
		Metadata: espresso.Metadata{
			Timestamp: timestamp,
			L1Head:    l1Origin.Number,
		},
	}

	// Randomly generate between 0 and 20 transactions.
	txs := make([]espresso.Bytes, 0)
	for i := 0; i < s.rng.Intn(20); i++ {
		txs = append(txs, []byte(fmt.Sprintf("mock sequenced tx %d", i)))
	}

	s.espresso.Blocks = append(s.espresso.Blocks, FakeEspressoBlock{
		Header:       header,
		Transactions: txs,
	})
	return &header
}

var _ espresso.QueryService = (*TestSequencer)(nil)

func mockL1Hash(num uint64) (out common.Hash) {
	out[31] = 1
	binary.BigEndian.PutUint64(out[:], num)
	return
}

func mockL2Hash(num uint64) (out common.Hash) {
	out[31] = 2
	binary.BigEndian.PutUint64(out[:], num)
	return
}

func mockL1ID(num uint64) eth.BlockID {
	return eth.BlockID{Hash: mockL1Hash(num), Number: num}
}

func mockL2ID(num uint64) eth.BlockID {
	return eth.BlockID{Hash: mockL2Hash(num), Number: num}
}

func SetupSequencer(t *testing.T, useEspresso bool) *TestSequencer {
	s := new(TestSequencer)
	s.t = t
	s.rng = rand.New(rand.NewSource(12345))

	l1Time := uint64(100000)

	// mute errors. We expect a lot of the mocked errors to cause error-logs. We check chain health at the end of the test.
	log := testlog.Logger(t, log.LvlCrit)

	s.cfg = rollup.Config{
		Genesis: rollup.Genesis{
			L1:     mockL1ID(100000),
			L2:     mockL2ID(200000),
			L2Time: l1Time + 300, // L2 may start with a relative old L1 origin and will have to catch it up
			SystemConfig: eth.SystemConfig{
				Espresso: useEspresso,
			},
		},
		L1ChainID:         big.NewInt(900),
		L2ChainID:         big.NewInt(901),
		BlockTime:         2,
		MaxSequencerDrift: 30,
	}
	// keep track of the L1 timestamps we mock because sometimes we only have the L1 hash/num handy
	s.l1Times = map[eth.BlockID]uint64{s.cfg.Genesis.L1: l1Time}

	genesisL2 := eth.L2BlockRef{
		Hash:           s.cfg.Genesis.L2.Hash,
		Number:         s.cfg.Genesis.L2.Number,
		ParentHash:     mockL2Hash(s.cfg.Genesis.L2.Number - 1),
		Time:           s.cfg.Genesis.L2Time,
		L1Origin:       s.cfg.Genesis.L1,
		SequenceNumber: 0,
	}
	// initialize our engine state
	s.engControl = FakeEngineControl{
		finalized: genesisL2,
		safe:      genesisL2,
		unsafe:    genesisL2,
		cfg:       &s.cfg,
	}

	// start wallclock at 5 minutes after the current L2 head. The sequencer has some catching up to do!
	s.clockTime = time.Unix(int64(s.engControl.unsafe.Time)+5*60, 0)
	s.clockFn = func() time.Time {
		return s.clockTime
	}
	s.engControl.timeNow = s.clockFn

	// mock payload building, we don't need to process any real txs.
	s.engControl.makePayload = func(onto eth.L2BlockRef, attrs *eth.PayloadAttributes) *eth.ExecutionPayload {
		txs := make([]eth.Data, 0)
		txs = append(txs, attrs.Transactions...) // include deposits
		if !attrs.NoTxPool {                     // if we are allowed to sequence from tx pool, mock some txs
			n := s.rng.Intn(20)
			for i := 0; i < n; i++ {
				txs = append(txs, []byte(fmt.Sprintf("mock sequenced tx %d", i)))
			}
		}
		return &eth.ExecutionPayload{
			ParentHash:   onto.Hash,
			BlockNumber:  eth.Uint64Quantity(onto.Number) + 1,
			Timestamp:    attrs.Timestamp,
			BlockHash:    mockL2Hash(onto.Number),
			Transactions: txs,
		}
	}

	// Set up a fake Espresso client if necessary.
	if useEspresso {
		s.espresso = new(FakeEspressoClient)
	}

	s.seq = NewSequencer(log, &s.cfg, &s.engControl, s, s, s, metrics.NoopMetrics)
	s.seq.timeNow = s.clockFn

	return s
}

// SequencerChaosMonkey runs the sequencer in a mocked adversarial environment with
// repeated random errors in dependencies and poor clock timing.
// At the end the health of the chain is checked to show that the sequencer kept the chain in shape.
func SequencerChaosMonkey(s *TestSequencer) {
	t := s.t

	// try to build 1000 blocks, with 5x as many planning attempts, to handle errors and clock problems
	desiredBlocks := 1000
	for i := 0; i < 5*desiredBlocks && s.engControl.totalBuiltBlocks < desiredBlocks; i++ {
		delta := s.seq.PlanNextSequencerAction()

		x := s.rng.Float32()
		if x < 0.01 { // 1%: mess a lot with the clock: simulate a hang of up to 30 seconds
			if i < desiredBlocks/2 { // only in first 50% of blocks to let it heal, hangs take time
				delta = time.Duration(s.rng.Float64() * float64(time.Second*30))
			}
		} else if x < 0.1 { // 9%: mess with the timing, -50% to 50% off
			delta = time.Duration((0.5 + s.rng.Float64()) * float64(delta))
		} else if x < 0.5 {
			// 40%: mess slightly with the timing, -10% to 10% off
			delta = time.Duration((0.9 + s.rng.Float64()*0.2) * float64(delta))
		}
		s.clockTime = s.clockTime.Add(delta)

		// reset errors
		s.originErr = nil
		s.attrsErr = nil
		s.espressoErr = nil
		if s.engControl.err != mockResetErr { // the mockResetErr requires the sequencer to Reset() to recover.
			s.engControl.err = nil
		}
		s.engControl.errTyp = derive.BlockInsertOK

		// maybe make something maybe fail, or try a new L1 origin
		switch s.rng.Intn(20) { // 9/20 = 45% chance to fail sequencer action (!!!)
		case 0, 1:
			s.originErr = errors.New("mock origin error")
		case 2, 3:
			s.attrsErr = errors.New("mock attributes error")
		case 4, 5:
			s.engControl.err = errors.New("mock temporary engine error")
			s.engControl.errTyp = derive.BlockInsertTemporaryErr
		case 6, 7:
			s.engControl.err = errors.New("mock prestate engine error")
			s.engControl.errTyp = derive.BlockInsertPrestateErr
		case 8:
			s.engControl.err = mockResetErr
		case 9:
			s.espressoErr = errors.New("mock espresso client error")
		default:
			// no error
		}
		payload, err := s.seq.RunNextSequencerAction(context.Background())
		require.NoError(t, err)
		if payload != nil {
			require.Equal(t, s.engControl.UnsafeL2Head().ID(), payload.ID(), "head must stay in sync with emitted payloads")
			var tx types.Transaction
			require.NoError(t, tx.UnmarshalBinary(payload.Transactions[0]))
			info, err := derive.L1InfoDepositTxData(tx.Data())
			require.NoError(t, err)
			require.GreaterOrEqual(t, uint64(payload.Timestamp), info.Time, "ensure L2 time >= L1 time")
		}
	}

	// Now, even though:
	// - the start state was behind the wallclock
	// - the L1 origin was far behind the L2
	// - we made all components fail at random
	// - messed with the clock
	// the L2 chain was still built and stats are healthy on average!
	l2Head := s.engControl.UnsafeL2Head()
	t.Logf("avg build time: %s, clock timestamp: %d, L2 head time: %d, L1 origin time: %d, avg txs per block: %f", s.engControl.avgBuildingTime(), s.clockFn().Unix(), l2Head.Time, s.l1Times[l2Head.L1Origin], s.engControl.avgTxsPerBlock())
	require.Equal(t, s.engControl.totalBuiltBlocks, desiredBlocks, "persist through random errors and build the desired blocks")
	require.Equal(t, l2Head.Time, s.cfg.Genesis.L2Time+uint64(desiredBlocks)*s.cfg.BlockTime, "reached desired L2 block timestamp")
	require.GreaterOrEqual(t, l2Head.Time, s.l1Times[l2Head.L1Origin], "the L2 time >= the L1 time")
	require.Less(t, l2Head.Time-s.l1Times[l2Head.L1Origin], uint64(100), "The L1 origin time is close to the L2 time")
	require.Greater(t, s.engControl.avgTxsPerBlock(), 3.0, "We expect at least 1 system tx per block, but with a mocked 0-10 txs we expect an higher avg")
}

func TestSequencerChaosMonkeyLegacy(t *testing.T) {
	s := SetupSequencer(t, false)
	SequencerChaosMonkey(s)

	l2Head := s.engControl.UnsafeL2Head()
	require.Less(t, s.clockTime.Sub(time.Unix(int64(l2Head.Time), 0)).Abs(), 2*time.Second, "L2 time is accurate, within 2 seconds of wallclock")
	require.Greater(t, s.engControl.avgBuildingTime(), time.Second, "With 2 second block time and 1 second error backoff and healthy-on-average errors, building time should at least be a second")
}

func TestSequencerChaosMonkeyEspresso(t *testing.T) {
	s := SetupSequencer(t, true)
	SequencerChaosMonkey(s)

	// Check that the L2 block time is accurate. The tolerance here is slightly higher than for the
	// legacy sequencer, since the Espresso mode sequencer has to wait for one additional HotShot
	// block to be sequenced after the time window for an L2 batch ends, before it can sequence that
	// batch. This problem is exacerbated with the chaos monkey, since it may take even more wall
	// clock time to fetch that last Espresso block due to injected errors.
	l2Head := s.engControl.UnsafeL2Head()
	require.Less(t, s.clockTime.Sub(time.Unix(int64(l2Head.Time), 0)).Abs(), 12*time.Second, "L2 time is accurate, within 12 seconds of wallclock")
	// Here, the legacy test checks `avgBuildingTime()`. This stat is meaningless for the Espresso
	// mode sequencer, since it builds blocks locally rather than in the engine.

	// After running the chaos monkey, check that the sequenced blocks satisfy the constraints of
	// the derivation pipeline. Count how many times we hit each interesting case.
	prevL1Origin := s.cfg.Genesis.L1.Number
	happyPath := 0
	noEspressoBlocks := 0
	oldL1Origin := 0
	newL1Origin := 0
	skippedL1Origin := 0
	decreasingL1Origin := 0
	l2Head = eth.L2BlockRef{
		Hash:           s.cfg.Genesis.L2.Hash,
		Number:         s.cfg.Genesis.L2.Number,
		ParentHash:     mockL2Hash(s.cfg.Genesis.L2.Number - 1),
		Time:           s.cfg.Genesis.L2Time,
		L1Origin:       s.cfg.Genesis.L1,
		SequenceNumber: 0,
	}
	for _, payload := range s.engControl.l2Batches {
		// Find the number of deposit transactions in the L2 block, or, equivalently, the offset of
		// the first transaction produced by Espresso.
		numDepositTxs := 0
		for tx := range payload.Transactions {
			if payload.Transactions[tx][0] == uint8(types.DepositTxType) {
				numDepositTxs += 1
			} else {
				break
			}
		}

		// Parse the L1 info from the payload.
		var tx types.Transaction
		require.NoError(t, tx.UnmarshalBinary(payload.Transactions[0]))
		l1Info, err := derive.L1InfoDepositTxData(tx.Data())
		require.NoError(t, err)
		jst := l1Info.Justification

		// Reconstruct the batch from the execution payload.
		batch := derive.NewSingularBatchData(derive.SingularBatch{
			Justification: jst,
			ParentHash:    payload.ParentHash,
			Timestamp:     uint64(payload.Timestamp),
			EpochNum:      rollup.Epoch(l1Info.Number),
			EpochHash:     l1Info.BlockHash,
			Transactions:  payload.Transactions[numDepositTxs:],
		})
		batchWithL1 := derive.BatchWithL1InclusionBlock{
			L1InclusionBlock: eth.L1BlockRef{},
			Batch:            batch,
		}

		// Check that the derivation pipeline would accept this batch.
		status := derive.CheckBatchEspresso(&s.cfg, testlog.Logger(t, log.LvlInfo), l2Head, &batchWithL1, s)
		require.Equal(t, status, derive.BatchValidity(derive.BatchAccept), "sequencer built a block that the derivation pipeline will not accept")

		// Figure out which interesting cases we hit.
		suggestedL1Origin := s.l1BlockByNumber(jst.Next.L1Head)
		if suggestedL1Origin.Number > prevL1Origin+1 {
			skippedL1Origin++
		} else if suggestedL1Origin.Number < prevL1Origin {
			decreasingL1Origin++
		} else if suggestedL1Origin.Time+s.cfg.MaxSequencerDrift < batch.Timestamp {
			oldL1Origin++
		} else if suggestedL1Origin.Time > batch.Timestamp {
			newL1Origin++
		} else if len(jst.Blocks) == 0 {
			noEspressoBlocks++
		} else {
			happyPath++
		}

		// Move to the next batch.
		l2Head, err = derive.PayloadToBlockRef(payload, &s.cfg.Genesis)
		require.Nil(t, err, "failed to convert payload to block ref")
		prevL1Origin = l1Info.Number
	}

	t.Logf("Espresso sequencing case coverage:")
	t.Logf("Happy path:           %d", happyPath)
	t.Logf("No Espresso blocks:   %d", noEspressoBlocks)
	t.Logf("Old L1 origin:        %d", oldL1Origin)
	t.Logf("New L1 origin:        %d", newL1Origin)
	t.Logf("Skipped L1 origin:    %d", skippedL1Origin)
	t.Logf("Decreasing L1 origin: %d", decreasingL1Origin)
}

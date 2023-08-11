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
	"github.com/ethereum-optimism/optimism/op-node/testlog"
	"github.com/ethereum-optimism/optimism/op-node/testutils"
	"github.com/ethereum-optimism/optimism/op-service/eth"
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

type testAttrBuilderFn func(ctx context.Context, l2Parent eth.L2BlockRef, epoch eth.BlockID, justification *eth.L2BatchJustification) (attrs *eth.PayloadAttributes, err error)

func (fn testAttrBuilderFn) PreparePayloadAttributes(ctx context.Context, l2Parent eth.L2BlockRef, epoch eth.BlockID, justification *eth.L2BatchJustification) (attrs *eth.PayloadAttributes, err error) {
	return fn(ctx, l2Parent, epoch, justification)
}

var _ derive.AttributesBuilder = (testAttrBuilderFn)(nil)

type testOriginSelectorFn func(ctx context.Context, l2Head eth.L2BlockRef) (eth.L1BlockRef, error)

func (fn testOriginSelectorFn) FindL1Origin(ctx context.Context, l2Head eth.L2BlockRef) (eth.L1BlockRef, error) {
	return fn(ctx, l2Head)
}

var _ L1OriginSelectorIface = (testOriginSelectorFn)(nil)

type FakeEspressoClient struct{}

type TestSequencer struct {
	rng *rand.Rand

	cfg        rollup.Config
	seq        *Sequencer
	engControl FakeEngineControl
	espresso   *FakeEspressoClient

	clockTime time.Time
	clockFn   func() time.Time
	l1Times   map[eth.BlockID]uint64

	attrsErr  error
	originErr error
}

func SetupSequencer(t *testing.T, useEspresso bool) *TestSequencer {
	s := new(TestSequencer)

	mockL1Hash := func(num uint64) (out common.Hash) {
		out[31] = 1
		binary.BigEndian.PutUint64(out[:], num)
		return
	}
	mockL2Hash := func(num uint64) (out common.Hash) {
		out[31] = 2
		binary.BigEndian.PutUint64(out[:], num)
		return
	}
	mockL1ID := func(num uint64) eth.BlockID {
		return eth.BlockID{Hash: mockL1Hash(num), Number: num}
	}
	mockL2ID := func(num uint64) eth.BlockID {
		return eth.BlockID{Hash: mockL2Hash(num), Number: num}
	}

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

	// We keep attribute building simple, we don't talk to a real execution engine in this test.
	// Sometimes we fake an error in the attributes preparation.
	attrBuilder := testAttrBuilderFn(func(ctx context.Context, l2Parent eth.L2BlockRef, epoch eth.BlockID, justification *eth.L2BatchJustification) (attrs *eth.PayloadAttributes, err error) {
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
		require.NoError(t, err)

		testGasLimit := eth.Uint64Quantity(10_000_000)
		return &eth.PayloadAttributes{
			Timestamp:             eth.Uint64Quantity(l2Parent.Time + s.cfg.BlockTime),
			PrevRandao:            eth.Bytes32{},
			SuggestedFeeRecipient: common.Address{},
			Transactions:          []eth.Data{infoDep},
			NoTxPool:              false,
			GasLimit:              &testGasLimit,
		}, nil
	})

	maxL1BlockTimeGap := uint64(100)
	// The origin selector just generates random L1 blocks based on RNG
	originSelector := testOriginSelectorFn(func(ctx context.Context, l2Head eth.L2BlockRef) (eth.L1BlockRef, error) {
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
		maxTimeIncrement := nextL2Time - origin.Time
		if maxTimeIncrement > maxL1BlockTimeGap {
			maxTimeIncrement = maxL1BlockTimeGap
		}
		if s.rng.Intn(10) == 0 {
			nextOrigin := eth.L1BlockRef{
				Hash:       mockL1Hash(origin.Number + 1),
				Number:     origin.Number + 1,
				ParentHash: origin.Hash,
				Time:       origin.Time + 1 + uint64(s.rng.Int63n(int64(maxTimeIncrement))),
			}
			s.l1Times[nextOrigin.ID()] = nextOrigin.Time
			return nextOrigin, nil
		} else {
			return origin, nil
		}
	})

	// Set up a fake Espresso client if necessary.
	if useEspresso {
		s.espresso = new(FakeEspressoClient)
	}

	s.seq = NewSequencer(log, &s.cfg, &s.engControl, attrBuilder, originSelector, s.espresso, metrics.NoopMetrics)
	s.seq.timeNow = s.clockFn

	return s
}

// SequencerChaosMonkey runs the sequencer in a mocked adversarial environment with
// repeated random errors in dependencies and poor clock timing.
// At the end the health of the chain is checked to show that the sequencer kept the chain in shape.
func SequencerChaosMonkey(t *testing.T, useEspresso bool) {
	s := SetupSequencer(t, useEspresso)

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
	require.Less(t, s.clockTime.Sub(time.Unix(int64(l2Head.Time), 0)).Abs(), 2*time.Second, "L2 time is accurate, within 2 seconds of wallclock")
	require.Greater(t, s.engControl.avgBuildingTime(), time.Second, "With 2 second block time and 1 second error backoff and healthy-on-average errors, building time should at least be a second")
	require.Greater(t, s.engControl.avgTxsPerBlock(), 3.0, "We expect at least 1 system tx per block, but with a mocked 0-10 txs we expect an higher avg")
}

func TestSequencerChaosMonkey(t *testing.T) {
	SequencerChaosMonkey(t, false)
}

func TestSequencerChaosMonkeyEspresso(t *testing.T) {
	SequencerChaosMonkey(t, true)
}

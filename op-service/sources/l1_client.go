package sources

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"

	"github.com/EspressoSystems/go-espresso-sequencer/hotshot"
	espresso "github.com/EspressoSystems/go-espresso-sequencer/types"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/sources/caching"
)

type L1ClientConfig struct {
	EthClientConfig

	L1BlockRefsCacheSize int
}

func L1ClientDefaultConfig(config *rollup.Config, trustRPC bool, kind RPCProviderKind) *L1ClientConfig {
	// Cache 3/2 worth of sequencing window of receipts and txs
	span := int(config.SeqWindowSize) * 3 / 2
	fullSpan := span
	if span > 1000 { // sanity cap. If a large sequencing window is configured, do not make the cache too large
		span = 1000
	}
	return &L1ClientConfig{
		EthClientConfig: EthClientConfig{
			// receipts and transactions are cached per block
			ReceiptsCacheSize:     span,
			TransactionsCacheSize: span,
			HeadersCacheSize:      span,
			PayloadsCacheSize:     span,
			MaxRequestsPerBatch:   20, // TODO: tune batch param
			MaxConcurrentRequests: 10,
			TrustRPC:              trustRPC,
			MustBePostMerge:       false,
			RPCProviderKind:       kind,
			MethodResetDuration:   time.Minute,
		},
		// Not bounded by span, to cover find-sync-start range fully for speedy recovery after errors.
		L1BlockRefsCacheSize: fullSpan,
	}
}

// L1Client provides typed bindings to retrieve L1 data from an RPC source,
// with optimized batch requests, cached results, and flag to not trust the RPC
// (i.e. to verify all returned contents against corresponding block hashes).
type L1Client struct {
	*EthClient

	// cache L1BlockRef by hash
	// common.Hash -> eth.L1BlockRef
	l1BlockRefsCache *caching.LRUCache[common.Hash, eth.L1BlockRef]
}

// NewL1Client wraps a RPC with bindings to fetch L1 data, while logging errors, tracking metrics (optional), and caching.
func NewL1Client(client client.RPC, log log.Logger, metrics caching.Metrics, config *L1ClientConfig) (*L1Client, error) {
	ethClient, err := NewEthClient(client, log, metrics, &config.EthClientConfig)
	if err != nil {
		return nil, err
	}

	return &L1Client{
		EthClient:        ethClient,
		l1BlockRefsCache: caching.NewLRUCache[common.Hash, eth.L1BlockRef](metrics, "blockrefs", config.L1BlockRefsCacheSize),
	}, nil
}

// L1BlockRefByLabel returns the [eth.L1BlockRef] for the given block label.
// Notice, we cannot cache a block reference by label because labels are not guaranteed to be unique.
func (s *L1Client) L1BlockRefByLabel(ctx context.Context, label eth.BlockLabel) (eth.L1BlockRef, error) {
	info, err := s.InfoByLabel(ctx, label)
	if err != nil {
		// Both geth and erigon like to serve non-standard errors for the safe and finalized heads, correct that.
		// This happens when the chain just started and nothing is marked as safe/finalized yet.
		if strings.Contains(err.Error(), "block not found") || strings.Contains(err.Error(), "Unknown block") {
			err = ethereum.NotFound
		}
		return eth.L1BlockRef{}, fmt.Errorf("failed to fetch head header: %w", err)
	}
	ref := eth.InfoToL1BlockRef(info)
	s.l1BlockRefsCache.Add(ref.Hash, ref)
	return ref, nil
}

// L1BlockRefByNumber returns an [eth.L1BlockRef] for the given block number.
// Notice, we cannot cache a block reference by number because L1 re-orgs can invalidate the cached block reference.
func (s *L1Client) L1BlockRefByNumber(ctx context.Context, num uint64) (eth.L1BlockRef, error) {
	info, err := s.InfoByNumber(ctx, num)
	if err != nil {
		return eth.L1BlockRef{}, fmt.Errorf("failed to fetch header by num %d: %w", num, err)
	}
	ref := eth.InfoToL1BlockRef(info)
	s.l1BlockRefsCache.Add(ref.Hash, ref)
	return ref, nil
}

// L1BlockRefByHash returns the [eth.L1BlockRef] for the given block hash.
// We cache the block reference by hash as it is safe to assume collision will not occur.
func (s *L1Client) L1BlockRefByHash(ctx context.Context, hash common.Hash) (eth.L1BlockRef, error) {
	if v, ok := s.l1BlockRefsCache.Get(hash); ok {
		return v, nil
	}
	info, err := s.InfoByHash(ctx, hash)
	if err != nil {
		return eth.L1BlockRef{}, fmt.Errorf("failed to fetch header by hash %v: %w", hash, err)
	}
	ref := eth.InfoToL1BlockRef(info)
	s.l1BlockRefsCache.Add(ref.Hash, ref)
	return ref, nil
}

// L1HotShotCommitmentsFromHeight returns an array of HotShot commitments to sequencer blocks
// This is used in the derivation pipeline to validate sequencer batches in Espresso mode
func (s *L1Client) L1HotShotCommitmentsFromHeight(firstBlockHeight uint64, numHeaders uint64, hotshotAddr common.Address) ([]espresso.Commitment, error) {
	var comms []espresso.Commitment
	client := ethclient.NewClient(s.client.RawClient())
	hotshot, err := hotshot.NewHotshot(hotshotAddr, client)
	if err != nil {
		return nil, err
	}

	// Check if the requested commitments are even available yet on L1.
	blockHeight, err := hotshot.HotshotCaller.BlockHeight(nil)
	if err != nil {
		return nil, err
	}
	if blockHeight.Cmp(big.NewInt(int64(firstBlockHeight+numHeaders))) < 0 {
		return nil, fmt.Errorf("commitments up to %d are not available (current block height %d)", firstBlockHeight+numHeaders, blockHeight)
	}

	for i := 0; i < int(numHeaders); i++ {
		height := big.NewInt(int64(firstBlockHeight + uint64(i)))
		commAsInt, err := hotshot.HotshotCaller.Commitments(nil, height)
		if err != nil {
			return comms, err
		}
		if commAsInt.Cmp(big.NewInt(0)) == 0 {
			// A commitment of 0 indicates that this commitment hasn't been set yet in the contract.
			// Since we checked the contract block height above, this can only happen if there was
			// a reorg on L1 just now. In this case, return an error rather than reporting
			// definitive commitments. The caller will retry and we will succeed eventually when we
			// manage to get a consistent snapshot of the L1.
			//
			// Note that in all other reorg cases, where the L1 reorgs but we read a nonzero
			// commitment, we are fine, since the HotShot contract will only ever record a single
			// ledger, consistent across all L1 forks, determined by HotShot consensus. The only
			// question is whether the recorded ledger extends far enough for the commitments we're
			// trying to read on the current fork of L1.
			return nil, fmt.Errorf("read 0 for commitment %d below block height %d, this indicates an L1 reorg", firstBlockHeight+uint64(i), blockHeight)
		}
		comm, err := espresso.CommitmentFromUint256(espresso.NewU256().SetBigInt(commAsInt))
		if err != nil {
			return comms, err
		}
		comms = append(comms, comm)
	}
	return comms, nil
}

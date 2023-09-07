package derive

import (
	"context"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-service/espresso"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type EspressoProvider struct {
	HotShotAddr common.Address
	L1Fetcher   L1Fetcher
	log         log.Logger
}

func NewEspressoProvider(log log.Logger, hotshotAddr common.Address, l1Fetcher L1Fetcher) *EspressoProvider {
	return &EspressoProvider{
		HotShotAddr: hotshotAddr,
		L1Fetcher:   l1Fetcher,
		log:         log,
	}

}

func (provider *EspressoProvider) VerifyCommitments(firstHeight uint64, comms []espresso.Commitment) (bool, error) {
	fetchedComms, err := provider.L1Fetcher.L1HotShotCommitmentsFromHeight(firstHeight, uint64(len(comms)), provider.HotShotAddr)
	if err != nil {
		return false, err
	}

	if len(fetchedComms) != len(comms) {
		return false, fmt.Errorf("fetched commitments has a different length than provided commitments (%d vs %d)", len(fetchedComms), len(comms))
	}

	for i, comm := range comms {
		if !comm.Equals(fetchedComms[i]) {
			provider.log.Warn("commitment does not match expected", "first", firstHeight, "i", i, "comm", comm, "expected", fetchedComms[i])
			return false, nil
		}
	}

	return true, nil
}

func (provider *EspressoProvider) L1BlockRefByNumber(ctx context.Context, num uint64) (eth.L1BlockRef, error) {
	return provider.L1Fetcher.L1BlockRefByNumber(ctx, num)
}

func (provider *EspressoProvider) FetchReceipts(ctx context.Context, blockHash common.Hash) (eth.BlockInfo, types.Receipts, error) {
	return provider.L1Fetcher.FetchReceipts(ctx, blockHash)
}

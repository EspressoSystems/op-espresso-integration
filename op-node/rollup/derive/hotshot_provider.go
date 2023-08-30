package derive

import (
	"bytes"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-service/espresso"
	"github.com/ethereum/go-ethereum/common"
)

type HotShotProvider struct {
	HotShotAddr common.Address
	L1Fetcher   L1Fetcher
}

func NewHotShotProvider(hotshotAddr common.Address, l1Fetcher L1Fetcher) *HotShotProvider {
	return &HotShotProvider{
		HotShotAddr: hotshotAddr,
		L1Fetcher:   l1Fetcher,
	}

}

func (provider *HotShotProvider) VerifyHeaders(headers []espresso.Header, height uint64) (bool, error) {
	fetchedHeaders, err := provider.GetCommitmentsFromHeight(height, uint64(len(headers)))
	if err != nil {
		return false, err
	}

	if len(fetchedHeaders) != len(headers) {
		return false, fmt.Errorf("fetched headers has a different length than provided headers (%d vs %d)", len(fetchedHeaders), len(headers))
	}

	for i := 0; i < len(fetchedHeaders); i++ {
		if !bytes.Equal(headers[i].TransactionsRoot.Root, fetchedHeaders[i].Root) {
			return false, nil
		}
	}

	return true, nil
}

func (provider *HotShotProvider) GetCommitmentsFromHeight(firstBlockHeight uint64, numHeaders uint64) ([]espresso.NmtRoot, error) {
	return provider.L1Fetcher.L1HotShotCommitmentsFromHeight(firstBlockHeight, numHeaders, provider.HotShotAddr)
}

package hotshot

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-service/espresso"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/ethclient"
)

type HotShotProvider struct {
	HotShot *Hotshot
}

func NewHotShotProvider(l1Url string, hotshotAddr string) (*HotShotProvider, error) {
	conn, err := ethclient.Dial(l1Url)
	if err != nil {
		return nil, err
	}
	hotshot, err := NewHotshot(common.HexToAddress(hotshotAddr), conn)
	if err != nil {
		return nil, err
	}
	return &HotShotProvider{
		HotShot: hotshot,
	}, nil

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
	var comms []espresso.NmtRoot
	for i := 0; i < int(numHeaders); i++ {
		var height big.Int
		height.SetUint64(firstBlockHeight + uint64(i))
		comm, err := provider.HotShot.HotshotCaller.Commitments(nil, &height)
		if err != nil {
			return comms, err
		}
		root := espresso.NmtRoot{
			Root: comm.Bytes(),
		}
		comms = append(comms, root)
	}
	return comms, nil
}

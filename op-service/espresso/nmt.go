package espresso

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// This function mocks batch transaction validatijon against a set of valid HotShot headers.
// It pretends to verify that the set of transactions (txns) in a batch correspond to a set of n NMT proofs
// (p1, ... pn) against headers h1,...hn.
//
// We assume that his function makes an external call to fetch the block headers beteen firstBlock and lastBlock.
//
// In other words, the function validates that txns = {...p1.txns, ..., ...pn.txns}. And that
// p1, ..., pn are all valid NMT proofs with respect to r1, ..., rn, the NMT roots of each header.
//
// The first and last block headers are also necessary to validate that the the NMT proofs are consistent with
// the transaction roots at the start and end of the window.
func ValidateBatchTransactions(transactions []hexutil.Bytes, nmtProofs []NmtProof, firstBlock Header, lastBlockHeader Header, firstBlockNumber uint64) error {
	// Even without true NMT verification, we can at least check that the number of NMT proofs is consistent
	// with the distance between the first and last block headers
	return nil
}

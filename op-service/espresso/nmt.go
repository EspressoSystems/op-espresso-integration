package espresso

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// This function mocks batch transaction validation against a set of valid HotShot headers.
// It pretends to verify that the set of transactions (txns) in a batch correspond to a set of n NMT proofs
// (p1, ... pn) against headers h1,...hn.
//
// In other words, the function validates that txns = {...p1.txns, ..., ...pn.txns}. And that
// p1, ..., pn are all valid NMT proofs with respect to r1, ..., rn, the NMT roots of each header.
func ValidateBatchTransactions(transactions []hexutil.Bytes, nmtProofs []NmtProof, headers []Header) error {
	return nil
}

package espresso

import (
	"context"
)

// Interface to the Espresso Sequencer query service.
type QueryService interface {
	// Get all the available headers whose timestamps fall in the window [start, end).
	FetchHeadersForWindow(ctx context.Context, start uint64, end uint64) (WindowStart, error)
	// Get all the available headers starting with the block numbered `from` whose timestamps are
	// less than `end`. This can be used to continue fetching headers in a time window if not all
	// headers in the window were available when `FetchHeadersForWindow` was called.
	FetchRemainingHeadersForWindow(ctx context.Context, from uint64, end uint64) (WindowMore, error)
	// Get the transactions belonging to the given namespace in the block numbered `block` with the
	// given header, along with a proof that these are all such transactions.
	FetchTransactionsInBlock(ctx context.Context, block uint64, header *Header, namespace uint64) (TransactionsInBlock, error)
}

// Response to `FetchHeadersForWindow`.
type WindowStart struct {
	// The block number of the first block in the window, unless the window is empty, in which case
	// this is the block number of `Next`.
	From uint64
	// The available block headers in the requested window.
	Window []Header
	// The header of the last block before the start of the window. This proves that the query
	// service did not omit any blocks from the beginning of the window. This will be `nil` if
	// `From` is 0.
	Prev *Header
	// The first block after the end of the window. This proves that the query service did not omit
	// any blocks from the end of the window. This will be `nil` if the full window is not available
	// yet, in which case `FetchRemainingHeadersForWindow` should be called to retrieve the rest of
	// the window.
	Next *Header
}

// Response to `FetchRemainingHeadersForWindow`.
type WindowMore struct {
	// The additional blocks within the window which are available, if any.
	Window []Header
	// The first block after the end of the window, if the full window is available.
	Next *Header
}

// Response to `FetchTransactionsInBlock`
type TransactionsInBlock struct {
	// The transactions.
	Transactions []Bytes
	// A proof that these are all the transactions in the block with the requested namespace.
	Proof NmtProof
}

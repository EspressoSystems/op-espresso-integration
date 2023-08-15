package espresso

import (
	"context"
	"fmt"
)

type Client struct {
	url string
}

func NewClient(url string) (*Client, error) {
	c := new(Client)
	c.url = url
	return c, nil
}

func (c *Client) FetchHeadersForWindow(ctx context.Context, start uint64, end uint64) ([]Header, uint64, error) {
	return nil, 0, fmt.Errorf("unimplemented: FetchHeadersForWindow")
}

func (c *Client) FetchRemainingHeadersForWindow(ctx context.Context, from uint64, end uint64) ([]Header, error) {
	return nil, fmt.Errorf("unimplemented: FetchRemainingHeadersForWindow")
}

func (c *Client) FetchTransactionsInBlock(ctx context.Context, block uint64, header *Header) ([]Bytes, NmtProof, error) {
	return nil, nil, fmt.Errorf("unimplemented: FetchTransactionsInBlock")
}

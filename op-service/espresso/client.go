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

func (c *Client) FetchHeadersForWindow(ctx context.Context, start uint64, end uint64) (WindowStart, error) {
	return WindowStart{}, fmt.Errorf("unimplemented: FetchHeadersForWindow")
}

func (c *Client) FetchRemainingHeadersForWindow(ctx context.Context, after uint64, end uint64) (WindowMore, error) {
	return WindowMore{}, fmt.Errorf("unimplemented: FetchRemainingHeadersForWindow")
}

func (c *Client) FetchTransactionsInBlock(ctx context.Context, block uint64, header *Header, namespace uint64) (TransactionsInBlock, error) {
	return TransactionsInBlock{}, fmt.Errorf("unimplemented: FetchTransactionsInBlock")
}

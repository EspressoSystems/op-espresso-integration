package espresso

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/log"
)

type Client struct {
	baseUrl string
	client  *http.Client
	log     log.Logger
}

func NewClient(log log.Logger, url string) *Client {
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return &Client{
		baseUrl: url,
		client:  http.DefaultClient,
		log:     log,
	}
}

func (c *Client) FetchHeadersForWindow(ctx context.Context, start uint64, end uint64) (WindowStart, error) {
	var res WindowStart
	if err := c.get(ctx, &res, "availability/headers/window/%d/%d", start, end); err != nil {
		return WindowStart{}, err
	}
	return res, nil
}

func (c *Client) FetchRemainingHeadersForWindow(ctx context.Context, from uint64, end uint64) (WindowMore, error) {
	var res WindowMore
	if err := c.get(ctx, &res, "availability/headers/window/from/%d/%d", from, end); err != nil {
		return WindowMore{}, err
	}
	return res, nil
}

func (c *Client) FetchTransactionsInBlock(ctx context.Context, block uint64, header *Header, namespace uint64) (TransactionsInBlock, error) {
	var res NamespaceResponse
	if err := c.get(ctx, &res, "availability/block/%d/namespace/%d", block, namespace); err != nil {
		return TransactionsInBlock{}, err
	}
	return res.Validate(header, namespace)
}

type NamespaceResponse struct {
	Proof        json.RawMessage
	Transactions []Transaction
}

// Validate a NamespaceResponse and extract the transactions.
// NMT proof validation is currently stubbed out.
func (res *NamespaceResponse) Validate(header *Header, namespace uint64) (TransactionsInBlock, error) {
	proof := NmtProof(res.Proof)
	// TODO validate `proof` against `header.TransactionsRoot`

	// Extract the transactions.
	var txs []Bytes
	for i, tx := range res.Transactions {
		if tx.Vm != namespace {
			return TransactionsInBlock{}, fmt.Errorf("transaction %d has wrong namespace (%d, expected %d)", i, tx.Vm, namespace)
		}
		txs = append(txs, tx.Payload)
	}

	return TransactionsInBlock{
		Transactions: txs,
		Proof:        proof,
	}, nil
}

type Transaction struct {
	Vm      uint64
	Payload Bytes
}

func (c *Client) get(ctx context.Context, out any, format string, args ...any) error {
	url := c.baseUrl + fmt.Sprintf(format, args...)

	c.log.Debug("get", "url", url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.log.Error("failed to build request", "err", err, "url", url)
		return err
	}
	res, err := c.client.Do(req)
	if err != nil {
		c.log.Error("error in request", "err", err, "url", url)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		// Try to get the response body to include in the error message, as it may have useful
		// information about why the request failed. If this call fails, the response will be `nil`,
		// which is fine to include in the log, so we can ignore errors.
		body, _ := io.ReadAll(res.Body)
		c.log.Error("request failed", "err", err, "url", url, "status", res.StatusCode, "response", string(body))
		return fmt.Errorf("request failed with status %d", res.StatusCode)
	}

	// Read the response body into memory before we unmarshal it, rather than passing the io.Reader
	// to the json decoder, so that we still have the body and can inspect it if unmarshalling
	// failed.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.log.Error("failed to read response body", "err", err, "url", url)
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		c.log.Error("failed to parse body as json", "err", err, "url", url, "response", string(body))
		return err
	}
	c.log.Debug("request completed successfully", "url", url, "res", res, "body", string(body), "out", out)
	return nil
}

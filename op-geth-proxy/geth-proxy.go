package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/peterbourgon/ff/v3"
)

// Environment variables beginning with this prefix can be used to instantiate command line flags
const ENV_PREFIX = "OP_GETH_PROXY"

var (
	fs            = flag.NewFlagSet("proxy", flag.ContinueOnError)
	fromAddr      = fs.String("listen-addr", "127.0.0.1:9090", "proxy's listening adress")
	sequencerAddr = fs.String("seq-addr", "http://127.0.0.1:50000", "address of espresso sequencer")
	gethAddr      = fs.String("geth-addr", "http://127.0.0.1:8545", "address of the op-geth node")
	vm_id         = fs.Int("vm-id", 1, "VM ID of the OP rollup instance")
)

type Transaction struct {
	Vm      int   `json:"vm"`
	Payload []int `json:"payload"`
}

type rpcMessage struct {
	Params []json.RawMessage `json:"params,omitempty"`
	Method string            `json:"method,omitempty"`
}

func ForwardToSequencer(message rpcMessage) {
	// json.RawMessage is a []byte array, which is marshalled
	// As a base64-encoded string. Our sequencer API expects a JSON array.
	payload := make([]int, len(message.Params[0]))
	for i := range payload {
		payload[i] = int(message.Params[0][i])
	}

	// Construct a transaction and send it to the sequencer
	txn := Transaction{
		Vm:      *vm_id,
		Payload: payload,
	}
	marshalled, err := json.Marshal(txn)
	if err != nil {
		panic(err)
	}
	request, err := http.NewRequest("POST", *sequencerAddr+"/submit/submit", bytes.NewBuffer(marshalled))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	log.Println("Transaction recieved, forwarding to sequencer.")
	response, err := client.Do(request)
	if err != nil {
		log.Println("Failed to connect to the sequencer: ", err)
	}
	if response.StatusCode != 200 {
		log.Println("Request failed. Here is the response: ", err)
	}
}

type baseHandle struct{}

func (h *baseHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	toUrl, err := url.Parse(*gethAddr)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(toUrl)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	// Once we've read the body, we need to replace it with another reader because
	// ReadAll can only be called once: https://blog.flexicondev.com/read-go-http-request-body-multiple-times
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	var message rpcMessage
	if err := json.Unmarshal(body, &message); err != nil {
		panic(err)
	}
	// Check for sendRawTransaction
	if message.Method == "eth_sendRawTransaction" {
		ForwardToSequencer(message)
	}
	proxy.ServeHTTP(w, r)
}

func main() {
	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix(ENV_PREFIX)); err != nil {
		panic(err)
	}

	h := &baseHandle{}
	http.Handle("/", h)

	log.Println("Starting proxy server on", *fromAddr)
	server := &http.Server{
		Addr:    *fromAddr,
		Handler: h,
	}
	log.Fatal(server.ListenAndServe())
}

Simple proxy service that parses payload data from an `eth_sendRawTransacton` RPC request and forwards it to Espresso sequencer.

### Usage
Default configuration:
``` bash
go run get-proxy.go
```
With CLI flags:
```bash
go run geth-proxy.go -listen-addr "127.0.0.1:9091" -vm-id 2
```
With env vars:
```
SEQUENCER_PROXY_LISTEN_ADDR=127.0.0.1:9095 go run geth-proxy.go 
```

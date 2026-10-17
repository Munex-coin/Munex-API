# MCX API Gateway (Go)

A statically-linked front door for the MCX node. It mirrors the node's HTTP/WS
surface exactly and forwards everything to the Python node unchanged.

## Design: no logic lives here

Consensus, storage, signing, fees, the maze game and the miner WebSocket
protocol all stay in `Desktop/node/` (the exact split of `MCX81.py` — untouched
logic). This service only:

- enforces the **same auth verdicts** as the node (`auth.go` mirrors
  `node/crypto.py::_rpc_remote_authorized`: loopback trust flag with
  fail-closed null refusal, `Authorization: Bearer` constant-time compare,
  `X-Forwarded-For` never consulted),
- proxies the **29 registered routes** (`routes.go`, same method+path table as
  the node's aiohttp setup) byte-for-byte — method, path, query, headers,
  body in; status, headers, body out, nothing parsed or rewritten,
- tunnels `/ws` + `/healthz` as a blind TCP stream to the miner-WS port, so
  WebSocket frames pass through bit-identical without the gateway ever
  parsing the miner protocol.

## Run

```powershell
# node on default ports (API 8085, WS 8080), gateway on 8086
.\mcxapi.exe --listen 127.0.0.1:8086 --upstream http://127.0.0.1:8085 --ws-target 127.0.0.1:8080

# bearer-gated deployment (proxy profile: loopback NOT trusted)
$env:MCX_RPC_TOKEN = "<same token as the node>"
.\mcxapi.exe --trust-loopback false
```

`--trust-loopback` accepts only `true`/`false`/empty (default true). Anything
else (`null`, typos) **refuses to boot naming `rpc_trust_loopback`** — the
same fail-closed rule as the node (FP1-null).

## Verify

```powershell
$env:GOTOOLCHAIN="local"
go vet ./...
go test -count=1 ./...
```

14 tests: the node's S1–S11 socket-suite auth verdicts, proxy transparency,
WS tunnel byte-echo, and full 29-route mounting.

## Route table (mirrors node registration order)

`GET /ws`, `GET /healthz` (tunnel) · `POST/GET /rpc` · `GET /stats /health
/peers /metrics` · `/api/v1/tokens…` (list, create, get, balance, transfer,
burn) + `POST /api/v1/burn` · `/api/v1/spatial/tokens…` (list, create, get,
balance, transfer) · `/api/v1/apps…` (list, deploy, get, state, call) ·
`/api/v1/automations…` (list, create, get, delete). Anything else → 404,
same as the node.

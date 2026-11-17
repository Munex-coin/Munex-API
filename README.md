# Munex-API

Developer scripts for the Munex testnet (chain 6621). Ten small
Python scripts plus a shared RPC helper, for reading chain state,
checking balances and nonces, sending transfers, deploying tokens,
and calling the SVF guard. The scripts carry no comments; they are
meant to be read and used as-is.

## What is inside

- `01_chain_info.py` — chain information
- `02_get_balance.py` — MCX balance of a wallet
- `03_get_nonce.py` — next nonce of a wallet
- `04_get_receipt.py` — transaction receipt
- `05_send_transfer.py` — send an MCX transfer
- `06_deploy_token.py` — deploy a token
- `07_register_profile.py` — register an SVF profile
- `08_call_requirepass.py` — call the SVF guard's requirePass
- `09_ws_balance.py` — live balance over WebSocket
- `munex_rpc.py` — shared RPC and WebSocket helper used by the
  scripts

The scripts talk to a running node: RPC at
`http://127.0.0.1:8085/rpc` and WebSocket at `ws://127.0.0.1:8080`
by default.

## Running a script

In this folder:

```
python 01_chain_info.py
```

## Status

Testnet software (chain 6621). TESTNET CANDIDATE. It has not been
independently audited. See [NOTICE.md](NOTICE.md).

## Related repositories

- [Munex-node](https://github.com/Munex-coin/Munex-node) — the node
- [Miner-codes](https://github.com/Munex-coin/Miner-codes) — miners
- [Munex-docs](https://github.com/Munex-coin/Munex-docs) — consensus rules, SVF, and the sample contract

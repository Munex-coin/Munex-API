#!/usr/bin/env python3
import argparse
import os
import sys
import time

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from munex_rpc import CHAIN_ID, MunexRPC, as_checksum              


def main():
    ap = argparse.ArgumentParser(description="Munex send native transfer")
    ap.add_argument("to", help="destination address (0x...)")
    ap.add_argument("amount", type=int, help="amount in wei (whole units)")
    ap.add_argument("--gas", type=int, default=100_000,
                    help="MAXIMUM gas (cap) — you only pay for gas actually "
                         "used (21000 for a simple transfer); default 100000")
    ap.add_argument("--gas-price", type=int, default=None,
                    help="gas price in wei (default: read eth_gasPrice)")
    ap.add_argument("--wait", action="store_true",
                    help="wait for the receipt (up to 420s)")
    ap.add_argument("--url", default=None, help="node /rpc URL")
    ap.add_argument("--token-file", default=None, help="rpc_token file")
    args = ap.parse_args()

    priv = os.environ.get("MUNEX_PRIVATE_KEY", "")
    if not priv:
        print("ERROR: MUNEX_PRIVATE_KEY is not set. Export it first:\n"
              "  export MUNEX_PRIVATE_KEY=0x...\n"
              "It is never written to a file or printed by this script.",
              file=sys.stderr)
        return 2
    try:
        from eth_account import Account
    except ImportError:
        print("ERROR: eth_account is not installed. Run this with the "
              "Munex venv python.", file=sys.stderr)
        return 2

    acct = Account.from_key(priv)
    sender = acct.address
    to = as_checksum(args.to)
    rpc = MunexRPC(url=args.url, token_file=args.token_file)
    nonce = rpc.get_nonce(sender)
    gas_price = args.gas_price if args.gas_price is not None else rpc.gas_price()

    tx = {"chainId": CHAIN_ID, "nonce": nonce, "gas": args.gas,
          "gasPrice": gas_price, "value": args.amount, "to": to}
    signed = acct.sign_transaction(tx)
    raw = "0x" + signed.raw_transaction.hex()
    tx_hash = rpc.send_raw(raw)
    print(f"sender: {sender}")
    print(f"to:     {to}")
                                                                  
                                                                      
    print(f"amount: {args.amount}  max fee: {args.gas * gas_price} "
          f"(cap: gas {args.gas} x gasPrice {gas_price})  nonce: {nonce}")
    print(f"tx_hash: {tx_hash}")
    if not args.wait:
        return 0
    t0 = time.time()
    while time.time() - t0 < 420:
        rcpt = rpc.get_receipt(tx_hash)
        if rcpt is not None:
            block = int(rcpt.get("blockNumber", "0x0"), 16)
            try:
                _gu = int(rcpt.get("gasUsed", "0x0"), 16)
            except (TypeError, ValueError):
                _gu = 0
            print(f"receipt: block {block} status {rcpt.get('status')} "
                  f"gas_used {_gu} fee charged {_gu * gas_price}")
            return 0 if rcpt.get("status") in ("0x1", 1, "1") else 1
        time.sleep(10)
    print(f"no receipt for {tx_hash} after 420s", file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main())

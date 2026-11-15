#!/usr/bin/env python3
import argparse
import os
import sys
import time

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from munex_rpc import MunexRPC              


def main():
    ap = argparse.ArgumentParser(description="Munex wait for receipt")
    ap.add_argument("tx_hash", help="transaction hash (0x...)")
    ap.add_argument("--timeout", type=int, default=420,
                    help="seconds to wait (default 420)")
    ap.add_argument("--interval", type=int, default=10,
                    help="poll interval in seconds (default 10)")
    ap.add_argument("--url", default=None, help="node /rpc URL")
    ap.add_argument("--token-file", default=None, help="rpc_token file")
    args = ap.parse_args()

    tx_hash = args.tx_hash if args.tx_hash.startswith("0x") else "0x" + args.tx_hash
    rpc = MunexRPC(url=args.url, token_file=args.token_file)

    tx = rpc.get_tx(tx_hash)
    print(f"tx: {tx}")
    if tx is None:
        print("transaction not known to this node yet", file=sys.stderr)
        return 1

    t0 = time.time()
    while time.time() - t0 < args.timeout:
        rcpt = rpc.get_receipt(tx_hash)
        if rcpt is not None:
            block = int(rcpt.get("blockNumber", "0x0"), 16)
            status = rcpt.get("status", "")
            print(f"receipt: block {block} status {status} "
                  f"gasUsed {rcpt.get('gasUsed')}")
            print(f"receipt: {rcpt}")
            if status in ("0x1", 1, "1"):
                print("sealed OK (status 0x1)")
                return 0
            print("sealed but FAILED (status 0x0)", file=sys.stderr)
            return 1
        print(f"waiting... {int(time.time() - t0)}s / {args.timeout}s",
              flush=True)
        time.sleep(args.interval)
    print(f"no receipt for {tx_hash} after {args.timeout}s", file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main())

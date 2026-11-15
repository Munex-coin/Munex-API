#!/usr/bin/env python3
import argparse
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from munex_rpc import MunexRPC              


def main():
    ap = argparse.ArgumentParser(description="Munex next nonce")
    ap.add_argument("address", help="wallet address (0x... or bare hex)")
    ap.add_argument("--url", default=None, help="node /rpc URL")
    ap.add_argument("--token-file", default=None, help="rpc_token file")
    args = ap.parse_args()

    addr = args.address if args.address.startswith("0x") else "0x" + args.address
    rpc = MunexRPC(url=args.url, token_file=args.token_file)
    nonce = rpc.get_nonce(addr)
    print(f"address: {addr.lower()}")
    print(f"next nonce: {nonce} (0x{nonce:x})")
    return 0


if __name__ == "__main__":
    sys.exit(main())

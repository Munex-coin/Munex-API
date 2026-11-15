#!/usr/bin/env python3
import argparse
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from munex_rpc import CHAIN_ID, MunexRPC              


def main():
    ap = argparse.ArgumentParser(description="Munex chain info")
    ap.add_argument("--url", default=None, help="node /rpc URL")
    ap.add_argument("--token-file", default=None, help="rpc_token file")
    args = ap.parse_args()

    rpc = MunexRPC(url=args.url, token_file=args.token_file)
    chain_id = rpc.chain_id()
    block = rpc.block_number()
    net_version = rpc.call("net_version")
    gas_price = rpc.gas_price()

    print(f"chain_id:     {chain_id} (0x{chain_id:x})")
    print(f"net_version:  {net_version}")
    print(f"block_number: {block}")
    print(f"gas_price:    {gas_price} wei")
    if chain_id != CHAIN_ID:
        print(f"ERROR: expected chain {CHAIN_ID} (Munex testnet), "
              f"got {chain_id}", file=sys.stderr)
        return 1
    print("chain OK: Munex testnet (6621)")
    return 0


if __name__ == "__main__":
    sys.exit(main())

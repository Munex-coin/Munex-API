#!/usr/bin/env python3
import argparse
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from munex_rpc import MunexRPC, as_checksum              


def require_pass_selector():
    from eth_hash.auto import keccak
    sel = keccak(b"requirePass()")[:4]
    assert sel.hex() == "4514a8d5", f"unexpected selector {sel.hex()}"
    return sel


def main():
    ap = argparse.ArgumentParser(description="Munex requirePass eth_call")
    ap.add_argument("token", help="token address (0x...)")
    ap.add_argument("--url", default=None, help="node /rpc URL")
    ap.add_argument("--token-file", default=None, help="rpc_token file")
    args = ap.parse_args()

    token = as_checksum(args.token)
    rpc = MunexRPC(url=args.url, token_file=args.token_file)
    data = "0x" + require_pass_selector().hex()
    result = rpc.eth_call(token, data)
    print(f"token: {token}")
    print(f"eth_call requirePass() -> {result}")
    if result == "0x01":
        print("APPROVED (0x01): certified, or unregistered token")
    elif result == "0x":
        print("REFUSED (0x): registered but not yet certified, "
              "or unsatisfiable profile")
    else:
        print(f"unexpected return value: {result}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())

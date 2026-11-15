#!/usr/bin/env python3
import argparse
import os
import struct
import sys
import time

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from munex_rpc import CHAIN_ID, MunexRPC, as_checksum              

PROFILE_MAGIC = b"MCXS"
SCHEMA_VERSION = 1
FAMILY_ID = 1


def selector(signature):
    from eth_hash.auto import keccak
    return keccak(signature.encode("ascii"))[:4]


def profile_bytes(sample_blocks=10, min_distinct_winners=2,
                  min_growth_blocks=0, min_weight=1000):
    out = bytearray()
    out += PROFILE_MAGIC
    out += struct.pack(">H", SCHEMA_VERSION)
    out += bytes([FAMILY_ID])
    out += bytes([sample_blocks])
    out += bytes([min_distinct_winners])
    out += bytes([min_growth_blocks])
    out += min_weight.to_bytes(32, "big")
    assert len(out) == 42, f"profile encoding is {len(out)} bytes, not 42"
    return bytes(out)


def abi_register_profile(prof):
    sel = selector("registerProfile(bytes)")
    assert sel.hex() == "b58f53b2", f"unexpected selector {sel.hex()}"
    data = sel
    data += (32).to_bytes(32, "big")                       
    data += len(prof).to_bytes(32, "big")                  
    data += prof
    data += b"\x00" * ((32 - len(prof) % 32) % 32)
    return data


def main():
    ap = argparse.ArgumentParser(description="Munex register SVF profile")
    ap.add_argument("token", help="deployed token address (0x...)")
    ap.add_argument("--sample-blocks", type=int, default=10)
    ap.add_argument("--min-winners", type=int, default=2)
    ap.add_argument("--min-weight", type=int, default=1000)
    ap.add_argument("--gas", type=int, default=1_000_000)
    ap.add_argument("--url", default=None, help="node /rpc URL")
    ap.add_argument("--token-file", default=None, help="rpc_token file")
    args = ap.parse_args()

    priv = os.environ.get("MUNEX_PRIVATE_KEY", "")
    if not priv:
        print("ERROR: MUNEX_PRIVATE_KEY is not set.", file=sys.stderr)
        return 2
    try:
        from eth_account import Account
    except ImportError:
        print("ERROR: eth_account is not installed. Run this with the "
              "Munex venv python.", file=sys.stderr)
        return 2

    token = as_checksum(args.token)
    prof = profile_bytes(args.sample_blocks, args.min_winners, 0,
                         args.min_weight)
    data = "0x" + abi_register_profile(prof).hex()
    print(f"profile: sample_blocks={args.sample_blocks} "
          f"min_distinct_winners={args.min_winners} "
          f"min_weight={args.min_weight} ({len(prof)} bytes)")

    acct = Account.from_key(priv)
    sender = acct.address
    rpc = MunexRPC(url=args.url, token_file=args.token_file)
    nonce = rpc.get_nonce(sender)
    gas_price = rpc.gas_price()
    tx = {"chainId": CHAIN_ID, "nonce": nonce, "gas": args.gas,
          "gasPrice": gas_price, "value": 0, "to": token, "data": data}
    signed = acct.sign_transaction(tx)
    raw = "0x" + signed.raw_transaction.hex()
    tx_hash = rpc.send_raw(raw)
    print(f"register tx: {tx_hash} nonce {nonce}")
    t0 = time.time()
    while time.time() - t0 < 420:
        rcpt = rpc.get_receipt(tx_hash)
        if rcpt is not None:
            block = int(rcpt.get("blockNumber", "0x0"), 16)
            print(f"receipt: block {block} status {rcpt.get('status')}")
            if rcpt.get("status") in ("0x1", 1, "1"):
                print(f"profile registered for {token}")
                return 0
            print("register tx failed (status 0x0)", file=sys.stderr)
            return 1
        time.sleep(10)
    print(f"no receipt for {tx_hash} after 420s", file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main())

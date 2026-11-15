#!/usr/bin/env python3
import argparse
import os
import sys
import time

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from munex_rpc import CHAIN_ID, MunexRPC              

                                              
                                                                  
TEST_TOKEN_INIT = "6038600c60003960386000f3"
TEST_TOKEN_RUNTIME = (
    "366000600037600060003660003473000000000000000000000000000000"
    "00000008005af13d600060003e156033573d6000f35b3d6000fd")
DEFAULT_CREATION_HEX = TEST_TOKEN_INIT + TEST_TOKEN_RUNTIME


def main():
    ap = argparse.ArgumentParser(description="Munex deploy a contract")
    ap.add_argument("--bytecode-file", default=None,
                    help="file with creation bytecode hex (default: "
                         "built-in SVF test token, direct_erc20)")
    ap.add_argument("--gas", type=int, default=1_000_000,
                    help="gas limit (default 1000000)")
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

    if args.bytecode_file:
        creation = open(args.bytecode_file).read().strip()
        creation = creation.removeprefix("0x")
    else:
        creation = DEFAULT_CREATION_HEX
        print("using built-in SVF test token bytecode "
              f"(direct_erc20, {len(creation) // 2} bytes: init + runtime)")

    acct = Account.from_key(priv)
    sender = acct.address
    rpc = MunexRPC(url=args.url, token_file=args.token_file)
    nonce = rpc.get_nonce(sender)
    gas_price = rpc.gas_price()
    tx = {"chainId": CHAIN_ID, "nonce": nonce, "gas": args.gas,
          "gasPrice": gas_price, "value": 0, "data": "0x" + creation}
    signed = acct.sign_transaction(tx)
    raw = "0x" + signed.raw_transaction.hex()
    tx_hash = rpc.send_raw(raw)
    print(f"deploy tx: {tx_hash} nonce {nonce}")
    t0 = time.time()
    while time.time() - t0 < 420:
        rcpt = rpc.get_receipt(tx_hash)
        if rcpt is not None:
            block = int(rcpt.get("blockNumber", "0x0"), 16)
            contract = rcpt.get("contractAddress", "")
            status = rcpt.get("status", "")
            print(f"receipt: block {block} status {status}")
            print(f"contract address: {contract}")
            if not contract or status not in ("0x1", 1, "1"):
                print("deploy failed or no contract address",
                      file=sys.stderr)
                return 1
            return 0
        time.sleep(10)
    print(f"no receipt for {tx_hash} after 420s", file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main())

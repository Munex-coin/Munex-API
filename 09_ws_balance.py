#!/usr/bin/env python3
import argparse
import asyncio
import json
import os
import sys


async def get_balance(ws_url, address):
    import websockets
                                                          
                                                           
    from urllib.parse import urlparse
    parts = urlparse(ws_url)
    if not parts.path or parts.path == "/":
        ws_url = ws_url.rstrip("/") + "/ws"
    async with websockets.connect(ws_url, max_size=10 * 1024 * 1024) as ws:
        await ws.send(json.dumps({"type": "get_balance", "wallet": address}))
        raw = await asyncio.wait_for(ws.recv(), timeout=30)
        return json.loads(raw)


def main():
    ap = argparse.ArgumentParser(description="Munex WS get_balance")
    ap.add_argument("address", help="wallet address (0x... or bare hex)")
    ap.add_argument("--ws-url", default=None,
                    help="node WS URL (default ws://127.0.0.1:8080)")
    args = ap.parse_args()

    addr = args.address if args.address.startswith("0x") else "0x" + args.address
    ws_url = args.ws_url or os.environ.get("MUNEX_WS_URL") or "ws://127.0.0.1:8080"
    try:
        reply = asyncio.run(get_balance(ws_url, addr.lower()))
    except Exception as e:
        print(f"WebSocket error: {e}", file=sys.stderr)
        return 1
    print(f"reply: {reply}")
    if reply.get("type") == "balance":
        bal = reply.get("balance", 0)
        print(f"address: {addr.lower()}")
        print(f"balance: {bal} wei ({bal / 1_000_000} MCX)")
        return 0
    print(f"unexpected reply: {reply}", file=sys.stderr)
    return 1


if __name__ == "__main__":
    sys.exit(main())

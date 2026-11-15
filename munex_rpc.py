
import json
import os
import urllib.error
import urllib.request

CHAIN_ID = 6621
DEFAULT_RPC_URL = "http://127.0.0.1:8085/rpc"
DEFAULT_WS_URL = "ws://127.0.0.1:8080"


def load_token(token=None, token_file=None):
    if token:
        return token
    env = os.environ.get("MUNEX_RPC_TOKEN")
    if env:
        return env.strip()
    if token_file:
        try:
            with open(token_file) as f:
                return f.read().strip()
        except OSError:
            return ""
    return ""


class MunexRPC:

    def __init__(self, url=None, token=None, token_file=None, timeout=30):
        self.url = url or os.environ.get("MUNEX_RPC_URL") or DEFAULT_RPC_URL
        self.token = load_token(token, token_file)
        if not self.token:
            import sys
            print("WARNING: no RPC token given (--token-file or "
                  "MUNEX_RPC_TOKEN). The node will likely reject the call "
                  "with 401. The token is the node's own rpc_token file, "
                  "in the same dir as its config.yaml.",
                  file=sys.stderr)
        self.timeout = timeout
        self._next_id = 1

    def call(self, method, params=None):
        body = json.dumps({
            "jsonrpc": "2.0",
            "id": self._next_id,
            "method": method,
            "params": params or [],
        }).encode()
        self._next_id += 1
        headers = {"Content-Type": "application/json"}
        if self.token:
            headers["Authorization"] = "Bearer " + self.token
        req = urllib.request.Request(self.url, data=body, headers=headers)
        try:
            with urllib.request.urlopen(req, timeout=self.timeout) as resp:
                out = json.load(resp)
        except urllib.error.HTTPError as e:
            raise RuntimeError(f"{method}: HTTP {e.code} from {self.url}") from e
        if "error" in out:
            raise RuntimeError(f"{method} failed: {out['error']}")
        return out["result"]

                                                                       
    def chain_id(self):
        return int(self.call("eth_chainId"), 16)

    def block_number(self):
        return int(self.call("eth_blockNumber"), 16)

    def get_balance(self, address):
        return int(self.call("eth_getBalance", [address, "latest"]), 16)

    def get_nonce(self, address):
        return int(
            self.call("eth_getTransactionCount", [address, "latest"]), 16)

    def gas_price(self):
        return int(self.call("eth_gasPrice"), 16)

    def get_receipt(self, tx_hash):
        return self.call("eth_getTransactionReceipt", [tx_hash])

    def get_tx(self, tx_hash):
        return self.call("eth_getTransactionByHash", [tx_hash])

    def eth_call(self, to, data, sender="0x0000000000000000000000000000000000000000"):
        return self.call("eth_call",
                         [{"from": sender, "to": to, "data": data}, "latest"])

    def send_raw(self, raw_tx_hex):
        return self.call("eth_sendRawTransaction", [raw_tx_hex])


def hex_to_int(value):
    return int(value, 16) if isinstance(value, str) else int(value)


def as_checksum(address):
    try:
        from eth_utils import to_checksum_address
        return to_checksum_address(address)
    except Exception:
        return address.lower()

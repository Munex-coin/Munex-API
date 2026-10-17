package main

// Message contracts shared with the Python node. The gateway never parses
// these (proxy/tunnel are opaque), but the registry documents the surface
// operators can expect through it. Source of truth stays node-side.

// RPCEnvelope is the JSON-RPC 2.0 envelope the node speaks on /rpc
// (both POST with a JSON body and GET are proxied; non-dict params yield a
// proper JSON-RPC error from the node itself).
type RPCEnvelope struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
	ID      any    `json:"id,omitempty"`
}

// Healthz is the exact /healthz contract served by the node's public
// miner-WS app: {"status": "ok", "service": "mcx-miner-ws"} — nothing else
// may ever appear there.
type Healthz struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// WSMessageTypes enumerates the miner-protocol message types the /ws tunnel
// carries (register, maze_move, wall_hit_solve, block_signed, ...). The
// tunnel forwards them opaquely; listed here for operator reference only.
var WSMessageTypes = []string{
	"register", "registered", "registration_failed",
	"maze_init", "maze_state", "maze_move", "maze_move_ack",
	"wall_hit", "wall_hit_solve", "wall_hit_solved",
	"memory_challenge_request", "memory_challenge_response",
	"get_maze_state", "get_balance", "balance",
	"produce_block", "block_signed", "block_height_update",
	"lottery_commit", "lottery_commit_ack", "lottery_reveal", "lottery_reveal_ack",
	"knowledge_report", "knowledge_ack",
	"send_tx", "tx_accepted", "tx_failed",
	"get_status", "get_miners", "get_blocks", "get_transactions", "get_tokens",
	"get_fee", "get_htlc", "ping", "pong", "uptime_ping", "error",
}

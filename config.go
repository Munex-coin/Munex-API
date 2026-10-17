// Package main implements the MCX API gateway: a Go front door that mirrors
// the Python node's HTTP/WS surface byte-for-byte and proxies to it.
//
// The gateway holds NO chain logic. Consensus, storage, signing and the miner
// WebSocket protocol all live in the Python node (Desktop/node/). Every
// request the gateway accepts is forwarded to the upstream node unchanged
// (method, path, query, headers, body); every response is relayed back
// untouched. This preserves node logic exactly while giving operators a
// statically-linked, low-footprint edge process.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config mirrors the subset of node operator settings the gateway enforces.
// Semantics intentionally match the Python node's RPC-AUTH-1 / RPC-LOOPBACK-1
// / FP1-null rules (see node/crypto.py::_rpc_remote_authorized and
// node/crypto.py::validate_rpc_security_config).
type Config struct {
	Listen       string // gateway bind address, e.g. "127.0.0.1:8086"
	Upstream     string // node base URL, e.g. "http://127.0.0.1:8085"
	WSTarget     string // node miner-WS address (host:port), e.g. "127.0.0.1:8080"
	RPCToken     string // bearer token; MCX_RPC_TOKEN env wins over --rpc-token
	TrustLoopRaw string // tri-state: "true", "false", or "" (unset)
}

// TrustLoopback resolves the tri-state operator flag exactly like the node:
//   - unset ("")            -> default true (local wallets keep working)
//   - "true" (any case)     -> loopback trusted without a token
//   - "false" (any case)    -> loopback NOT trusted; bearer token required
//   - anything else ("null" included, i.e. a malformed/null-ish value) ->
//     FAIL CLOSED: the gateway refuses to boot and names the key, mirroring
//     the node's null-trust refusal (FP1-null).
func (c Config) TrustLoopback() (bool, error) {
	v := strings.TrimSpace(c.TrustLoopRaw)
	if v == "" {
		return true, nil
	}
	switch strings.ToLower(v) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf(
			"refusing to boot: rpc_trust_loopback has an invalid value %q "+
				"(expected \"true\" or \"false\"); failing closed: loopback NOT trusted",
			c.TrustLoopRaw)
	}
}

// EffectiveToken returns the bearer token with the node's precedence:
// MCX_RPC_TOKEN env wins, then the configured value. Empty = unset.
func (c Config) EffectiveToken() string {
	if v := strings.TrimSpace(os.Getenv("MCX_RPC_TOKEN")); v != "" {
		return v
	}
	return strings.TrimSpace(c.RPCToken)
}

// LoadConfig parses flags/env. Fail-closed values are rejected by Validate.
func LoadConfig() Config {
	var c Config
	flag.StringVar(&c.Listen, "listen", "127.0.0.1:8086", "gateway bind address")
	flag.StringVar(&c.Upstream, "upstream", "http://127.0.0.1:8085", "node HTTP base URL")
	flag.StringVar(&c.WSTarget, "ws-target", "127.0.0.1:8080", "node miner-WS host:port")
	flag.StringVar(&c.RPCToken, "rpc-token", "", "bearer token (MCX_RPC_TOKEN env wins)")
	flag.StringVar(&c.TrustLoopRaw, "trust-loopback", "",
		"loopback trust: \"true\", \"false\", or empty for default-true; anything else refuses boot")
	flag.Parse()
	return c
}

// Validate enforces boot-time invariants. It returns nil only when the
// gateway may serve traffic.
func (c Config) Validate() error {
	if _, err := c.TrustLoopback(); err != nil {
		return err
	}
	if strings.TrimSpace(c.Upstream) == "" {
		return fmt.Errorf("refusing to boot: upstream node URL is empty")
	}
	if strings.TrimSpace(c.WSTarget) == "" {
		return fmt.Errorf("refusing to boot: ws-target is empty")
	}
	return nil
}

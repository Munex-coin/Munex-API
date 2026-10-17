package main

import (
	"crypto/subtle"
	"net"
	"net/http"
	"strings"
)

// loopback reports whether r's peer is a loopback address, mirroring the
// node's `ipaddress.ip_address(request.remote or "").is_loopback` check:
// 127.0.0.0/8 and ::1 count; everything else (including unparseable) does
// not. X-Forwarded-For / X-Real-IP are NEVER consulted — exactly like the
// node, a header cannot change the verdict in either direction.
func loopback(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(strings.TrimSpace(host))
	if ip == nil {
		return false
	}
	return ip.IsLoopback()
}

// bearer extracts a `Bearer <token>` credential. Scheme match is
// case-insensitive; anything else (Basic, raw token, empty) is absent.
func bearer(r *http.Request) (string, bool) {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if auth == "" {
		return "", false
	}
	if len(auth) < 7 || !strings.EqualFold(auth[:7], "bearer ") {
		return "", false
	}
	tok := strings.TrimSpace(auth[7:])
	if tok == "" {
		return "", false
	}
	return tok, true
}

// Authorized mirrors node/crypto.py::_rpc_remote_authorized:
//
//  1. Loopback peer + effective trust flag -> allow without a token.
//  2. Otherwise the configured bearer token must be present, compared
//     constant-time. No token configured, or unknown remote -> deny.
//
// It is pure apart from cfg reads, like the Python original.
func Authorized(r *http.Request, trustLoopback bool, want string) bool {
	if loopback(r) {
		if trustLoopback {
			return true
		}
	}
	if want == "" {
		return false
	}
	got, ok := bearer(r)
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

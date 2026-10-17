package main

import (
	"io"
	"net"
	"net/http"
	"time"
)

// proxyHandler forwards the request to the upstream node byte-for-byte
// (method, path, raw query, headers, body) and relays the response back
// untouched (status, headers, body). No JSON is parsed, no field is added,
// removed or rewritten: node logic is preserved exactly.
func proxyHandler(upstream string, trust bool, token string) http.HandlerFunc {
	client := &http.Client{Timeout: 60 * time.Second}
	return func(w http.ResponseWriter, r *http.Request) {
		if !Authorized(r, trust, token) {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		target := upstream + r.URL.RequestURI()
		req, err := http.NewRequestWithContext(r.Context(), r.Method, target, r.Body)
		if err != nil {
			http.Error(w, `{"error":"bad gateway"}`, http.StatusBadGateway)
			return
		}
		req.Header = r.Header.Clone()
		req.ContentLength = r.ContentLength
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, `{"error":"bad gateway"}`, http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		for k, vs := range resp.Header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}
}

// wsTunnelHandler bridges /ws to the node's miner WebSocket port as a blind
// TCP tunnel after the same auth gate. WebSocket frames are never parsed,
// so the miner protocol (register, maze_move, wall_hit_solve, block_signed,
// ...) passes through bit-identical in both directions.
func wsTunnelHandler(target string, trust bool, token string) http.HandlerFunc {
	dialer := &net.Dialer{Timeout: 30 * time.Second}
	return func(w http.ResponseWriter, r *http.Request) {
		if !Authorized(r, trust, token) {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		up, err := dialer.DialContext(r.Context(), "tcp", target)
		if err != nil {
			http.Error(w, `{"error":"bad gateway"}`, http.StatusBadGateway)
			return
		}
		defer up.Close()
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, `{"error":"hijack unsupported"}`, http.StatusInternalServerError)
			return
		}
		down, _, err := hj.Hijack()
		if err != nil {
			http.Error(w, `{"error":"hijack failed"}`, http.StatusInternalServerError)
			return
		}
		defer down.Close()
		// net/http already consumed the request line + headers, so re-emit
		// them (method, RequestURI, headers), then stream everything both
		// ways. All WebSocket frames after the 101 pass through untouched.
		// Suppress Go's default User-Agent: the handshake reaching the node
		// must contain only what the client sent.
		r.Header["User-Agent"] = []string{""}
		if err := r.Write(up); err != nil {
			return
		}
		go func() { _, _ = io.Copy(up, down) }()
		_, _ = io.Copy(down, up)
	}
}

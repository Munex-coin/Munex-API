package main

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The proxy must relay method/path/query/headers/body untouched and stream
// the upstream response back verbatim.
func TestProxyTransparent(t *testing.T) {
	var gotMethod, gotTarget, gotBody, gotHeader string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotTarget = r.Method, r.URL.RequestURI()
		gotHeader = r.Header.Get("X-Custom")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("X-Up", "1")
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer up.Close()

	gw := httptest.NewServer(proxyHandler(up.URL, true, ""))
	defer gw.Close()

	req, _ := http.NewRequest("POST", gw.URL+"/api/v1/tokens/0xabc/transfer?x=1&y=2",
		strings.NewReader(`{"a":1}`))
	req.Header.Set("X-Custom", "zz")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if gotMethod != "POST" || gotTarget != "/api/v1/tokens/0xabc/transfer?x=1&y=2" ||
		gotBody != `{"a":1}` || gotHeader != "zz" {
		t.Fatalf("upstream saw %q %q %q %q", gotMethod, gotTarget, gotBody, gotHeader)
	}
	if resp.StatusCode != 201 || resp.Header.Get("X-Up") != "1" || string(body) != `{"ok":true}` {
		t.Fatalf("client saw %d %q %q", resp.StatusCode, resp.Header.Get("X-Up"), body)
	}
}

func TestProxyAuthGate(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("upstream"))
	}))
	defer up.Close()

	gw := httptest.NewServer(proxyHandler(up.URL, false, "tok"))
	defer gw.Close()

	// Loopback without token -> 401 (trust=false in this gateway).
	resp, err := http.Get(gw.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

// The /ws tunnel must carry bytes both ways without parsing. A raw TCP echo
// server stands in for the node: whatever the client sends (upgrade request
// + framed payload) comes back verbatim.
func TestWSTunnelEcho(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer c.Close()
				_, _ = io.Copy(c, c) // echo
			}()
		}
	}()

	gw := httptest.NewServer(wsTunnelHandler(ln.Addr().String(), true, ""))
	defer gw.Close()

	addr := strings.TrimPrefix(gw.URL, "http://")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	marker := "X-Marker: tunnel-bytes-ok-12345\r\n"
	raw := "GET /ws HTTP/1.1\r\nHost: x\r\nUpgrade: websocket\r\n" + marker + "\r\n"
	if _, err := conn.Write([]byte(raw)); err != nil {
		t.Fatal(err)
	}
	br := bufio.NewReader(conn)
	echo := make([]byte, len(raw))
	if _, err := io.ReadFull(br, echo); err != nil {
		t.Fatal(err)
	}
	if string(echo) != raw {
		t.Fatalf("tunnel mangled bytes:\n%q", echo)
	}
	// A second frame-like payload round-trips too.
	frame := []byte{0x81, 0x05, 'h', 'e', 'l', 'l', 'o'}
	if _, err := conn.Write(frame); err != nil {
		t.Fatal(err)
	}
	back := make([]byte, len(frame))
	if _, err := io.ReadFull(br, back); err != nil {
		t.Fatal(err)
	}
	for i := range frame {
		if back[i] != frame[i] {
			t.Fatalf("frame byte %d mangled: %x vs %x", i, back[i], frame[i])
		}
	}
}

// All 31 node routes must be mounted (no silent 404s on real paths).
func TestRouteTableComplete(t *testing.T) {
	if len(Routes) != 29 {
		t.Fatalf("route table has %d entries, want 29", len(Routes))
	}
	mux := http.NewServeMux()
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer up.Close()
	RegisterRoutes(mux, Config{Upstream: up.URL, WSTarget: up.URL}, true, "")
	srv := httptest.NewServer(mux)
	defer srv.Close()

	paths := map[string]string{
		"/rpc": "ok", "/stats": "ok", "/health": "ok", "/peers": "ok",
		"/metrics": "ok", "/api/v1/tokens": "ok",
		"/api/v1/tokens/0xabc":                       "ok",
		"/api/v1/tokens/0xabc/balance/0xdef":         "ok",
		"/api/v1/spatial/tokens":                     "ok",
		"/api/v1/spatial/tokens/0xabc":               "ok",
		"/api/v1/spatial/tokens/0xabc/balance/0xdef": "ok",
		"/api/v1/apps":                               "ok", "/api/v1/apps/0xabc": "ok",
		"/api/v1/apps/0xabc/state": "ok", "/api/v1/automations": "ok",
		"/api/v1/automations/42": "ok",
	}
	for p := range paths {
		resp, err := http.Get(srv.URL + p)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("GET %s -> %d, want proxied 200", p, resp.StatusCode)
		}
	}
}

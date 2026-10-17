package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Mirrors the node's FP1 socket-suite verdicts (S1-S11) at the gateway layer.

func req(t *testing.T, remote, headers map[string]string) *http.Request {
	t.Helper()
	r := httptest.NewRequest("GET", "http://x/rpc", nil)
	r.RemoteAddr = "127.0.0.1:1234"
	if remote != nil {
		if v, ok := remote["remote"]; ok {
			r.RemoteAddr = v
		}
	}
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	return r
}

func TestLoopbackDefaultTrusted(t *testing.T) {
	if !Authorized(req(t, nil, nil), true, "") {
		t.Fatal("S1: loopback POST no token should be trusted by default")
	}
}

func TestXFFExternalNoEffect(t *testing.T) {
	// S2: X-Forwarded-For: external does not change a loopback verdict.
	r := req(t, nil, map[string]string{"X-Forwarded-For": "203.0.113.9"})
	if !Authorized(r, true, "") {
		t.Fatal("S2: XFF external must not remove loopback trust")
	}
	// S2b: XFF external does not GRANT trust to a real remote either.
	r2 := req(t, map[string]string{"remote": "203.0.113.9:555"}, nil)
	if Authorized(r2, true, "tok") {
		t.Fatal("S2b: remote without bearer must be denied despite default trust")
	}
}

func TestXFFLoopbackAddsNothing(t *testing.T) {
	// S3: XFF loopback adds no trust signal for a real remote.
	r := req(t, map[string]string{"remote": "203.0.113.9:555"},
		map[string]string{"X-Forwarded-For": "127.0.0.1"})
	if Authorized(r, true, "tok") {
		t.Fatal("S3: XFF loopback must not authenticate a remote")
	}
}

func TestFlagFalseDeniesLoopbackNoToken(t *testing.T) {
	// S5.
	if Authorized(req(t, nil, nil), false, "tok") {
		t.Fatal("S5: loopback no-token must be denied when trust=false")
	}
}

func TestBearerAccepted(t *testing.T) {
	// S6/S9: bearer accepted (loopback with flag=false, and remote).
	if !Authorized(req(t, nil, map[string]string{"Authorization": "Bearer tok"}), false, "tok") {
		t.Fatal("S6: bearer must be accepted")
	}
	r := req(t, map[string]string{"remote": "203.0.113.9:555"},
		map[string]string{"Authorization": "Bearer tok"})
	if !Authorized(r, false, "tok") {
		t.Fatal("S9: remote bearer must be accepted")
	}
}

func TestXFFCannotBypass(t *testing.T) {
	// S7: X-Forwarded-For: 127.0.0.1 cannot bypass trust=false.
	r := req(t, map[string]string{"remote": "203.0.113.9:555"},
		map[string]string{"X-Forwarded-For": "127.0.0.1"})
	if Authorized(r, false, "tok") {
		t.Fatal("S7: XFF must not bypass trust=false")
	}
}

func TestBearerSchemeStrict(t *testing.T) {
	for _, h := range []string{"tok", "Basic tok", "Bearer ", "bearer"} {
		r := req(t, map[string]string{"remote": "203.0.113.9:555"},
			map[string]string{"Authorization": h})
		if Authorized(r, false, "tok") {
			t.Fatalf("non-bearer auth %q must be denied", h)
		}
	}
	// Wrong token denied (constant-time compare).
	r := req(t, map[string]string{"remote": "203.0.113.9:555"},
		map[string]string{"Authorization": "Bearer wrong"})
	if Authorized(r, false, "tok") {
		t.Fatal("wrong bearer must be denied")
	}
}

func TestNoTokenConfiguredDeniesRemote(t *testing.T) {
	r := req(t, map[string]string{"remote": "203.0.113.9:555"},
		map[string]string{"Authorization": "Bearer anything"})
	if Authorized(r, false, "") {
		t.Fatal("remote must be denied when no token is configured")
	}
}

func TestNullTrustRefusesBoot(t *testing.T) {
	// S10/S11: null/malformed trust value refuses boot and names the key.
	for _, v := range []string{"null", "None", "maybe", "2"} {
		c := Config{TrustLoopRaw: v, Upstream: "http://127.0.0.1:8085", WSTarget: "127.0.0.1:8080"}
		err := c.Validate()
		if err == nil || !strings.Contains(err.Error(), "rpc_trust_loopback") {
			t.Fatalf("value %q must refuse boot naming rpc_trust_loopback, got %v", v, err)
		}
	}
	for raw, want := range map[string]bool{"": true, "true": true, "TRUE": true, "false": false} {
		c := Config{TrustLoopRaw: raw, Upstream: "http://127.0.0.1:8085", WSTarget: "127.0.0.1:8080"}
		if err := c.Validate(); err != nil {
			t.Fatalf("value %q must boot, got %v", raw, err)
		}
		got, _ := c.TrustLoopback()
		if got != want {
			t.Fatalf("value %q resolved %v, want %v", raw, got, want)
		}
	}
}

func TestIPv6Loopback(t *testing.T) {
	r := req(t, map[string]string{"remote": "[::1]:1234"}, nil)
	if !Authorized(r, true, "") {
		t.Fatal("::1 must count as loopback")
	}
}

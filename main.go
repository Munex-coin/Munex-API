package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	cfg := LoadConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("[mcxapi] %v", err)
	}
	trust, err := cfg.TrustLoopback()
	if err != nil {
		log.Fatalf("[mcxapi] %v", err) // unreachable: Validate covers it
	}
	token := cfg.EffectiveToken()

	mux := http.NewServeMux()
	RegisterRoutes(mux, cfg, trust, token)

	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           mux,
		ReadHeaderTimeout: 15 * time.Second,
	}
	log.Printf("[mcxapi] listening on %s -> node %s (ws %s) trust_loopback=%v token_set=%v",
		cfg.Listen, cfg.Upstream, cfg.WSTarget, trust, token != "")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[mcxapi] serve: %v", err)
	}
}

package main

import "net/http"

// Route mirrors one aiohttp registration in the Python node
// (node/core.py MicroCoreNode route setup + node/crypto.py::_ws_app_routes).
type Route struct {
	Method  string // GET, POST, DELETE
	Path    string // Go 1.22+ ServeMux pattern; {name} wildcards preserved
	Handler string // node-side handler name (for operators cross-checking logs)
	IsWS    bool   // true: blind TCP tunnel to the miner-WS port
}

// Routes is the complete node HTTP surface, in node registration order.
// Anything not listed here does not exist upstream and is answered 404 by
// the mux itself — identical to the node answering 404.
var Routes = []Route{
	{Method: "GET", Path: "/ws", Handler: "websocket_handler", IsWS: true},
	{Method: "GET", Path: "/healthz", Handler: "_ws_health_handler", IsWS: true},
	{Method: "POST", Path: "/rpc", Handler: "rpc_handler"},
	{Method: "GET", Path: "/rpc", Handler: "rpc_handler"},
	{Method: "GET", Path: "/stats", Handler: "stats_handler"},
	{Method: "GET", Path: "/health", Handler: "health_handler"},
	{Method: "GET", Path: "/peers", Handler: "peers_handler"},
	{Method: "GET", Path: "/metrics", Handler: "metrics_handler"},
	{Method: "GET", Path: "/api/v1/tokens", Handler: "api_tokens_list"},
	{Method: "POST", Path: "/api/v1/tokens", Handler: "api_tokens_create"},
	{Method: "GET", Path: "/api/v1/tokens/{address}", Handler: "api_tokens_get"},
	{Method: "GET", Path: "/api/v1/tokens/{address}/balance/{wallet}", Handler: "api_tokens_balance"},
	{Method: "POST", Path: "/api/v1/tokens/{address}/transfer", Handler: "api_tokens_transfer"},
	{Method: "POST", Path: "/api/v1/tokens/{address}/burn", Handler: "api_tokens_burn"},
	{Method: "POST", Path: "/api/v1/burn", Handler: "api_native_burn"},
	{Method: "GET", Path: "/api/v1/spatial/tokens", Handler: "api_spatial_tokens_list"},
	{Method: "POST", Path: "/api/v1/spatial/tokens", Handler: "api_spatial_token_create"},
	{Method: "GET", Path: "/api/v1/spatial/tokens/{address}", Handler: "api_spatial_token_get"},
	{Method: "GET", Path: "/api/v1/spatial/tokens/{address}/balance/{wallet}", Handler: "api_spatial_token_balance"},
	{Method: "POST", Path: "/api/v1/spatial/tokens/{address}/transfer", Handler: "api_spatial_token_transfer"},
	{Method: "GET", Path: "/api/v1/apps", Handler: "api_apps_list"},
	{Method: "POST", Path: "/api/v1/apps", Handler: "api_apps_deploy"},
	{Method: "GET", Path: "/api/v1/apps/{address}", Handler: "api_apps_get"},
	{Method: "GET", Path: "/api/v1/apps/{address}/state", Handler: "api_apps_state"},
	{Method: "POST", Path: "/api/v1/apps/{address}/call", Handler: "api_apps_call"},
	{Method: "GET", Path: "/api/v1/automations", Handler: "api_automations_list"},
	{Method: "POST", Path: "/api/v1/automations", Handler: "api_automations_create"},
	{Method: "GET", Path: "/api/v1/automations/{id}", Handler: "api_automations_get"},
	{Method: "DELETE", Path: "/api/v1/automations/{id}", Handler: "api_automations_delete"},
}

// RegisterRoutes mounts the full table on mux. HTTP routes proxy; WS routes
// tunnel. Auth (Authorized) is enforced identically on both.
func RegisterRoutes(mux *http.ServeMux, cfg Config, trust bool, token string) {
	for _, rt := range Routes {
		pattern := rt.Method + " " + rt.Path
		if rt.IsWS {
			mux.HandleFunc(pattern, wsTunnelHandler(cfg.WSTarget, trust, token))
		} else {
			mux.HandleFunc(pattern, proxyHandler(cfg.Upstream, trust, token))
		}
	}
}

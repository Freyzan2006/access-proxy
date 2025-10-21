package server

import (
	"net/http"
)

// Handler структуры для группировки связанных обработчиков
type infoHandlers struct {
	server *httpServer
}

func newInfoHandlers(server *httpServer) *infoHandlers {
	return &infoHandlers{server: server}
}

func (h *infoHandlers) rootHandler(w http.ResponseWriter, r *http.Request) {
	if !h.validateMethod(w, r, http.MethodGet) {
		return
	}

	response := map[string]interface{}{
		"service": "access-proxy",
		"status":  "running",
		"port":    h.server.port,
		"target":  h.server.target,
		"features": map[string]interface{}{
			"rate_limiting":       h.server.useRateLimit,
			"request_logging":     h.server.logRequests,
			"client_domain_check": len(h.server.allowedDomains) > 0,
			"method_restrictions": len(h.server.blockedMethods) > 0,
		},
		"endpoints": map[string]string{
			"health":      "/health",
			"config":      "/config",
			"ratelimit":   "/ratelimit-info",
			"client_info": "/client-info",
			"methods":     "/methods",
			"domains":     "/domains",
			"proxy":       "/* (proxies to target)",
		},
	}

	h.server.jsonResponse(w, response)
}

func (h *infoHandlers) healthHandler(w http.ResponseWriter, r *http.Request) {
	if !h.validateMethod(w, r, http.MethodGet) {
		return
	}

	response := map[string]interface{}{
		"status":  "healthy",
		"service": "access-proxy",
		"port":    h.server.port,
		"target":  h.server.target,
		"features": map[string]bool{
			"rate_limiting":       h.server.useRateLimit,
			"request_logging":     h.server.logRequests,
			"client_domain_check": len(h.server.allowedDomains) > 0,
			"method_restrictions": len(h.server.blockedMethods) > 0,
		},
		"client_allowed": h.server.isClientAllowed(r),
	}

	h.server.jsonResponse(w, response)
}

func (h *infoHandlers) configHandler(w http.ResponseWriter, r *http.Request) {
	if !h.validateMethod(w, r, http.MethodGet) {
		return
	}

	response := map[string]interface{}{
		"config": map[string]interface{}{
			"port":                  h.server.port,
			"target":                h.server.target,
			"rate_limit_per_minute": h.server.GetRateLimit(),
			"log_requests":          h.server.logRequests,
			"allowed_domains":       h.server.allowedDomains,
			"blocked_methods":       h.server.blockedMethods,
		},
	}

	h.server.jsonResponse(w, response)
}

func (h *infoHandlers) clientInfoHandler(w http.ResponseWriter, r *http.Request) {
	if !h.validateMethod(w, r, http.MethodGet) {
		return
	}

	clientDomain := h.server.extractClientDomain(r)
	clientIP := h.server.getClientIP(r)

	response := map[string]interface{}{
		"client_info": map[string]string{
			"ip":     clientIP,
			"domain": clientDomain,
		},
		"domain_restrictions": map[string]interface{}{
			"enabled":         len(h.server.allowedDomains) > 0,
			"allowed_domains": h.server.allowedDomains,
			"client_allowed":  h.server.isClientAllowed(r),
		},
	}

	h.server.jsonResponse(w, response)
}

func (h *infoHandlers) rateLimitInfoHandler(w http.ResponseWriter, r *http.Request) {
	if !h.validateMethod(w, r, http.MethodGet) {
		return
	}

	if !h.server.useRateLimit {
		h.server.jsonResponse(w, map[string]interface{}{
			"rate_limiting": false,
			"message":       "Rate limiting is disabled",
		})
		return
	}

	identifier := h.server.getClientIP(r)
	remaining := h.server.rateLimiter.GetRemaining(identifier)

	h.server.jsonResponse(w, map[string]interface{}{
		"rate_limiting": true,
		"limit":         h.server.rateLimiter.GetLimit(),
		"remaining":     remaining,
		"window":        "1 minute",
		"your_ip":       identifier,
	})
}

func (h *infoHandlers) domainsHandler(w http.ResponseWriter, r *http.Request) {
	if !h.validateMethod(w, r, http.MethodGet) {
		return
	}

	response := map[string]interface{}{
		"domain_restrictions": len(h.server.allowedDomains) > 0,
		"allowed_domains":     h.server.allowedDomains,
		"current_target":      h.server.target,
		"target_allowed":      h.server.isTargetAllowed(),
	}

	h.server.jsonResponse(w, response)
}

func (h *infoHandlers) methodsHandler(w http.ResponseWriter, r *http.Request) {
	if !h.validateMethod(w, r, http.MethodGet) {
		return
	}

	response := map[string]interface{}{
		"method_restrictions": map[string]interface{}{
			"enabled":         len(h.server.blockedMethods) > 0,
			"blocked_methods": h.server.blockedMethods,
			"allowed_methods": h.server.getAllowedMethods(),
		},
	}

	h.server.jsonResponse(w, response)
}

func (h *infoHandlers) validateMethod(w http.ResponseWriter, r *http.Request, allowedMethod string) bool {
	if r.Method != allowedMethod {
		h.server.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}
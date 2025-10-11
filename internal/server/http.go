// server/http_server.go
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"access-proxy/internal/middleware"
	"access-proxy/internal/ratelimit"

	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)

type HttpServer interface {
	RegisterEndpoints()
	ListenAndServe() error
}

type httpServer struct {
	proxy          ProxyServer
	port           int
	log            logger.Logger
	rateLimiter    *ratelimit.RateLimiter
	useRateLimit   bool
	target         string
	logRequests    bool
	allowedDomains []string
}

func NewHttpServer(proxy ProxyServer, port int, log logger.Logger, rateLimitPerMinute int, target string, logRequests bool, allowedDomains []string) HttpServer {
	var rateLimiter *ratelimit.RateLimiter
	useRateLimit := rateLimitPerMinute > 0
	
	if useRateLimit {
		rateLimiter = ratelimit.NewRateLimiter(rateLimitPerMinute, log)
		log.Infof("🔒 Rate limiting enabled: %d requests per minute", rateLimitPerMinute)
	}

	if logRequests {
		log.Info("📝 Request logging enabled")
	}

	if len(allowedDomains) > 0 {
		log.Infof("🌐 Client domain restrictions enabled: %v", allowedDomains)
	}

	return &httpServer{
		proxy:          proxy,
		port:           port,
		log:            log,
		rateLimiter:    rateLimiter,
		useRateLimit:   useRateLimit,
		target:         target,
		logRequests:    logRequests,
		allowedDomains: allowedDomains,
	}
}

func (s *httpServer) RegisterEndpoints() {
	// Создаем основной обработчик
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			s.rootHandler(w, r)
			return
		}
		
		if r.URL.Path == "/health" {
			s.healthHandler(w, r)
			return
		}
		
		if r.URL.Path == "/ratelimit-info" {
			s.rateLimitInfoHandler(w, r)
			return
		}
		
		if r.URL.Path == "/config" {
			s.configHandler(w, r)
			return
		}
		
		if r.URL.Path == "/client-info" {
			s.clientInfoHandler(w, r)
			return
		}
		
		// Все остальные пути через прокси
		s.proxy.ServeHTTP(w, r)
	})

	// Применяем middleware в правильном порядке
	var handler http.Handler = mainHandler
	
	// 1. Сначала проверка домена клиента
	if len(s.allowedDomains) > 0 {
		handler = middleware.ClientDomainValidator(s.log, s.allowedDomains)(handler)
	}
	
	// 2. Затем логирование
	if s.logRequests {
		handler = middleware.RequestLoggerMiddleware(s.log, true)(handler)
	}
	
	// 3. Затем rate limiting
	if s.useRateLimit {
		handler = s.rateLimiter.Middleware(handler)
	}

	// Устанавливаем финальный обработчик
	http.Handle("/", handler)
}

// Новый endpoint для информации о клиенте
func (s *httpServer) clientInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	clientDomain := s.extractClientDomain(r)
	clientIP := s.getClientIP(r)

	response := map[string]interface{}{
		"client_info": map[string]string{
			"ip":     clientIP,
			"domain": clientDomain,
		},
		"domain_restrictions": map[string]interface{}{
			"enabled":         len(s.allowedDomains) > 0,
			"allowed_domains": s.allowedDomains,
			"client_allowed":  s.isClientAllowed(r),
		},
	}

	s.jsonResponse(w, response)
}

// Вспомогательные методы для работы с клиентскими доменами
func (s *httpServer) extractClientDomain(r *http.Request) string {
	if origin := r.Header.Get("Origin"); origin != "" {
		return s.extractDomainFromURL(origin)
	}
	if referer := r.Header.Get("Referer"); referer != "" {
		return s.extractDomainFromURL(referer)
	}
	if host := r.Header.Get("Host"); host != "" {
		return strings.Split(host, ":")[0]
	}
	return ""
}

func (s *httpServer) extractDomainFromURL(urlStr string) string {
	if !strings.Contains(urlStr, "://") {
		urlStr = "https://" + urlStr
	}
	
	parts := strings.Split(urlStr, "://")
	if len(parts) < 2 {
		return ""
	}
	
	hostParts := strings.Split(parts[1], "/")
	host := hostParts[0]
	return strings.Split(host, ":")[0]
}

func (s *httpServer) isClientAllowed(r *http.Request) bool {
	if len(s.allowedDomains) == 0 {
		return true
	}
	
	clientDomain := s.extractClientDomain(r)
	for _, allowed := range s.allowedDomains {
		if clientDomain == allowed {
			return true
		}
	}
	return false
}

// Обновляем корневой handler
func (s *httpServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"service": "access-proxy",
		"status":  "running",
		"port":    s.port,
		"target":  s.target,
		"features": map[string]interface{}{
			"rate_limiting":        s.useRateLimit,
			"request_logging":      s.logRequests,
			"client_domain_check": len(s.allowedDomains) > 0,
		},
		"endpoints": map[string]string{
			"health":      "/health",
			"config":      "/config", 
			"ratelimit":   "/ratelimit-info",
			"client_info": "/client-info",
			"proxy":       "/* (proxies to target)",
		},
	}

	s.jsonResponse(w, response)
}

// Обновляем health handler
func (s *httpServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":  "healthy",
		"service": "access-proxy",
		"port":    s.port,
		"target":  s.target,
		"features": map[string]bool{
			"rate_limiting":        s.useRateLimit,
			"request_logging":      s.logRequests,
			"client_domain_check": len(s.allowedDomains) > 0,
		},
		"client_allowed": s.isClientAllowed(r),
	}

	s.jsonResponse(w, response)
}

// Информация о конфигурации
func (s *httpServer) configHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"config": map[string]interface{}{
			"port":                  s.port,
			"target":                s.target,
			"rate_limit_per_minute": s.GetRateLimit(),
			"log_requests":          s.logRequests,
			"allowed_domains":       s.allowedDomains,
			"blocked_methods":       []string{"DELETE", "PATCH"}, // из конфига
		},
	}

	s.jsonResponse(w, response)
}

// Информация о доменах
func (s *httpServer) domainsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"domain_restrictions": len(s.allowedDomains) > 0,
		"allowed_domains":     s.allowedDomains,
		"current_target":      s.target,
		"target_allowed":      s.isTargetAllowed(),
	}

	s.jsonResponse(w, response)
}

// Информация о rate limit
func (s *httpServer) rateLimitInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.useRateLimit {
		s.jsonResponse(w, map[string]interface{}{
			"rate_limiting": false,
			"message":       "Rate limiting is disabled",
		})
		return
	}

	identifier := s.getClientIP(r)
	remaining := s.rateLimiter.GetRemaining(identifier)

	s.jsonResponse(w, map[string]interface{}{
		"rate_limiting": true,
		"limit":         s.rateLimiter.GetLimit(),
		"remaining":     remaining,
		"window":        "1 minute",
		"your_ip":       identifier,
	})
}

// Вспомогательные методы
func (s *httpServer) isTargetAllowed() bool {
	if len(s.allowedDomains) == 0 {
		return true
	}
	
	targetDomain := s.extractDomain(s.target)
	for _, allowed := range s.allowedDomains {
		if targetDomain == allowed {
			return true
		}
	}
	return false
}

func (s *httpServer) extractDomain(urlStr string) string {
	// Упрощенная версия без обработки ошибок
	parts := make([]string, 2)
	copy(parts, splitTwo(urlStr, "://"))
	if len(parts) < 2 {
		return ""
	}
	
	hostParts := splitTwo(parts[1], "/")
	return splitTwo(hostParts[0], ":")[0]
}

func splitTwo(s, sep string) []string {
	parts := make([]string, 2)
	if idx := stringIndex(s, sep); idx >= 0 {
		parts[0] = s[:idx]
		parts[1] = s[idx+len(sep):]
	} else {
		parts[0] = s
	}
	return parts
}

func stringIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func (s *httpServer) getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func (s *httpServer) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.log.Errorf("❌ JSON encoding error: %v", err)
		http.Error(w, `{"error": "internal_server_error"}`, http.StatusInternalServerError)
	}
}

func (s *httpServer) jsonError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   http.StatusText(statusCode),
		"message": message,
	})
}

func (s *httpServer) ListenAndServe() error {
	addr := fmt.Sprintf(":%d", s.port)
	s.log.Infof("🚀 JSON Proxy API starting on http://localhost%s", addr)
	s.log.Infof("🎯 Target: %s", s.target)
	s.log.Infof("🔒 Rate limiting: %t", s.useRateLimit)
	s.log.Infof("📝 Request logging: %t", s.logRequests)
	s.log.Infof("🌐 Client domain restrictions: %t", len(s.allowedDomains) > 0)
	return http.ListenAndServe(addr, nil)
}

func (s *httpServer) GetRateLimit() int {
	if s.rateLimiter != nil {
		return s.rateLimiter.GetLimit()
	}
	return 0
}
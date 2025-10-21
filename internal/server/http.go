package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	blockedMethods []string

	// Внедренные компоненты
	domainUtils *domainUtils
}

func NewHttpServer(proxy ProxyServer, port int, log logger.Logger, rateLimitPerMinute int, target string, logRequests bool, allowedDomains []string, blockedMethods []string) HttpServer {
	server := &httpServer{
		proxy:          proxy,
		port:           port,
		log:            log,
		target:         target,
		logRequests:    logRequests,
		allowedDomains: allowedDomains,
		blockedMethods: blockedMethods,
		domainUtils:    newDomainUtils(allowedDomains),
	}

	server.setupRateLimiter(rateLimitPerMinute)
	server.logConfiguration()

	return server
}

func (s *httpServer) setupRateLimiter(rateLimitPerMinute int) {
	s.useRateLimit = rateLimitPerMinute > 0
	if s.useRateLimit {
		s.rateLimiter = ratelimit.NewRateLimiter(rateLimitPerMinute, s.log)
		s.log.Infof("🔒 Rate limiting enabled: %d requests per minute", rateLimitPerMinute)
	}
}

func (s *httpServer) logConfiguration() {
	if s.logRequests {
		s.log.Info("📝 Request logging enabled")
	}

	if len(s.allowedDomains) > 0 {
		s.log.Infof("🌐 Client domain restrictions enabled: %v", s.allowedDomains)
	}

	if len(s.blockedMethods) > 0 {
		s.log.Infof("🚫 Blocked methods: %v", s.blockedMethods)
	}
}

func (s *httpServer) RegisterEndpoints() {
	// Создаем компоненты
	handlers := newInfoHandlers(s)
	router := newRouter(s, handlers)
	middlewareBuilder := newMiddlewareBuilder(s)

	// Создаем и настраиваем обработчик
	mainHandler := router.createMainHandler()
	finalHandler := middlewareBuilder.build(mainHandler)

	http.Handle("/", finalHandler)
}

// Делегируем методы domainUtils
func (s *httpServer) extractClientDomain(r *http.Request) string {
	return s.domainUtils.extractClientDomain(r)
}

func (s *httpServer) isClientAllowed(r *http.Request) bool {
	return s.domainUtils.isClientAllowed(r)
}

func (s *httpServer) isTargetAllowed() bool {
	return s.domainUtils.isTargetAllowed(s.target)
}

func (s *httpServer) extractDomain(urlStr string) string {
	return s.domainUtils.extractDomain(urlStr)
}

// Остальные методы остаются в основном файле
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
	s.log.Infof("🚫 Method restrictions: %t", len(s.blockedMethods) > 0)
	return http.ListenAndServe(addr, nil)
}

func (s *httpServer) GetRateLimit() int {
	if s.rateLimiter != nil {
		return s.rateLimiter.GetLimit()
	}
	return 0
}

func (s *httpServer) getAllowedMethods() []string {
	allMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	if len(s.blockedMethods) == 0 {
		return allMethods
	}

	var allowed []string
	for _, method := range allMethods {
		if !s.isMethodBlocked(method) {
			allowed = append(allowed, method)
		}
	}
	return allowed
}

func (s *httpServer) isMethodBlocked(method string) bool {
	for _, blocked := range s.blockedMethods {
		if strings.EqualFold(blocked, method) {
			return true
		}
	}
	return false
}

// Вспомогательная функция (можно вынести в отдельный utils файл)
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
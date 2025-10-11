// server/http_server.go
package server

import (
	"fmt"
	"net/http"
	"strings"

	"os"

	"access-proxy/internal/middleware"
	"access-proxy/internal/ratelimit"

	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)

type HttpServer interface {
	RegisterEndpoints()
	ListenAndServe() error
}

type httpServer struct {
	proxy        ProxyServer
	port         int
	log          logger.Logger
	rateLimiter  *ratelimit.RateLimiter
	useRateLimit bool
	target       string
	logRequests  bool
}

func NewHttpServer(proxy ProxyServer, port int, log logger.Logger, rateLimitPerMinute int, target string, logRequests bool) HttpServer {
	var rateLimiter *ratelimit.RateLimiter
	useRateLimit := rateLimitPerMinute > 0
	
	if useRateLimit {
		rateLimiter = ratelimit.NewRateLimiter(rateLimitPerMinute, log)
		log.Infof("🔒 Rate limiting enabled: %d requests per minute", rateLimitPerMinute)
	}

	if logRequests {
		log.Info("📝 Request logging enabled")
	}

	return &httpServer{
		proxy:        proxy,
		port:         port,
		log:          log,
		rateLimiter:  rateLimiter,
		useRateLimit: useRateLimit,
		target:       target,
		logRequests:  logRequests,
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
		
		if r.URL.Path == "/favicon.ico" {
			s.faviconHandler(w, r)
			return
		}
		
		// Все остальные пути через прокси
		s.proxy.ServeHTTP(w, r)
	})

	// Применяем middleware в правильном порядке
	var handler http.Handler = mainHandler
	
	// 1. Сначала логирование (самый внешний слой)
	if s.logRequests {
		handler = middleware.RequestLoggerMiddleware(s.log, true)(handler)
		s.log.Info("📝 Request logging middleware applied")
	}
	
	// 2. Затем rate limiting
	if s.useRateLimit {
		handler = s.rateLimiter.Middleware(handler)
		s.log.Info("🛡️  Rate limit middleware applied")
	}

	// Устанавливаем финальный обработчик
	http.Handle("/", handler)
}

func (s *httpServer) rateLimitInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	
	if !s.useRateLimit {
		fmt.Fprintf(w, `{"rate_limiting": false, "message": "Rate limiting is disabled"}`)
		return
	}
	
	identifier := s.getClientIP(r)
	remaining := s.rateLimiter.GetRemaining(identifier)
	
	fmt.Fprintf(w, `{
		"rate_limiting": true,
		"limit": %d,
		"remaining": %d,
		"window": "1 minute",
		"your_ip": "%s"
	}`, s.rateLimiter.GetLimit(), remaining, identifier)
}

func (s *httpServer) getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func (s *httpServer) faviconHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

func (s *httpServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Читаем HTML из файла
	htmlContent, err := os.ReadFile("static/index.html")
	if err != nil {
		// Если файла нет, используем простой HTML
		s.log.Warnf("Файл index.html не найден, используем встроенный шаблон")
		htmlContent = []byte(`
<!DOCTYPE html>
<html>
<head><title>Access Proxy</title></head>
<body>
	<h1>🚀 Access Proxy Server</h1>
	<p>Работает на порту: ` + fmt.Sprintf("%d", s.port) + `</p>
	<p>Target: ` + s.target + `</p>
	{{if .RateLimitEnabled}}<p>Rate Limit: ` + fmt.Sprintf("%d", s.GetRateLimit()) + `/min</p>{{end}}
	{{if .LogRequests}}<p>📝 Logging: ENABLED</p>{{end}}
	<p><a href="/json">/json</a> | <a href="/ip">/ip</a></p>
</body>
</html>`)
	}

	// Заменяем простые шаблоны
	htmlStr := string(htmlContent)
	
	if s.useRateLimit {
		htmlStr = strings.ReplaceAll(htmlStr, "{{.RateLimitEnabled}}", "true")
		htmlStr = strings.ReplaceAll(htmlStr, "{{.RateLimit}}", fmt.Sprintf("%d", s.GetRateLimit()))
	} else {
		htmlStr = strings.ReplaceAll(htmlStr, "{{.RateLimitEnabled}}", "false")
		htmlStr = strings.ReplaceAll(htmlStr, "{{if .RateLimitEnabled}}", "")
		htmlStr = strings.ReplaceAll(htmlStr, "{{end}}", "")
	}
	
	htmlStr = strings.ReplaceAll(htmlStr, "{{.Target}}", s.target)
	
	if s.logRequests {
		htmlStr = strings.ReplaceAll(htmlStr, "{{.LogRequests}}", "true")
		htmlStr = strings.ReplaceAll(htmlStr, "{{if .LogRequests}}", "")
		htmlStr = strings.ReplaceAll(htmlStr, "{{end}}", "")
	} else {
		htmlStr = strings.ReplaceAll(htmlStr, "{{.LogRequests}}", "false")
		htmlStr = strings.ReplaceAll(htmlStr, "{{if .LogRequests}}", "")
		htmlStr = strings.ReplaceAll(htmlStr, "{{end}}", "")
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlStr))
}

func (s *httpServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"status": "healthy", 
		"service": "access-proxy", 
		"port": %d, 
		"rate_limiting": %t,
		"request_logging": %t,
		"target": "%s"
	}`, s.port, s.useRateLimit, s.logRequests, s.target)
}

func (s *httpServer) ListenAndServe() error {
	addr := fmt.Sprintf(":%d", s.port)
	s.log.Infof("🚀 Server starting on http://localhost%s", addr)
	s.log.Infof("🔒 Rate limiting: %t (limit: %d req/min)", s.useRateLimit, s.rateLimiter.GetLimit())
	return http.ListenAndServe(addr, nil)
}

// Добавим метод для получения лимита
func (s *httpServer) GetRateLimit() int {
	if s.rateLimiter != nil {
		return s.rateLimiter.GetLimit()
	}
	return 0
}
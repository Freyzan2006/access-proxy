// server/http_server.go
package server

import (
	"fmt"
	"net/http"
	"os"

	"access-proxy/internal/ratelimit"

	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)

type HttpServer interface {
	RegisterEndpoints()
	ListenAndServe() error
}

type httpServer struct {
	proxy      ProxyServer
	port       int
	log        logger.Logger
	rateLimiter *ratelimit.RateLimiter
	useRateLimit bool
}

func NewHttpServer(proxy ProxyServer, port int, log logger.Logger, rateLimitPerMinute int) HttpServer {
	var rateLimiter *ratelimit.RateLimiter
	useRateLimit := rateLimitPerMinute > 0
	
	if useRateLimit {
		rateLimiter = ratelimit.NewRateLimiter(rateLimitPerMinute, log)
		log.Infof("🔒 Rate limiting enabled: %d requests per minute", rateLimitPerMinute)
	}

	return &httpServer{
		proxy:        proxy,
		port:         port,
		log:          log,
		rateLimiter:  rateLimiter,
		useRateLimit: useRateLimit,
	}
}

func (s *httpServer) RegisterEndpoints() {
	// Создаем основной обработчик
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.log.Infof("🌐 %s %s", r.Method, r.URL.Path)
		
		// Корневой путь - информационная страница
		if r.URL.Path == "/" {
			s.rootHandler(w, r)
			return
		}
		
		// Все остальные пути через прокси
		s.proxy.ServeHTTP(w, r)
	})

	// Оборачиваем в rate limiter если включен
	var finalHandler http.Handler = mainHandler
	if s.useRateLimit {
		finalHandler = s.rateLimiter.Middleware(mainHandler)
		s.log.Info("🛡️  Rate limit middleware applied")
	}

	// Устанавливаем обработчики
	http.Handle("/", finalHandler)
	http.HandleFunc("/favicon.ico", s.faviconHandler)
	http.HandleFunc("/health", s.healthHandler)
	http.HandleFunc("/ratelimit-info", s.rateLimitInfoHandler)
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
		s.log.Warnf("Файл index.html не найден.")
	}

	// Заменяем простые шаблоны
	htmlStr := string(htmlContent)
	
	

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlStr))
}

func (s *httpServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := "healthy"
	if s.useRateLimit {
		status = "healthy_with_rate_limiting"
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "%s", "service": "access-proxy", "port": %d, "rate_limiting": %t}`,
		status, s.port, s.useRateLimit)
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
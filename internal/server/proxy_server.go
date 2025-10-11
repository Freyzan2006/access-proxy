// server/proxy_server.go
package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)

type ProxyServer interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type proxyServer struct {
	proxy *httputil.ReverseProxy
	log   logger.Logger
}

func NewProxyServer(target string, log logger.Logger) ProxyServer {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("❌ Failed to parse target URL in proxy: %v", err)
	}
	
	log.Infof("🎯 Proxy target: %s", targetURL.String())

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	
	// Модифицируем Director для правильных заголовков
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		// Устанавливаем правильные заголовки для API
		req.Header.Set("User-Agent", "Access-Proxy-Server/1.0")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Accept-Encoding", "identity") // Отключаем сжатие для простоты
		
		// Удаляем проблемные заголовки
		req.Header.Del("Accept-Encoding")
		req.Header.Del("X-Forwarded-Proto")
		
		log.Infof("➡️  Forwarding to %s %s", req.Method, req.URL.String())
	}

	// Улучшенное логирование
	proxy.ModifyResponse = func(resp *http.Response) error {
		log.Infof("📨 Response: %d %s for %s", resp.StatusCode, resp.Status, resp.Request.URL.Path)
		return nil
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Errorf("❌ Proxy error: %v", err)
		log.Errorf("❌ Request: %s %s", r.Method, r.URL.String())
		w.WriteHeader(http.StatusBadGateway)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "Bad Gateway", "message": "` + err.Error() + `"}`))
	}

	return &proxyServer{
		proxy: proxy,
		log:   log,
	}
}

func (p *proxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	p.log.Infof("📥 Incoming: %s %s", r.Method, r.URL.String())
	
	// Устанавливаем правильные заголовки ответа
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	
	p.proxy.ServeHTTP(w, r)
	
	duration := time.Since(start)
	p.log.Infof("✅ Completed: %s %s in %v", r.Method, r.URL.Path, duration)
}


// server/proxy_server.go
// package server

// import (
// 	"net/http"
// 	"net/http/httputil"
// 	"net/url"

// 	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
// )

// type ProxyServer interface {
// 	ServeHTTP(w http.ResponseWriter, r *http.Request)
// }

// type proxyServer struct {
// 	proxy *httputil.ReverseProxy
// 	log   logger.Logger
// }

// func NewProxyServer(target string, log logger.Logger) ProxyServer {
// 	targetURL, err := url.Parse(target)
// 	if err != nil {
// 		log.Fatalf("Failed to parse target URL: %v", err)
// 	}

// 	// Создаем базовый ReverseProxy
// 	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	
// 	// Сохраняем оригинальный Director для логов
// 	originalDirector := proxy.Director
// 	proxy.Director = func(req *http.Request) {
// 		originalDirector(req) // Вызываем оригинальную логику
		
// 		// Логируем полный URL после применения директора
// 		log.Infof("➡️  Forwarding to %s %s", req.Method, req.URL.String())
		
// 		// Устанавливаем дополнительные заголовки
// 		req.Header.Set("X-Forwarded-Host", req.Host)
// 		req.Header.Set("X-Proxy-Server", "access-proxy")
// 	}

// 	// Обработка ошибок
// 	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
// 		log.Errorf("❌ Proxy error: %v", err)
// 		w.WriteHeader(http.StatusBadGateway)
// 		w.Header().Set("Content-Type", "application/json")
// 		w.Write([]byte(`{"error": "Bad Gateway", "message": "` + err.Error() + `"}`))
// 	}

// 	// Модификация ответов
// 	proxy.ModifyResponse = func(resp *http.Response) error {
// 		log.Infof("⬅️  Response received: %d %s", resp.StatusCode, resp.Status)
// 		return nil
// 	}

// 	return &proxyServer{
// 		proxy: proxy,
// 		log:   log,
// 	}
// }

// func (p *proxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	p.proxy.ServeHTTP(w, r)
// }
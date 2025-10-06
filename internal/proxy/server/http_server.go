package server

import (
	"fmt"
	"net/http"


	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)

type HttpServer interface {
	Endpoints()
	ListenAndServe() error
}

type httpServer struct {
	proxy 	ProxyServer
	port 	int
	log 	logger.Logger
	handler http.Handler
}

func NewHttpServer(proxy ProxyServer, port int, log logger.Logger) *httpServer {
	return &httpServer{
		proxy:  proxy,
		port:   port,
		log:    log,
		handler: http.DefaultServeMux, // стандартный обработчик
	}
}


func (s *httpServer) ListenAndServe() error {
	s.log.Infof("🚀 Proxy сервер запущен на http://localhost:%d", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.handler)
}


func (s *httpServer) Endpoints() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.log.Infof("📡 %s %s", r.Method, r.URL.Path)
		s.proxy.ServeHTTP(w, r)
	})
}

package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)


type ProxyServer interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type proxyServer struct {
	core *httputil.ReverseProxy
	log  logger.Logger
}


func NewProxyServer(target string, log logger.Logger) ProxyServer {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Ошибка разбора URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
	}

	
	return &proxyServer{
		core: proxy,
	}
}

func (p *proxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.core.ServeHTTP(w, r)
}


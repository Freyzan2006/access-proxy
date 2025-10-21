package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)

type proxyBuilder struct {
	targetURL *url.URL
	log       logger.Logger
	req 	  *requestProcessor
	res       *responseProcessor
	err       *errorHandler
}

func newProxyBuilder(targetURL *url.URL, log logger.Logger) *proxyBuilder {
	return &proxyBuilder{
		targetURL: targetURL,
		log:       log,
		req:       newRequestProcessor(log),
		res:       newResponseProcessor(log),
		err:       newErrorHandler(log),
	}
}

func (b *proxyBuilder) build() http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(b.targetURL)
	
	b.setupDirector(proxy)
	b.setupResponseModifier(proxy)
	b.setupErrorHandler(proxy)

	return proxy
}

func (b *proxyBuilder) setupDirector(proxy *httputil.ReverseProxy) {
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		b.modifyRequestHeaders(req)
		b.req.logRequest(req)
	}
}

func (b *proxyBuilder) modifyRequestHeaders(req *http.Request) {
	b.req.setStandardHeaders(req)
	b.req.removeProblematicHeaders(req)
}

func (b *proxyBuilder) setupResponseModifier(proxy *httputil.ReverseProxy) {
	proxy.ModifyResponse = func(resp *http.Response) error {
		b.res.logResponse(resp)
		return nil
	}
}

func (b *proxyBuilder) setupErrorHandler(proxy *httputil.ReverseProxy) {
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		b.err.handleError(w, r, err)
	}
}


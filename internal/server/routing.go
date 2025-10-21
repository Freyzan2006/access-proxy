package server

import "net/http"

type router struct {
	handlers *infoHandlers
	server   *httpServer
}

func newRouter(server *httpServer, handlers *infoHandlers) *router {
	return &router{
		handlers: handlers,
		server:   server,
	}
}

func (r *router) createMainHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/":
			r.handlers.rootHandler(w, req)
		case "/health":
			r.handlers.healthHandler(w, req)
		case "/ratelimit-info":
			r.handlers.rateLimitInfoHandler(w, req)
		case "/config":
			r.handlers.configHandler(w, req)
		case "/client-info":
			r.handlers.clientInfoHandler(w, req)
		case "/methods":
			r.handlers.methodsHandler(w, req)
		case "/domains":
			r.handlers.domainsHandler(w, req)
		default:
			// Все остальные пути через прокси
			r.server.proxy.ServeHTTP(w, req)
		}
	})
}
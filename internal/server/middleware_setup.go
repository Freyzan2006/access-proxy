package server

import (
	"net/http"

	"access-proxy/internal/middleware"
)

type middlewareBuilder struct {
	server *httpServer
}

func newMiddlewareBuilder(server *httpServer) *middlewareBuilder {
	return &middlewareBuilder{server: server}
}

func (b *middlewareBuilder) build(handler http.Handler) http.Handler {
	// Порядок применения middleware (от внешнего к внутреннему)
	middlewares := []func(http.Handler) http.Handler{}

	// 1. Блокировка методов
	if len(b.server.blockedMethods) > 0 {
		middlewares = append(middlewares, 
			middleware.MethodBlockerMiddleware(b.server.log, b.server.blockedMethods))
	}

	// 2. Проверка домена клиента
	if len(b.server.allowedDomains) > 0 {
		middlewares = append(middlewares, 
			middleware.ClientDomainValidator(b.server.log, b.server.allowedDomains))
	}

	// 3. Логирование
	if b.server.logRequests {
		middlewares = append(middlewares, 
			middleware.RequestLoggerMiddleware(b.server.log, true))
	}

	// 4. Rate limiting
	if b.server.useRateLimit {
		middlewares = append(middlewares, b.server.rateLimiter.Middleware)
	}

	// Применяем middleware в обратном порядке (последний становится самым внешним)
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}
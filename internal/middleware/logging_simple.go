// internal/middleware/logging_simple.go
package middleware

import (
	"net/http"
	"time"

	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)

// RequestLoggerMiddleware упрощенное логирование запросов
func RequestLoggerMiddleware(log logger.Logger, enabled bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !enabled {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			
			// Логируем начало запроса
			log.Infof("➡️  REQUEST_START: %s %s from %s", r.Method, r.URL.String(), r.RemoteAddr)
			
			// Продолжаем обработку
			next.ServeHTTP(w, r)
			
			// Логируем завершение
			duration := time.Since(start)
			log.Infof("⬅️  REQUEST_END: %s %s - %v", r.Method, r.URL.Path, duration)
		})
	}
}
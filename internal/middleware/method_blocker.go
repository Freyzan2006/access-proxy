// internal/middleware/method_blocker.go
package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)

// MethodBlockerMiddleware блокирует указанные HTTP методы
func MethodBlockerMiddleware(log logger.Logger, blockedMethods []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Если список методов пустой - пропускаем все
			if len(blockedMethods) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Нормализуем метод (верхний регистр)
			method := strings.ToUpper(r.Method)

			// Проверяем заблокирован ли метод
			if isMethodBlocked(method, blockedMethods) {
				log.Warnf("🚫 Method blocked: %s %s", method, r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "method_not_allowed",
					"message": "HTTP method is not allowed",
					"method":  method,
					"blocked_methods": blockedMethods,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isMethodBlocked проверяет заблокирован ли метод
func isMethodBlocked(method string, blockedMethods []string) bool {
	for _, blocked := range blockedMethods {
		if strings.ToUpper(blocked) == method {
			return true
		}
	}
	return false
}
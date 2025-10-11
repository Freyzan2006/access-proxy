// internal/middleware/domain.go
package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)

// DomainValidator проверяет что целевой домен разрешен
func DomainValidator(log logger.Logger, allowedDomains []string, target string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(allowedDomains) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			targetDomain := extractDomain(target)
			if targetDomain == "" {
				log.Errorf("❌ Invalid target URL: %s", target)
				jsonError(w, "Configuration error", http.StatusInternalServerError)
				return
			}

			if !isDomainAllowed(targetDomain, allowedDomains) {
				log.Warnf("🚫 Domain not allowed: %s", targetDomain)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":           "domain_not_allowed",
					"message":         "Target domain is not in allowed list",
					"target_domain":   targetDomain,
					"allowed_domains": allowedDomains,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractDomain извлекает домен из URL
func extractDomain(urlStr string) string {
	if !strings.Contains(urlStr, "://") {
		urlStr = "https://" + urlStr
	}
	
	parts := strings.Split(urlStr, "://")
	if len(parts) < 2 {
		return ""
	}
	
	hostParts := strings.Split(parts[1], "/")
	host := hostParts[0]
	
	// Убираем порт если есть
	return strings.Split(host, ":")[0]
}

// isDomainAllowed проверяет наличие домена в списке
func isDomainAllowed(domain string, allowedDomains []string) bool {
	for _, allowed := range allowedDomains {
		if domain == allowed {
			return true
		}
		// Поддержка wildcard
		if strings.HasPrefix(allowed, "*.") {
			wildcardDomain := allowed[2:]
			if strings.HasSuffix(domain, wildcardDomain) {
				return true
			}
		}
	}
	return false
}

// formatDomains форматирует домены для JSON
func formatDomains(domains []string) string {
	return `["` + strings.Join(domains, `","`) + `"]`
}

func jsonError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   http.StatusText(statusCode),
		"message": message,
	})
}
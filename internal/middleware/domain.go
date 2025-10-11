// internal/middleware/domain.go
package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)

// ClientDomainValidator проверяет что клиент разрешен
func ClientDomainValidator(log logger.Logger, allowedDomains []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Если список доменов пустой - пропускаем все
			if len(allowedDomains) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Извлекаем идентификатор клиента
			clientIdentifier := extractClientIdentifier(r)
			clientType := getClientType(r)
			
			// Логируем для отладки
			log.Infof("🔍 Client check - Identifier: %s, Type: %s", clientIdentifier, clientType)

			// Проверяем разрешен ли клиент
			if !isClientAllowed(clientIdentifier, allowedDomains) {
				log.Warnf("🚫 Client not allowed: %s (allowed: %v)", clientIdentifier, allowedDomains)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":             "client_not_allowed",
					"message":           "Client is not in allowed list",
					"client_identifier": clientIdentifier,
					"client_type":       clientType,
					"allowed_clients":   allowedDomains,
				})
				return
			}

			log.Infof("✅ Client allowed: %s (%s)", clientIdentifier, clientType)
			next.ServeHTTP(w, r)
		})
	}
}

// extractClientIdentifier извлекает идентификатор клиента
func extractClientIdentifier(r *http.Request) string {
	// 1. Пробуем получить из Origin (для CORS запросов из браузера)
	if origin := r.Header.Get("Origin"); origin != "" {
		domain := extractDomainFromURL(origin)
		if domain != "" {
			return domain
		}
	}

	// 2. Пробуем получить из Referer
	if referer := r.Header.Get("Referer"); referer != "" {
		domain := extractDomainFromURL(referer)
		if domain != "" {
			return domain
		}
	}

	// 3. Пробуем получить из Host
	if host := r.Header.Get("Host"); host != "" {
		domain := strings.Split(host, ":")[0]
		if domain != "" {
			return domain
		}
	}

	// 4. Для локальных запросов используем "localhost"
	clientIP := extractClientIP(r)
	if isLocalRequest(clientIP) {
		return "localhost"
	}

	// 5. Для внешних запросов без домена используем IP
	return clientIP
}

// extractClientIP извлекает IP клиента
func extractClientIP(r *http.Request) string {
	// Пробуем получить из X-Real-IP
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Пробуем получить из X-Forwarded-For (если за прокси)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Берем первый IP из списка
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// Или используем RemoteAddr
	return extractIPFromRemoteAddr(r.RemoteAddr)
}

// extractIPFromRemoteAddr извлекает IP из RemoteAddr
func extractIPFromRemoteAddr(remoteAddr string) string {
	if remoteAddr == "" {
		return ""
	}

	// Убираем порт
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// Если нет порта, возвращаем как есть
		return remoteAddr
	}
	return host
}

// isLocalRequest проверяет локальный ли запрос
func isLocalRequest(ip string) bool {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return false
	}

	// Проверяем специальные локальные имена
	if ip == "localhost" {
		return true
	}

	// Парсим IP для проверки
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Проверяем приватные диапазоны
	if parsedIP.IsLoopback() {
		return true
	}
	if parsedIP.IsPrivate() {
		return true
	}
	if parsedIP.IsUnspecified() {
		return true
	}

	// Проверяем link-local адреса
	if parsedIP.IsLinkLocalUnicast() || parsedIP.IsLinkLocalMulticast() {
		return true
	}

	return false
}

// isClientAllowed проверяет разрешен ли клиент
func isClientAllowed(clientIdentifier string, allowedClients []string) bool {
	clientIdentifier = strings.TrimSpace(clientIdentifier)
	if clientIdentifier == "" {
		return false
	}

	for _, allowed := range allowedClients {
		allowed = strings.TrimSpace(allowed)
		if allowed == "" {
			continue
		}

		// Точное совпадение
		if clientIdentifier == allowed {
			return true
		}

		// Проверка по IP диапазону (например, "192.168.")
		if strings.HasSuffix(allowed, ".") && isIPAddress(clientIdentifier) {
			if strings.HasPrefix(clientIdentifier, allowed) {
				return true
			}
		}
		
		// Wildcard домены
		if strings.HasPrefix(allowed, "*.") {
			wildcardDomain := allowed[2:]
			if strings.HasSuffix(clientIdentifier, wildcardDomain) {
				return true
			}
		}
		
		// Локальные запросы
		if clientIdentifier == "localhost" && (allowed == "localhost" || allowed == "127.0.0.1" || allowed == "::1") {
			return true
		}
		
		// IPv6 localhost
		if clientIdentifier == "::1" && (allowed == "localhost" || allowed == "127.0.0.1" || allowed == "::1") {
			return true
		}

		// Проверка CIDR диапазона
		if isCIDRRange(allowed) && isIPAddress(clientIdentifier) {
			if isIPInCIDR(clientIdentifier, allowed) {
				return true
			}
		}
	}
	return false
}

// isIPAddress проверяет является ли строка IP адресом
func isIPAddress(ip string) bool {
	return net.ParseIP(ip) != nil
}

// isCIDRRange проверяет является ли строка CIDR диапазоном
func isCIDRRange(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

// isIPInCIDR проверяет входит ли IP в CIDR диапазон
func isIPInCIDR(ipStr, cidrStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	_, network, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false
	}

	return network.Contains(ip)
}

// getClientType определяет тип клиента
func getClientType(r *http.Request) string {
	if r.Header.Get("Origin") != "" {
		return "browser"
	}
	clientIP := extractClientIP(r)
	if isLocalRequest(clientIP) {
		return "local"
	}
	return "external"
}

// extractDomainFromURL извлекает домен из URL
func extractDomainFromURL(urlStr string) string {
	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" {
		return ""
	}
	
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

func jsonError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   http.StatusText(statusCode),
		"message": message,
	})
}
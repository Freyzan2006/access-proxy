package ratelimit

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)

type RateLimiter struct {
	mu          sync.Mutex
	requests    map[string][]time.Time
	limit       int
	window      time.Duration
	log         logger.Logger
}

func NewRateLimiter(limitPerMinute int, log logger.Logger) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limitPerMinute,
		window:   time.Minute,
		log:      log,
	}
}

func (rl *RateLimiter) Allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	
	// Очищаем старые запросы вне временного окна
	rl.cleanup(identifier, now)
	
	// Проверяем количество запросов
	if len(rl.requests[identifier]) >= rl.limit {
		return false
	}
	
	// Добавляем текущий запрос
	rl.requests[identifier] = append(rl.requests[identifier], now)
	return true
}

func (rl *RateLimiter) cleanup(identifier string, now time.Time) {
	validRequests := []time.Time{}
	
	for _, t := range rl.requests[identifier] {
		if now.Sub(t) <= rl.window {
			validRequests = append(validRequests, t)
		}
	}
	
	rl.requests[identifier] = validRequests
}

func (rl *RateLimiter) GetRemaining(identifier string) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.cleanup(identifier, time.Now())
	return rl.limit - len(rl.requests[identifier])
}

// Middleware возвращает HTTP middleware для ограничения запросов
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Используем IP адрес как идентификатор
		identifier := getClientIP(r)
		
		if !rl.Allow(identifier) {
			rl.log.Warnf("🚫 Rate limit exceeded for %s: %s %s", identifier, r.Method, r.URL.Path)
			
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
			
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{
				"error": "rate_limit_exceeded",
				"message": "Too many requests",
				"limit": "` + fmt.Sprintf("%d", rl.limit) + ` per minute",
				"retry_after": "60 seconds"
			}`))
			return
		}
		
		// Добавляем заголовки с информацией о лимите
		remaining := rl.GetRemaining(identifier)
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		
		rl.log.Infof("📊 Rate limit: %s has %d/%d requests remaining", identifier, remaining, rl.limit)
		next.ServeHTTP(w, r)
	})
}

// Вспомогательная функция для получения IP клиента
func getClientIP(r *http.Request) string {
	// Пробуем получить IP из X-Forwarded-For (если за прокси)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	
	// Или используем RemoteAddr
	return r.RemoteAddr
}

func (rl *RateLimiter) GetLimit() int {
	return rl.limit
}
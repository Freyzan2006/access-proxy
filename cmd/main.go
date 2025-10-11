// main.go
package main

import (
	"access-proxy/internal/config"
	"access-proxy/internal/server"
	"fmt"

	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)

func main() {
	cfg := config.LoadConfig()
	log := logger.New("access-proxy", logger.LevelInfo, logger.ModeProd)
	
	fmt.Printf("=== CONFIGURATION ===\n")
	fmt.Printf("Target: %s\n", cfg.Target)
	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("Rate Limit: %d/min\n", cfg.RateLimitPerMinute)
	fmt.Printf("Log Requests: %t\n", cfg.LogRequests)
	fmt.Printf("=====================\n")
	
	proxy := server.NewProxyServer(cfg.Target, log)
	
	// Передаем rate limit в HTTP сервер
	ser := server.NewHttpServer(proxy, cfg.Port, log, cfg.RateLimitPerMinute)

	ser.RegisterEndpoints()
	
	log.Infof("🚀 Proxy server starting: %s -> :%d", cfg.Target, cfg.Port)
	if cfg.RateLimitPerMinute > 0 {
		log.Infof("🔒 Rate limiting enabled: %d requests per minute", cfg.RateLimitPerMinute)
	}
	
	ser.ListenAndServe()
}
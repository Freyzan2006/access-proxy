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
	fmt.Printf("Allowed Domains: %v\n", cfg.AllowedDomains)
	fmt.Printf("=====================\n")
	
	proxy := server.NewProxyServer(cfg.Target, log)
	
	ser := server.NewHttpServer(
		proxy, 
		cfg.Port, 
		log, 
		cfg.RateLimitPerMinute,
		cfg.Target,
		cfg.LogRequests,
		cfg.AllowedDomains,
	)

	ser.RegisterEndpoints()
	
	log.Infof("🚀 Proxy server starting: %s -> :%d", cfg.Target, cfg.Port)
	if cfg.RateLimitPerMinute > 0 {
		log.Infof("🔒 Rate limiting enabled: %d requests per minute", cfg.RateLimitPerMinute)
	}
	if cfg.LogRequests {
		log.Info("📝 Request logging enabled")
	}
	if len(cfg.AllowedDomains) > 0 {
		log.Infof("🌐 Domain restrictions enabled: %v", cfg.AllowedDomains)
	}
	
	ser.ListenAndServe()
}
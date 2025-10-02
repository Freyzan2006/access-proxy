package config

import (
	"flag"
	"access-proxy/internal/types"
)

type flagConfig struct {
	Port int 
	AllowedDomains types.StringSlice 
	BlockedMethods types.StringSlice 
	RateLimitPerMinute int 
	LogRequests bool 
}


func newFlagConfig() *flagConfig {
	var cfg flagConfig

	// var (
	// 	allowedDomains types.StringSlice
	// 	blockedMethods []types.StringSlice
	// )

	// cfg.Port = *flag.Int("port", 8000, "Порт запуска proxy server");
	// cfg.AllowedDomains = *flag.Var(&cfg.AllowedDomains, "domains", "Разрешённые домены");
	// cfg.BlockedMethods = *flag.Var(&cfg.BlockedMethods, "blocks", "Запрещённые мыши");
	// cfg.RateLimitPerMinute = *flag.Int("rate", 100, "Лимит на запросы");
	// cfg.LogRequests = *flag.Bool("log", true, "Сохранять логи или нет");
	
	flag.Parse()

	return &flagConfig{
		cfg.Port,
		cfg.AllowedDomains,
		cfg.BlockedMethods,
		cfg.RateLimitPerMinute,
		cfg.LogRequests,
	}
}
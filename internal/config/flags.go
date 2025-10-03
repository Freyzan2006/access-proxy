package config

import (
	"flag"
	"access-proxy/internal/types"
)

type flagConfig struct {
	Config string
	Port int 
	AllowedDomains types.StringSlice 
	BlockedMethods types.StringSlice 
	RateLimitPerMinute int 
	LogRequests bool 
}

func newFlagConfig() *flagConfig {
    var cfg flagConfig

	config := flag.String("config", "config.yaml", "Путь к файлу для конфигурации")
    port := flag.Int("port", 8000, "Порт запуска proxy server")
    flag.Var(&cfg.AllowedDomains, "domains", "Разрешённые домены (значение можно указывать через запятую)")
	flag.Var(&cfg.BlockedMethods, "blocks", "Запрещённые методы (значение можно указывать через запятую)")
	rate := flag.Int("rate", 100, "Лимит на запросы")
    logRequests := flag.Bool("log", true, "Сохранять логи или нет")
   
    flag.Parse()
	
	cfg.Config = *config
    cfg.Port = *port
    cfg.RateLimitPerMinute = *rate
    cfg.LogRequests = *logRequests

    return &cfg
}
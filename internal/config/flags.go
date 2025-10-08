package config

import (
	"flag"
	"access-proxy/internal/types"
)

type flagRefs struct {
	port     *int
	rate     *int
	log      *bool
	domains  *types.StringSlice
	blocked  *types.StringSlice
}

func defineFlags(defaults *Config) *flagRefs {
	var allowed types.StringSlice
	var blocked types.StringSlice

	return &flagRefs{
		port:    flag.Int("port", defaults.Port, "Порт запуска proxy server"),
		rate:    flag.Int("rate", defaults.RateLimitPerMinute, "Лимит запросов в минуту"),
		log:     flag.Bool("log", defaults.LogRequests, "Логировать запросы"),
		domains: &allowed,
		blocked: &blocked,
	}
}

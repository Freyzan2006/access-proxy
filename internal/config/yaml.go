package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type yamlFileConfig struct {
	Target            string   `yaml:"target"`
	Port              int      `yaml:"port"`
	AllowedDomains    []string `yaml:"allowed_domains"`
	BlockedMethods    []string `yaml:"blocked_methods"`
	RateLimitPerMinute int     `yaml:"rate_limit_per_minute"`
	LogRequests       bool     `yaml:"log_requests"`
	Env               string   `yaml:"environment"`
}

// loadFromYAML читает конфиг из YAML и возвращает Config
func loadFromYAML(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("⚠️  Не удалось прочитать конфиг (%s): %v — используются дефолты.\n", path, err)
		return &Config{}
	}

	var yml yamlFileConfig
	if err := yaml.Unmarshal(data, &yml); err != nil {
		fmt.Printf("⚠️  Ошибка разбора YAML (%s): %v — используются дефолты.\n", path, err)
		return &Config{}
	}

	return &Config{
		Target:            yml.Target,
		Port:              yml.Port,
		AllowedDomains:    yml.AllowedDomains,
		BlockedMethods:    yml.BlockedMethods,
		RateLimitPerMinute: yml.RateLimitPerMinute,
		LogRequests:       yml.LogRequests,
		Env:               yml.Env,
	}
}

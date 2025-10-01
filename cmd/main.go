package main

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Port int `yaml:"port"`
	AllowedDomains []string `yaml:"allowed_domains"`
	BlockedMethods []string `yaml:"blocked_methods"`
	RateLimitPerMinute int `yaml:"rate_limit_per_minute"`
	LogRequests bool `yaml:"log_requests"`
}

func main() {
    // 1. Открываем файл
    data, err := os.ReadFile("config.yaml")
    if err != nil {
        panic(err)
    }

    // 2. Создаём переменную для конфигурации
    var cfg Config

    // 3. Парсим YAML → Go-структуру
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        panic(err)
    }

    // 4. Используем данные
    fmt.Println(cfg.Port);
	fmt.Println(cfg.AllowedDomains);
}
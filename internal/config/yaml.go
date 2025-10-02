package config

import (
	"os"
	"gopkg.in/yaml.v3"
)


type yamlConfig struct {
	Port int `yaml:"port"`
	AllowedDomains []string `yaml:"allowed_domains"`
	BlockedMethods []string `yaml:"blocked_methods"`
	RateLimitPerMinute int `yaml:"rate_limit_per_minute"`
	LogRequests bool `yaml:"log_requests"`
}



func newYamlConfig(path string) *yamlConfig {

	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}


	var cfg yamlConfig

	if err := yaml.Unmarshal(data, &cfg); err != nil {
        panic(err)
    }

	return &yamlConfig{
		cfg.Port,
		cfg.AllowedDomains,
		cfg.BlockedMethods,
		cfg.RateLimitPerMinute,
		cfg.LogRequests,
	}
}


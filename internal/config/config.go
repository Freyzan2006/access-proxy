package config

import (
	"flag"
)

type Config struct {
	Target             string
	Port               int
	AllowedDomains     []string
	BlockedMethods     []string
	RateLimitPerMinute int
	LogRequests        bool
}

func LoadConfig() *Config {
	configPath := flag.String("config", "config.yaml", "Путь к YAML конфигурации")
	

	flag.CommandLine.Parse([]string{}) 
	yamlCfg := loadFromYAML(*configPath)


	flagsRefs := defineFlags(yamlCfg)


	flag.Parse()


	final := mergeConfigs(yamlCfg, flagsRefs)

	
	return final
}

func mergeConfigs(yml *Config, flags *flagRefs) *Config {
	final := *yml

	if isFlagPassed("port") {
		final.Port = *flags.port
	}
	if isFlagPassed("rate") {
		final.RateLimitPerMinute = *flags.rate
	}
	if isFlagPassed("log") {
		final.LogRequests = *flags.log
	}
	if isFlagPassed("domains") {
		final.AllowedDomains = *flags.domains
	}
	if isFlagPassed("blocks") {
		final.BlockedMethods = *flags.blocked
	}
	if isFlagPassed("target") {
		final.Target = *flags.target
	}

	return &final
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

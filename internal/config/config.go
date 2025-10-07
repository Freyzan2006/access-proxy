package config

import (
	"flag"
	"fmt"
)

type Config struct {
	Port               int
	AllowedDomains     []string
	BlockedMethods     []string
	RateLimitPerMinute int
	LogRequests        bool
}

func LoadConfig() *Config {
	configPath := flag.String("config", "config.yaml", "Путь к YAML конфигурации")

	flag.CommandLine.Parse([]string{}) // "предварительный" парсинг
	yamlCfg := loadFromYAML(*configPath)

	// 3️⃣ объявляем остальные флаги с дефолтами
	flagsRefs := defineFlags(yamlCfg)

	// 4️⃣ теперь парсим настоящие аргументы
	flag.Parse()

	// 5️⃣ собираем конфиг, учитывая приоритет флагов
	final := mergeConfigs(yamlCfg, flagsRefs)

	fmt.Printf("✅ Конфигурация загружена: %+v\n", final)
	return final
}

// mergeConfigs теперь использует значения флагов после Parse()
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

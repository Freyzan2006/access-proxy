package main

import (
	"access-proxy/internal/config"
	"access-proxy/internal/proxy/server"

	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)


const Target = "https://jsonplaceholder.typicode.com"

func main() {
	cfg := config.LoadConfig();
	log := logger.New("access-proxy", logger.LevelDebug, "internal", logger.ModeDev)
	proxy := server.NewProxyServer(Target, log);
	ser := server.NewHttpServer(proxy, cfg.Port, log);

	ser.Endpoints();
	ser.ListenAndServe();
}

package main

import (
	"access-proxy/internal/proxy/server"


	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)


const Target = "https://jsonplaceholder.typicode.com"

func main() {
	log := logger.New("access-proxy.log", logger.LevelDebug, "internal", logger.ModeDev)
	proxy := server.NewProxyServer(Target, log);
	ser := server.NewHttpServer(proxy, 8080, log);

	ser.Endpoints();
	ser.ListenAndServe();
}

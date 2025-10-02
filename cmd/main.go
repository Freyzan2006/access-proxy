package main

import "fmt";
import "access-proxy/internal/config";

func main() {
    cfg := config.NewConfig("config.yaml");

    fmt.Println(cfg.Yaml.Port);
}
// package main

// import "fmt";
// import "access-proxy/internal/config";

// func main() {
//     cfg := config.NewConfig();

//     fmt.Printf("Port: %d\n", cfg.Yaml.Port);
//     fmt.Printf("Allowed Domains: %v\n", cfg.Yaml.AllowedDomains);
//     fmt.Printf("Blocked Methods: %v\n", cfg.Yaml.BlockedMethods);
//     fmt.Printf("Rate Limit: %d\n", cfg.Yaml.RateLimitPerMinute);
//     fmt.Printf("Log Requests: %v\n", cfg.Yaml.LogRequests);
// }


package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Target — адрес, на который прокси будет пересылать запросы
const Target = "https://jsonplaceholder.typicode.com"

func main() {
	// Парсим целевой адрес
	targetURL, err := url.Parse(Target)
	if err != nil {
		log.Fatalf("Ошибка разбора URL: %v", err)
	}

	// Создаём reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Дополнительно можно модифицировать запрос перед отправкой
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		// Например, можно переписать путь:
		// req.URL.Path = "/posts"
	}

	// Оборачиваем в handler с логированием
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("📡 %s %s", r.Method, r.URL.Path)
		proxy.ServeHTTP(w, r)
        
	})

	log.Println("🚀 Proxy сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

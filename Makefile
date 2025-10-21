dev: 
	go build -o bin/access-proxy ./cmd/main.go
	./bin/access-proxy
prod:
	docker run --network=host -p 8080:8080 access-proxy ./access-proxy -port 8080 -rate 10 -log true
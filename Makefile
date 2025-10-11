dev: 
	go run cmd/main.go

prod:
	docker run --network=host -p 8080:8080 access-proxy ./access-proxy -port 8080 -rate 10 -log true
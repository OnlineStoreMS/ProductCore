.PHONY: run build tidy swag docker-up docker-down

run:
	go run ./cmd/api -config configs/config.yaml

build:
	go build -o bin/productcore ./cmd/api

tidy:
	go mod tidy

# 需安装 swag: go install github.com/swaggo/swag/cmd/swag@latest
swag:
	swag init -g cmd/api/main.go -o docs/swagger --parseDependency

docker-up:
	docker compose -f deploy/docker-compose.dev.yml up -d

docker-down:
	docker compose -f deploy/docker-compose.dev.yml down

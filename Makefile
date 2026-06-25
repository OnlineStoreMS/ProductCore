.PHONY: run build tidy swag docker-up docker-down init-db

run:
	go run ./cmd/api -config configs/config.yaml

# 需安装 swag: go install github.com/swaggo/swag/cmd/swag@latest
SWAG_DIRS := cmd/api,admin,api,internal/pkg/response,internal/dto

swag:
	@command -v swag >/dev/null 2>&1 || (echo "请先安装: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
	swag init -g main.go -o docs/swagger --parseInternal --dir $(SWAG_DIRS)
	@chmod +x scripts/check_swag.sh 2>/dev/null || true
	@scripts/check_swag.sh

build: swag
	GOTMPDIR=.tmp go build -o bin/productcore ./cmd/api

tidy:
	GOTMPDIR=.tmp go mod tidy

# 创建 PostgreSQL 用户与数据库
init-db:
	@test -n "$(APP_PASSWORD)" || (echo "用法: make init-db APP_PASSWORD=你的密码 [SUDO=1]"; exit 1)
	chmod +x deploy/setup_db.sh
ifeq ($(SUDO),1)
	./deploy/setup_db.sh "$(APP_PASSWORD)" --sudo
else
	./deploy/setup_db.sh "$(APP_PASSWORD)" "$${PG_SUPERUSER:-postgres}"
endif

docker-up:
	docker compose -f deploy/docker-compose.dev.yml up -d

docker-down:
	docker compose -f deploy/docker-compose.dev.yml down

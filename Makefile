.PHONY: clean swag mockgen critic security test
.PHONY: docker.build docker.run docker.stop
.PHONY: dc.build dc.up dc.down

APP_NAME = smrv2-api
APP_VERSION = 0.0.1

clean:
	rm -rf ./build

swag:
	swag init --parseDependency --parseInternal --parseDepth=2 -g ./cmd/app/main.go

mockgen:
	sh ./bin/generate-mock.sh

critic:
	gocritic check -enableAll ./internal/domain/... ./internal/repository/... ./internal/service/... ./internal/delivery/http/... ./internal/builder/...

security:
	gosec ./...

test: clean critic security
	go test -v -timeout 180s -coverprofile=cover.out -cover ./internal/... ./test/...
	go tool cover -func=cover.out

docker.build:
	docker build -t $(APP_NAME):$(APP_VERSION) .

docker.run: docker.build
	docker run -d -p 3000:3000 --name $(APP_NAME) $(APP_NAME):$(APP_VERSION)

docker.stop:
	docker stop $(APP_NAME)
	docker rm $(APP_NAME)

dc.build:
	docker compose -f docker-compose.yml build

dc.up: dc.build
	docker compose up -d

dc.down:
	docker compose down
	
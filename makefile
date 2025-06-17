.PHONY: build run test clean proto docker

build:
	go build -o bin/main ./main.go
	go build -o bin/emailparser ./cmd/emailparser/main.go
	go build -o bin/rulesengine ./cmd/rulesengine/main.go
	go build -o bin/reportgenerator ./cmd/reportgenerator/main.go

proto:
	protoc \
      --go_out=. \
      --go-grpc_out=. \
      --go_opt=paths=source_relative \
      --go-grpc_opt=paths=source_relative \
      proto/*.proto


run-all:
	make run-emailparser &
	make run-rulesengine &
	make run-reportgenerator &
	make run-http

run-http:
	go run main.go

run-emailparser:
	go run cmd/emailparser/main.go

run-rulesengine:
	go run cmd/rulesengine/main.go

run-reportgenerator:
	go run cmd/reportgenerator/main.go

test:
	chmod +x test_api.sh
	./test_api.sh

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

clean:
	rm -rf bin/
	docker-compose down
	docker system prune -f

deps:
	go mod download
	go mod tidy

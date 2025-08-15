BINARY_NAME=overlock-mcp-server
DOCKER_IMAGE=overlock-mcp-server
DOCKER_TAG=latest

.PHONY: build test test-unit test-e2e run clean docker-build docker-run docker-test docker-test-unit docker-test-e2e docker-clean

build:
	@mkdir -p bin
	go build -o bin/$(BINARY_NAME) ./cmd/server

test: test-unit test-e2e

test-unit:
	go test ./pkg/... ./internal/... -v

test-e2e:
	go run github.com/onsi/ginkgo/v2/ginkgo run ./test/...

run: build
	./bin/$(BINARY_NAME)

clean:
	go clean
	rm -rf bin/

docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-build
	docker run -p 8080:8080 --rm $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose-up:
	docker-compose up --build

docker-compose-down:
	docker-compose down

docker-test: docker-test-unit docker-test-e2e

docker-test-unit:
	docker run --rm -v $(PWD):/app -w /app golang:1.24-alpine sh -c "go test ./pkg/... ./internal/... -v"

docker-test-e2e:
	docker run --rm -v $(PWD):/app -w /app golang:1.24-alpine sh -c "go run github.com/onsi/ginkgo/v2/ginkgo run ./test/..."

docker-clean:
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	docker system prune -f
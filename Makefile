.PHONY: fmt fmt-check test test-race vet build verify run docker-build docker-run compose-up compose-down

VERSION ?= dev

fmt:
	gofmt -w .

fmt-check:
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "Unformatted files:"; echo "$$out"; exit 1; fi

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

build:
	go build -ldflags="-s -w -X github.com/kirasync2748/github-widget/internal/version.Version=$(VERSION)" -o bin/server ./cmd/server

verify: fmt-check vet test test-race build
	@echo "All checks passed."

run:
	go run ./cmd/server

docker-build:
	docker build --build-arg VERSION=$(VERSION) -t github-widget:latest .

docker-run:
	docker run --rm -p 3000:3000 github-widget:latest

compose-up:
	docker compose -f docker-compose.yml up -d --build

compose-down:
	docker compose -f docker-compose.yml down

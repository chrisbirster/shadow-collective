.PHONY: test build run

test:
	go test ./...
	go vet ./...

build:
	go build ./cmd/shadow-collective

run:
	go run ./cmd/shadow-collective -config config/services.json

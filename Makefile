APP=cdn-cert-sync

.PHONY: fmt test build run

fmt:
	gofmt -w ./cmd ./internal

test:
	go test ./...

build:
	go build ./...

run:
	go run ./cmd/$(APP) --config ./configs/config.example.yaml

APP=cdn-cert-sync

.PHONY: fmt test test-integration build run

fmt:
	gofmt -w ./cmd ./internal

test:
	go test ./...

test-integration:
	go test -tags integration ./internal/aliyun

build:
	go build ./...

run:
	go run ./cmd/$(APP) --config ./configs/config.example.yaml

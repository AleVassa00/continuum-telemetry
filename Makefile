APP_CONFIG ?= config/local.yml

.PHONY: fmt tidy test check run-edge run-sensor run-worker run-api

fmt:
	go fmt ./...

tidy:
	go mod tidy

test:
	go test ./...

check: fmt tidy test

run-edge:
	APP_CONFIG=$(APP_CONFIG) go run ./cmd/edge-gateway

run-sensor:
	APP_CONFIG=$(APP_CONFIG) go run ./cmd/sensor-simulator

run-worker:
	APP_CONFIG=$(APP_CONFIG) go run ./cmd/cloud-worker

run-api:
	APP_CONFIG=$(APP_CONFIG) go run ./cmd/cloud-api
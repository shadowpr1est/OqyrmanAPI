.PHONY: generate run build test

## Regenerate Swagger docs and run the server
run: generate
	go run ./cmd/app

## Generate Swagger documentation
generate:
	swag init -g cmd/app/main.go -o docs

## Build binary
build: generate
	go build -o app.exe ./cmd/app

## Run tests
test:
	go test ./...

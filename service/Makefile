default: build

build:
	go build -o DataPipes

test:
	go test ./...

lint:
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed on:"; gofmt -l .; exit 1)
	go vet ./...

run:
	go run .

.PHONY: default build test lint run

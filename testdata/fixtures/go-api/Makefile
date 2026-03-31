.PHONY: build test lint run clean

build:
	go build -o bin/server ./cmd/server

test:
	go test ./... -v -race

lint:
	golangci-lint run ./...

run: build
	./bin/server

clean:
	rm -rf bin/

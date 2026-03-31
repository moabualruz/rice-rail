BINARY := rice-rail
MODULE := github.com/mkh/rice-railing
BUILD_DIR := bin
MAIN := ./cmd/rice-rail

.PHONY: build run test lint fmt clean install

build:
	go build -o $(BUILD_DIR)/$(BINARY) $(MAIN)

run: build
	./$(BUILD_DIR)/$(BINARY) $(ARGS)

test:
	go test ./... -v

test-short:
	go test ./... -short

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	goimports -w .

clean:
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(BINARY) $(GOPATH)/bin/$(BINARY)

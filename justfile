ci:
    make build
    go vet ./...
    go test ./... -count=1

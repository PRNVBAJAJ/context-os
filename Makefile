.PHONY: build test lint vet clean snapshot

build:
	go build -o bin/context ./cmd/context/...

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/ dist/

# Build a local snapshot release (requires goreleaser)
snapshot:
	goreleaser build --snapshot --clean

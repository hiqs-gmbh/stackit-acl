.PHONY: build test lint snapshot clean

build:
	go build -o bin/stackit-acl .

test:
	go test ./...

lint:
	golangci-lint run

snapshot:
	goreleaser build --snapshot --clean

clean:
	rm -rf bin/ dist/

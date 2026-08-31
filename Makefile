.PHONY: build test lint clean

build:
	go build -o bin/stackit-acl .

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/

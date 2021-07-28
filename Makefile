version := v1.0.0

format:
		goimports -w -l .
		go fmt
		gofumpt -w .

check:
		golangci-lint run

test:
		go test -coverprofile=coverage.txt -covermode=atomic

build: format check test


.PHONY: build test lint tidy snapshot clean

STATICCHECK ?= go run honnef.co/go/tools/cmd/staticcheck@latest

build:
	go build ./cmd/tea-eyes

test:
	go test ./...

lint:
	go vet ./...
	$(STATICCHECK) ./...

tidy:
	go mod tidy

snapshot:
	goreleaser release --snapshot --clean

clean:
	rm -rf dist/ test/golden/*.actual

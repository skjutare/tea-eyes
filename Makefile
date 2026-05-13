.PHONY: build test test-integration lint tidy snapshot release demo-render clean

GOLANGCI_LINT ?= golangci-lint
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X gitlab.com/skjutare/tea-eyes/internal/server.Version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" ./cmd/tea-eyes

test:
	go test ./...

# Integration tests rely on external binaries (vhs, ttyd, ffmpeg). Guarded by
# the `integration` build tag so CI can opt out by default.
test-integration:
	go test -tags=integration ./test/integration/...

lint:
	$(GOLANGCI_LINT) fmt
	$(GOLANGCI_LINT) run --fix

tidy:
	go mod tidy

snapshot:
	goreleaser release --snapshot --clean

# Real release flow run locally. Requires GITHUB_TOKEN / GITLAB_TOKEN env vars.
release:
	goreleaser release --clean

# Regenerate the committed demo PNG used in the README. Requires vhs/ttyd/ffmpeg.
demo-render: build
	@mkdir -p docs/img dist/demo
	go build -o dist/demo/multi-pane ./examples/multi-pane
	./tea-eyes cache clean >/dev/null || true
	scripts/demo-render.sh dist/demo/multi-pane docs/img/multi-pane-demo.png

clean:
	rm -rf dist/ test/golden/*.actual

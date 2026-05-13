.PHONY: build test test-integration lint tidy snapshot demo-render clean

STATICCHECK ?= go run honnef.co/go/tools/cmd/staticcheck@latest

build:
	go build ./cmd/tea-eyes

test:
	go test ./...

# Integration tests rely on external binaries (vhs, ttyd, ffmpeg). Guarded by
# the `integration` build tag so CI can opt out by default.
test-integration:
	go test -tags=integration ./test/integration/...

lint:
	go vet ./...
	$(STATICCHECK) ./...

tidy:
	go mod tidy

snapshot:
	goreleaser release --snapshot --clean

# Regenerate the committed demo PNG used in the README. Requires vhs/ttyd/ffmpeg.
demo-render: build
	@mkdir -p docs/img dist/demo
	go build -o dist/demo/multi-pane ./examples/multi-pane
	./tea-eyes cache clean >/dev/null || true
	scripts/demo-render.sh dist/demo/multi-pane docs/img/multi-pane-demo.png

clean:
	rm -rf dist/ test/golden/*.actual

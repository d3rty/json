# Variables
GOLANGCI_LINT := $(shell which golangci-lint)

# Default target
all: tidy

# Tidy: format and vet the code
tidy:
	@go fmt $$(go list ./...)
	@go vet $$(go list ./...)
	@go mod tidy

# Install golangci-lint only if it's not already installed
lint-install:
	@if ! [ -x "$(GOLANGCI_LINT)" ]; then \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi

# Lint the code using golangci-lint
lint: lint-install
	$(shell which golangci-lint) fmt
	$(shell which golangci-lint) run

test:
	go test ./...

# Generate HTML form for your Config
.PHONY: config-schema.json
config-schema.json:
	@echo "🔧 Generating config-schema.json…"
	@go run ./cmd/config-schema > demo/config-schema.json


wasm:
	@echo "Building WASM module..."
	GOOS=js GOARCH=wasm go build -o demo/main.wasm ./cmd/wasm-demo
	@echo "Installing wasm_exec.js…"
	@GOROOT=$$(go env GOROOT); \
	cp $$GOROOT/lib/wasm/wasm_exec.js demo/; \


demoserve:
	@echo "📡  Serving demo/ at http://localhost:8080"
	cd demo && python3 -m http.server 8080

# Phony targets
.PHONY: all tidy lint-install lint test wasm demoserve

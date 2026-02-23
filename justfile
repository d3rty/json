# Variables
golangci_lint := `which golangci-lint || echo ""`

# Default target
default: tidy

# Tidy: format and vet the code
tidy:
    @go fmt $(go list ./...)
    @go vet $(go list ./...)
    @go mod tidy

# Tidy WASM code: format and vet with js/wasm build tags
wasm-tidy:
    @echo "🔧 Formatting and vetting WASM code..."
    @go fmt ./cmd/wasm-demo
    @GOOS=js GOARCH=wasm go vet ./cmd/wasm-demo

# Install golangci-lint only if it's not already installed
lint-install:
    @if ! command -v golangci-lint &> /dev/null; then \
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
    fi

# Lint the code using golangci-lint
lint: lint-install
    golangci-lint fmt
    golangci-lint run

# Run tests
test:
    go test ./...

# Generate config-schema.json
config-schema:
    @echo "🔧 Generating config-schema.json…"
    @go run ./cmd/config-schema > demo/config-schema.json

# Build WASM module
wasm:
    @echo "Building WASM module..."
    GOOS=js GOARCH=wasm go build -o demo/main.wasm ./cmd/wasm-demo
    @echo "Installing wasm_exec.js…"
    @cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" demo/

# Install demo dependencies
demo-install:
    @echo "📦 Installing demo dependencies..."
    cd demo && pnpm install

# Run ESLint on demo
demo-eslint:
    @echo "🔍 Running ESLint on demo..."
    cd demo && pnpm run lint

# Run ESLint with auto-fix on demo
demo-eslint-fix:
    @echo "🔧 Running ESLint with auto-fix on demo..."
    cd demo && pnpm run lint:fix

# Clean demo dependencies
demo-clean:
    @echo "🧹 Cleaning demo dependencies..."
    cd demo && rm -rf node_modules pnpm-lock.yaml

# Serve demo locally
demo-serve:
    @echo "📡  Serving demo/ at http://localhost:8080"
    cd demo && python3 -m http.server 8080

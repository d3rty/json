.PHONY: help install eslint eslint-fix clean

# Default target
help:
	@echo "Available targets:"
	@echo "  make install     - Install pnpm dependencies"
	@echo "  make eslint      - Run ESLint"
	@echo "  make eslint-fix  - Run ESLint with auto-fix"
	@echo "  make clean       - Remove node_modules"

# Install dependencies
install:
	pnpm install

# Run ESLint
eslint:
	pnpm run lint

# Run ESLint with auto-fix
eslint-fix:
	pnpm run lint:fix

# Clean dependencies
clean:
	rm -rf node_modules pnpm-lock.yaml

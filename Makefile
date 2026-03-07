SHELL := /bin/sh

APP_NAME := ix-sustainment-os
CMD_PATH := ./cmd/ix-sustainment-os
BIN_DIR := ./bin
WEB_DIR := ./web
API_SPEC := ./api/openapi.yaml
SCHEMA_DIR := ./schemas

GO ?= go
NODE ?= node
NPM ?= npm

.PHONY: help
help:
	@echo "IX Sustainment OS — common targets"
	@echo ""
	@echo "Build and run:"
	@echo "  make build              Build Go binary"
	@echo "  make run                Run backend locally"
	@echo "  make clean              Remove local build artifacts"
	@echo ""
	@echo "Go quality:"
	@echo "  make fmt                Format Go files"
	@echo "  make vet                Run go vet"
	@echo "  make test               Run Go tests"
	@echo "  make lint               Run repo lint checks"
	@echo ""
	@echo "Frontend:"
	@echo "  make web-install        Install frontend dependencies"
	@echo "  make web-dev            Start frontend dev server"
	@echo "  make web-build          Build frontend"
	@echo "  make web-test           Run frontend tests"
	@echo ""
	@echo "API and schema:"
	@echo "  make api-check          Validate OpenAPI spec"
	@echo "  make schema-check       Validate JSON schemas"
	@echo ""
	@echo "Repo:"
	@echo "  make docs-check         Run documentation consistency checks"
	@echo "  make check              Run core local quality gates"

.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(APP_NAME) $(CMD_PATH)

.PHONY: run
run:
	$(GO) run $(CMD_PATH)

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
	rm -rf $(WEB_DIR)/dist
	rm -rf $(WEB_DIR)/build
	rm -rf coverage
	rm -f coverage.out profile.out

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: test
test:
	$(GO) test ./... -coverprofile=coverage.out

.PHONY: lint
lint: fmt vet docs-check

.PHONY: check
check: lint test api-check schema-check

.PHONY: docs-check
docs-check:
	@echo "Checking required documentation files..."
	@test -f README.md || (echo "Missing README.md" && exit 1)
	@test -f LICENSE.md || (echo "Missing LICENSE.md" && exit 1)
	@test -f COMMERCIAL_TERMS.md || (echo "Missing COMMERCIAL_TERMS.md" && exit 1)
	@echo "Documentation presence checks passed."

.PHONY: api-check
api-check:
	@if [ -f "$(API_SPEC)" ]; then \
		echo "Validating OpenAPI spec at $(API_SPEC)..."; \
		if command -v npx >/dev/null 2>&1; then \
			npx @redocly/cli lint $(API_SPEC); \
		else \
			echo "npx not found. Install Node.js to lint OpenAPI locally."; \
			exit 1; \
		fi; \
	else \
		echo "API spec not found at $(API_SPEC)"; \
		exit 1; \
	fi

.PHONY: schema-check
schema-check:
	@if [ -d "$(SCHEMA_DIR)" ]; then \
		echo "Checking JSON schema directory: $(SCHEMA_DIR)"; \
		find $(SCHEMA_DIR) -type f -name "*.json" | grep -q . || (echo "No schema files found." && exit 1); \
		echo "Schema directory contains JSON files."; \
	else \
		echo "Schema directory not found at $(SCHEMA_DIR)"; \
		exit 1; \
	fi

.PHONY: web-install
web-install:
	@if [ -d "$(WEB_DIR)" ]; then \
		cd $(WEB_DIR) && $(NPM) install; \
	else \
		echo "Web directory not found at $(WEB_DIR)"; \
		exit 1; \
	fi

.PHONY: web-dev
web-dev:
	@if [ -d "$(WEB_DIR)" ]; then \
		cd $(WEB_DIR) && $(NPM) run dev; \
	else \
		echo "Web directory not found at $(WEB_DIR)"; \
		exit 1; \
	fi

.PHONY: web-build
web-build:
	@if [ -d "$(WEB_DIR)" ]; then \
		cd $(WEB_DIR) && $(NPM) run build; \
	else \
		echo "Web directory not found at $(WEB_DIR)"; \
		exit 1; \
	fi

.PHONY: web-test
web-test:
	@if [ -d "$(WEB_DIR)" ]; then \
		cd $(WEB_DIR) && $(NPM) test; \
	else \
		echo "Web directory not found at $(WEB_DIR)"; \
		exit 1; \
	fi

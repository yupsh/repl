# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
#
# Quality gate for the yupsh REPL. `make check` must pass with zero findings
# before any change is considered complete. Tools are declared in the go.mod
# `tool` stanza and run via `go tool` (no global installs); releases are driven
# by goreleaser from the `.goreleaser.yaml` config.
.DEFAULT_GOAL := build

.PHONY: help
help: ## Show this help
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Project variables
NAME      ?= yupsh
BUILD_DIR ?= bin
GO        ?= go

# Packages checked by vet and staticcheck (library plus the main shim).
PKGS       = . ./yupsh
# The library package, whose statement coverage must be 100%. The ./yupsh main
# package is pure os/signal/term wiring exercised through the library, so it is
# excluded from the coverage target (consistent with the framework's own gate).
COVERPKGS  = .
# Production (non-test) Go files, used by the cognitive-complexity gate.
GO_FILES  := $(shell find . -type f -name '*.go' ! -name '*_test.go')

GOOS      ?= $(shell $(GO) env GOOS)
GOARCH    ?= $(shell $(GO) env GOARCH)
DIST_BINARY = dist/$(NAME)-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe,)
BIN_NAME    = $(NAME)-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe,)
BIN_TARGET  = $(BUILD_DIR)/$(BIN_NAME)

## Build

$(BUILD_DIR):
	@mkdir -p $@

$(DIST_BINARY): $(GO_FILES) .goreleaser.yaml
	$(GO) tool goreleaser build --single-target --snapshot --clean

$(BIN_TARGET): $(BUILD_DIR) $(DIST_BINARY)
	cp $(DIST_BINARY) $@
	ln -sf $(BIN_NAME) $(BUILD_DIR)/$(NAME)

.PHONY: build
build: $(BIN_TARGET) ## Build binary for the current platform only

.PHONY: build-all
build-all: ## Build binaries for all release platforms (snapshot)
	$(GO) tool goreleaser build --snapshot --clean

.PHONY: release
release: ## Create a release with goreleaser (requires a git tag)
	$(GO) tool goreleaser release --clean

.PHONY: release-snapshot
release-snapshot: ## Create a snapshot release (no git tag required)
	$(GO) tool goreleaser release --snapshot --clean

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)/$(NAME)*
	rm -rf dist/
	rm -f cover.out

## Code Quality

.PHONY: check
check: fmt-check vet staticcheck cognit cover goreleaser-check vuln ## Full quality gate: gofumpt, vet, staticcheck, complexity<=7, 100% coverage, goreleaser config, vuln scan

.PHONY: fmt
fmt: ## Rewrite all files with the strict formatter
	$(GO) tool gofumpt -w .

.PHONY: fmt-check
fmt-check: ## Fail if any file is not gofumpt-clean
	@out="$$($(GO) tool gofumpt -l .)"; \
	if [ -n "$$out" ]; then echo "gofumpt findings:"; echo "$$out"; exit 1; fi

.PHONY: vet
vet: ## Run go vet
	$(GO) vet $(PKGS)

.PHONY: staticcheck
staticcheck: ## Run staticcheck static analysis (zero findings)
	$(GO) tool staticcheck $(PKGS)

.PHONY: cognit
cognit: ## Assert cognitive complexity <= 7 for every production function
	$(GO) tool gocognit -over 7 $(GO_FILES)

.PHONY: goreleaser-check
goreleaser-check: ## Validate the goreleaser release configuration
	$(GO) tool goreleaser check

.PHONY: vuln
vuln: ## Scan dependencies for known vulnerabilities
	$(GO) tool govulncheck ./...

.PHONY: cover
cover: ## Run tests and assert 100.0% statement coverage of the library
	@$(GO) test -coverprofile=cover.out $(COVERPKGS) >/dev/null
	@total="$$($(GO) tool cover -func=cover.out | awk '/^total:/ {print $$3}')"; \
	echo "coverage: $$total"; \
	if [ "$$total" != "100.0%" ]; then \
		echo "FAIL: coverage $$total is below 100.0%"; \
		$(GO) tool cover -func=cover.out | awk '$$3 != "100.0%"'; \
		exit 1; \
	fi

## Test

.PHONY: test
test: ## Run the unit tests
	$(GO) test $(COVERPKGS)

.PHONY: integration
integration: ## Run opt-in black-box integration tests (builds the real binary)
	$(GO) test -tags integration ./...

## Utilities

.PHONY: tidy
tidy: ## Tidy and verify module dependencies
	$(GO) mod tidy
	$(GO) mod verify

GO ?= go
RAGEL ?= ragel
RAGEL_VERSION ?= 6.10
GOYACC_VERSION ?= v0.47.0
GOBIN ?= $(shell $(GO) env GOBIN)
GOPATH ?= $(shell $(GO) env GOPATH)
GO_BIN_DIR ?= $(if $(GOBIN),$(GOBIN),$(GOPATH)/bin)
GOYACC ?= $(GO_BIN_DIR)/goyacc
GOYACC_REPORT ?= /dev/null

.PHONY: install-tools install-ragel install-goyacc check-tools check-ragel check-goyacc generate test vet

install-tools: install-ragel install-goyacc check-tools

install-ragel:
	@if command -v "$(RAGEL)" >/dev/null 2>&1; then \
		echo "ragel already installed: $$(command -v "$(RAGEL)")"; \
	elif command -v brew >/dev/null 2>&1; then \
		brew install ragel; \
	elif command -v apt-get >/dev/null 2>&1; then \
		sudo apt-get update && sudo apt-get install -y ragel; \
	elif command -v dnf >/dev/null 2>&1; then \
		sudo dnf install -y ragel; \
	elif command -v yum >/dev/null 2>&1; then \
		sudo yum install -y ragel; \
	else \
		echo "ragel is not installed. Install it with your system package manager."; \
		exit 1; \
	fi
	@$(MAKE) check-ragel

install-goyacc:
	$(GO) install golang.org/x/tools/cmd/goyacc@$(GOYACC_VERSION)
	@$(MAKE) check-goyacc

check-tools: check-ragel check-goyacc

check-ragel:
	@version="$$( $(RAGEL) -v 2>&1 | sed -n 's/.*version \([0-9][^ ]*\).*/\1/p' | head -n 1 )"; \
	if [ "$$version" != "$(RAGEL_VERSION)" ]; then \
		echo "expected ragel $(RAGEL_VERSION), got $${version:-unknown}"; \
		exit 1; \
	fi

check-goyacc:
	@test -x "$(GOYACC)" || { echo "goyacc not found at $(GOYACC). Run 'make install-goyacc'."; exit 1; }
	@version="$$( $(GO) version -m "$(GOYACC)" 2>/dev/null | awk '$$1 == "mod" && $$2 == "golang.org/x/tools" { print $$3 }' )"; \
	if [ "$$version" != "$(GOYACC_VERSION)" ]; then \
		echo "expected goyacc from golang.org/x/tools $(GOYACC_VERSION), got $${version:-unknown}"; \
		exit 1; \
	fi

generate:
	$(RAGEL) -Z lexer.rl
	$(GOYACC) -v $(GOYACC_REPORT) -o parser.go parser.y
	gofmt -w parser.go lexer.go

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

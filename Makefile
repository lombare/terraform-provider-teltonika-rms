BINARY  := terraform-provider-teltonika-rms
VERSION := 1.1.0
HOSTNAME := registry.terraform.io
NAMESPACE := lombare
NAME     := teltonika-rms
OS_ARCH  := $(shell go env GOOS)_$(shell go env GOARCH)

TFPLUGINDOCS := go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest

.PHONY: build install test tidy fmt vet docs docs-check

build:
	go build -o $(BINARY)

install: build
	mkdir -p ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)
	mv $(BINARY) ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)/

test:
	go test ./... -v

tidy:
	go mod tidy

fmt:
	gofmt -s -w .

vet:
	go vet ./...

# Regenerate the docs/ tree from the provider's live schemas + examples/.
docs:
	$(TFPLUGINDOCS) generate --provider-name teltonika --rendered-provider-name "Teltonika RMS"

# CI-friendly: fail if docs/ is out of date.
docs-check:
	$(TFPLUGINDOCS) generate --provider-name teltonika --rendered-provider-name "Teltonika RMS"
	@git diff --exit-code docs/ || (echo "docs/ is stale — run 'make docs' and commit"; exit 1)

BINARY  := terraform-provider-teltonika-rms
VERSION := 0.1.0
HOSTNAME := registry.terraform.io
NAMESPACE := lombare
NAME     := teltonika-rms
OS_ARCH  := $(shell go env GOOS)_$(shell go env GOARCH)

.PHONY: build install test tidy fmt vet

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

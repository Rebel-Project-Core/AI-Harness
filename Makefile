BINARY_NAME=ai-harness
GO_FILES=$(shell find . -name '*.go')
CONTAINER_TOOL ?= docker

# Detect architecture
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M), x86_64)
	CREDO_ARCH ?= amd64
else ifeq ($(UNAME_M), arm64)
	CREDO_ARCH ?= arm64
else ifeq ($(UNAME_M), aarch64)
	CREDO_ARCH ?= arm64
else
	CREDO_ARCH ?= amd64
endif

CREDO_IMAGE ?= ghcr.io/rebel-project-core/core:latest-$(CREDO_ARCH)

.PHONY: all build test clean run vet image docker

all: build

build:
	go build -o $(BINARY_NAME) ./cmd/ai-harness

test:
	go test -v ./...

clean:
	rm -f $(BINARY_NAME)

vet:
	go vet ./...

run: build
	./$(BINARY_NAME)

image: docker

docker:
	$(CONTAINER_TOOL) build --build-arg CREDO_IMAGE=$(CREDO_IMAGE) --build-arg GOARCH=$(CREDO_ARCH) -t ai-harness:latest .

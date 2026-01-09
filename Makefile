BINARY_NAME=ai-harness
GO_FILES=$(shell find . -name '*.go')

.PHONY: all build test clean run vet

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

docker:
	docker build -t ai-harness:latest .

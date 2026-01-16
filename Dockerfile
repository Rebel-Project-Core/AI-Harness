ARG CREDO_IMAGE=ghcr.io/rebel-project-core/core:latest-amd64
ARG GOARCH=amd64

FROM golang:latest AS builder
ARG GOARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build -o ai-harness ./cmd/ai-harness

FROM ${CREDO_IMAGE}

COPY --from=builder /app/ai-harness /usr/local/bin/ai-harness

ENTRYPOINT ["ai-harness"]

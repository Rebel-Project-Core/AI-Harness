FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ai-harness ./cmd/ai-harness

ARG CREDO_IMAGE=ghcr.io/rebel-project-core/core:latest-amd64
FROM ${CREDO_IMAGE}

COPY --from=builder /app/ai-harness /usr/local/bin/ai-harness

ENTRYPOINT ["ai-harness"]

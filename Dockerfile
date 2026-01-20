ARG CREDO_ARCH=amd64

FROM golang:latest AS builder
ARG CREDO_ARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${CREDO_ARCH} go build -o ai-harness ./cmd/ai-harness

FROM ghcr.io/rebel-project-core/rebel:latest

COPY --from=builder /app/ai-harness /usr/local/bin/ai-harness

ENTRYPOINT ["ai-harness"]

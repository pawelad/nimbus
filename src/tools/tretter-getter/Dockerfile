FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build the binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o tretter-getter .

FROM alpine:3.21 AS runtime

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary and web assets from builder
COPY --from=builder /app/tretter-getter .

# Expose port (Echo)
EXPOSE 1945

# Configure volumes
VOLUME ["/app/downloads", "/app/data"]

ENTRYPOINT ["/app/tretter-getter"]

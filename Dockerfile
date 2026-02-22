# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary, -trimpath for smaller binaries
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o jvs ./cmd/jvs

# Final stage
FROM alpine:3.20

# Install runtime dependencies
# - ca-certificates: for HTTPS connections
# - bash: for better shell scripting support
# - coreutils: for common utilities (chmod, chown, etc.)
RUN apk add --no-cache \
    ca-certificates \
    bash \
    coreutils

# Create a non-root user
RUN addgroup -g 1000 jvs && \
    adduser -D -u 1000 -G jvs jvs

# Set working directory
WORKDIR /workspace

# Copy the binary from builder
COPY --from=builder /build/jvs /usr/local/bin/jvs

# Change ownership to jvs user
RUN chown -R jvs:jvs /workspace && \
    chmod +x /usr/local/bin/jvs

# Switch to non-root user
USER jvs

# Set the default command
ENTRYPOINT ["/usr/local/bin/jvs"]
CMD ["--help"]

# Labels for metadata
LABEL org.opencontainers.image.title="JVS (Juicy Versioned Workspaces)"
LABEL org.opencontainers.image.description="Workspace versioning system built on JuiceFS"
LABEL org.opencontainers.image.vendor="JVS Project"
LABEL org.opencontainers.image.source="https://github.com/jvs-project/jvs"
LABEL org.opencontainers.image.licenses="MIT"

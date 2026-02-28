# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies including FUSE for JuiceFS support
RUN apk add --no-cache \
    git \
    ca-certificates \
    fuse3-dev \
    gcc \
    musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# Enable CGO for FUSE support, -trimpath for smaller binaries
RUN CGO_ENABLED=1 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -linkmode=external" \
    -tags "fuse" \
    -o jvs ./cmd/jvs

# Final stage
FROM alpine:3.19

# Install runtime dependencies
# - ca-certificates: for HTTPS connections
# - bash: for better shell scripting support and completion
# - coreutils: for common utilities (chmod, chown, etc.)
# - fuse3: for FUSE mount support (required for JuiceFS)
# - juicefs-cmd: JuiceFS client (optional, for mount operations)
RUN apk add --no-cache \
    ca-certificates \
    bash \
    coreutils \
    fuse3 \
    curl \
    tzdata \
    && curl -L https://github.com/juicedata/juicefs/releases/download/v1.2.2/juicefs-1.2.2-linux-amd64.tar.gz \
    | tar -xz -C /tmp \
    && mv /tmp/juicefs /usr/local/bin/juicefs \
    && chmod +x /usr/local/bin/juicefs \
    && rm -rf /tmp/juicefs*

# Install shell completion
RUN mkdir -p /usr/share/bash-completion/completions /usr/share/zsh/site-functions

# Create a non-root user (with access to FUSE devices)
RUN addgroup -g 1000 jvs && \
    adduser -D -u 1000 -G jvs jvs

# Set working directory
WORKDIR /workspace

# Copy the binary from builder
COPY --from=builder /build/jvs /usr/local/bin/jvs

# Copy completion scripts
COPY --from=builder /build/scripts/completion/jvs.bash /usr/share/bash-completion/completions/jvs 2>/dev/null || true
COPY --from=builder /build/scripts/completion/jvs.zsh /usr/share/zsh/site-functions/_jvs 2>/dev/null || true

# Change ownership to jvs user
RUN chown -R jvs:jvs /workspace && \
    chmod +x /usr/local/bin/jvs

# Enable bash completion in the shell
RUN echo 'source /usr/share/bash-completion/completions/jvs 2>/dev/null || true' >> /etc/bashrc

# Switch to non-root user
USER jvs

# Set the default command
ENTRYPOINT ["/usr/local/bin/jvs"]
CMD ["--help"]

# Labels for metadata
LABEL org.opencontainers.image.title="JVS (Juicy Versioned Workspaces)"
LABEL org.opencontainers.image.description="Workspace versioning system built on JuiceFS with FUSE support"
LABEL org.opencontainers.image.vendor="JVS Project"
LABEL org.opencontainers.image.source="https://github.com/jvs-project/jvs"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.version="v1.0.0"
LABEL org.opencontainers.image.documentation="https://github.com/jvs-project/jvs/blob/main/docs/DOCKER.md"

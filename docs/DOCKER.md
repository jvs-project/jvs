# Docker Image for JVS

This document describes how to use the JVS Docker image.

## Quick Start

### Using Docker Run

```bash
# Pull the latest image
docker pull ghcr.io/jvs-project/jvs:latest

# Initialize a new repository
docker run --rm -v "$(pwd)/workspace:/workspace" ghcr.io/jvs-project/jvs init my-repo

# Create a snapshot
docker run --rm -v "$(pwd)/workspace:/workspace" ghcr.io/jvs-project/jvs snapshot "initial"

# List snapshots
docker run --rm -v "$(pwd)/workspace:/workspace" ghcr.io/jvs-project/jvs history
```

### Using Docker Compose

```bash
# Start the service
docker-compose up -d

# Run JVS commands
docker-compose run jvs init my-repo
docker-compose run jvs snapshot "checkpoint"

# Stop the service
docker-compose down
```

## Image Tags

- `ghcr.io/jvs-project/jvs:latest` - Latest stable release
- `ghcr.io/jvs-project/jvs:v1.0.0` - Versioned releases
- `ghcr.io/jvs-project/jvs:v1` - Major version
- `ghcr.io/jvs-project/jvs:main` - Latest main branch build

## Platforms

The Docker image is built for multiple platforms:
- `linux/amd64` - x86_64 / AMD64
- `linux/arm64` - ARM64 / aarch64

## Volume Mounts

The JVS Docker image expects a workspace volume to be mounted:

```bash
docker run -v /path/to/workspace:/workspace ghcr.io/jvs-project/jvs <command>
```

The `.jvs` directory will be created inside the workspace for storing:
- Snapshots
- Descriptors
- Metadata
- Configuration

## Configuration

### Environment Variables

- `JVS_ENGINE` - Default snapshot engine (copy, reflink-copy, juicefs-clone)

### Config File

You can mount a custom config file:

```bash
docker run -v "$(pwd)/config.yaml:/workspace/.jvs/config.yaml:ro" \
           -v "$(pwd)/workspace:/workspace" \
           ghcr.io/jvs-project/jvs snapshot "custom config"
```

## Examples

### Basic Usage

```bash
# Create a workspace directory
mkdir -p ./workspace
cd ./workspace

# Initialize a JVS repository
docker run --rm -v "$(pwd):/workspace" ghcr.io/jvs-project/jvs init myproject

# Create a snapshot
docker run --rm -v "$(pwd):/workspace" ghcr.io/jvs-project/jvs snapshot "initial commit"

# List history
docker run --rm -v "$(pwd):/workspace" ghcr.io/jvs-project/jvs history
```

### With Compression

```bash
docker run --rm -v "$(pwd):/workspace" \
    ghcr.io/jvs-project/jvs snapshot "compressed" --compress fast
```

### Partial Snapshots

```bash
docker run --rm -v "$(pwd):/workspace" \
    ghcr.io/jvs-project/jvs snapshot "models only" --paths models/
```

### JSON Output

```bash
docker run --rm -v "$(pwd):/workspace" \
    ghcr.io/jvs-project/jvs --json history
```

## JuiceFS Integration

For JuiceFS integration, you need to:

1. Mount JuiceFS in the container
2. Set the appropriate engine

```bash
docker run --rm \
    -v "$(pwd):/workspace" \
    -v /path/to/juicefs/mount:/mnt/juicefs:ro \
    -e JVS_ENGINE=juicefs-clone \
    ghcr.io/jvs-project/jvs init myproject
```

## Building Locally

```bash
# Build the image
docker build -t jvs:local .

# Run the local image
docker run --rm -v "$(pwd)/workspace:/workspace" jvs:local version
```

## Multi-Platform Builds

To build for multiple platforms locally:

```bash
# Install buildx
docker buildx create --use

# Build and load for current platform
docker buildx build --platform linux/amd64 --load -t jvs:local .

# Build for multiple platforms (requires push)
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    -t ghcr.io/jvs-project/jvs:local \
    --push .
```

## Security

### Non-Root User

The JVS Docker image runs as a non-root user (uid 1000) for improved security.

### Image Scanning

Each published image is scanned for vulnerabilities using Trivy. Results are uploaded to GitHub Security.

### SBOM

Software Bill of Materials (SBOM) is generated for each image build.

## Production Usage

### Docker Compose Example

```yaml
version: '3.8'

services:
  jvs:
    image: ghcr.io/jvs-project/jvs:v1
    volumes:
      - ./workspace:/workspace
      - jvs-data:/workspace/.jvs
    environment:
      - JVS_ENGINE=copy
    # Use entrypoint wrapper for custom scripts
    # entrypoint: /usr/local/bin/jvs-wrapper.sh

volumes:
  jvs-data:
```

### Kubernetes Example

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: jvs-workspace
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi

---
apiVersion: batch/v1
kind: Job
metadata:
  name: jvs-snapshot
spec:
  template:
    spec:
      containers:
      - name: jvs
        image: ghcr.io/jvs-project/jvs:v1
        command: ["jvs", "snapshot", "scheduled"]
        volumeMounts:
        - name: workspace
          mountPath: /workspace
      volumes:
      - name: workspace
        persistentVolumeClaim:
          claimName: jvs-workspace
      restartPolicy: OnFailure
```

## Troubleshooting

### Permission Issues

If you encounter permission issues, ensure your workspace directory is writable by uid 1000:

```bash
mkdir -p ./workspace
chmod 777 ./workspace
```

### Volume Mount Issues

Always use absolute paths when mounting volumes:

```bash
# Correct
docker run -v "/absolute/path/to/workspace:/workspace" ...

# Incorrect (may fail)
docker run -v "./relative/path:/workspace" ...
```

### Debug Mode

Enable debug logging:

```bash
docker run --rm -v "$(pwd):/workspace" \
    ghcr.io/jvs-project/jvs --debug history
```

## See Also

- [Main README](../README.md)
- [CLI Reference](CLI_SPEC.md)
- [Quick Start Guide](QUICKSTART.md)

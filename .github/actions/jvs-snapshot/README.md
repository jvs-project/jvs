# JVS Snapshot Action

A GitHub Action that takes snapshots of your workspace using [JVS (Juicy Versioned Workspaces)](https://github.com/jvs-project/jvs).

## Features

- 📸 **Automatic Snapshots**: Capture workspace state before/after critical operations
- 🏷️ **Tag Support**: Organize snapshots with tags
- 🔍 **Snapshot Metadata**: Returns snapshot IDs for restoration
- 📦 **Auto-installation**: Automatically installs JVS if not present
- 🚀 **Fast**: Uses JuiceFS clone for O(1) snapshots when available

## Usage

### Basic Usage

```yaml
name: CI with Snapshot

on: [push]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # Run your tests
      - name: Run tests
        run: make test

      # Take snapshot after tests
      - name: Snapshot workspace
        uses: ./.github/actions/jvs-snapshot@v1
        with:
          note: 'After tests - ${{ github.sha }}'
          tags: 'ci,after-tests'
```

### Advanced Usage

#### Custom Path

Snapshot a specific subdirectory:

```yaml
- name: Snapshot data directory
  uses: ./.github/actions/jvs-snapshot@v1
  with:
    path: 'data/'
    note: 'Data backup'
```

#### Multiple Tags

```yaml
- name: Tagged snapshot
  uses: ./.github/actions/jvs@v1
  with:
    note: 'Release build'
    tags: 'release,v1.0,production'
```

#### Install Only (for custom workflows)

```yaml
- name: Install JVS
  uses: ./.github/actions/jvs-snapshot@v1
  with:
    install-only: 'true'

- name: Custom snapshot logic
  shell: bash
  run: |
    jvs snapshot "Custom snapshot with ${{ github.ref_name }}"
```

#### Specific JVS Version

```yaml
- name: Snapshot with specific version
  uses: ./.github/actions/jvs@v1
  with:
    jvs-version: 'v7.0'
    note: 'Pinned version snapshot'
```

### Outputs

The action provides the following outputs that can be used in subsequent steps:

```yaml
- name: Snapshot workspace
  id: snap
  uses: ./.github/actions/jvs-snapshot@v1
  with:
    note: 'Before deployment'

- name: Use snapshot ID
  run: |
    echo "Snapshot ID: ${{ steps.snap.outputs.snapshot-id }}"
    echo "Short ID: ${{ steps.snap.outputs.snapshot-short-id }}"

- name: Deploy
  run: ./deploy.sh
  env:
    SNAPSHOT_ID: ${{ steps.snap.outputs.snapshot-id }}
```

## Use Cases

### 1. Pre/Post Deployment Snapshots

```yaml
- name: Snapshot before deploy
  uses: ./.github/actions/jvs-snapshot@v1
  with:
    note: 'Before deployment - ${{ github.sha }}'

- name: Deploy
  run: ./deploy.sh

- name: Snapshot after deploy
  uses: ./.github/actions/jvs-snapshot@v1
  with:
    note: 'After deployment - ${{ github.sha }}'
```

### 2. Tagged Releases

```yaml
on:
  push:
    tags:
      - 'v*'

jobs:
  snapshot-release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Snapshot release
        uses: ./.github/actions/jvs-snapshot@v1
        with:
          note: "Release ${{ github.ref_name }}"
          tags: "release,${{ github.ref_name }}"
```

### 3. Multi-Stage CI

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Snapshot before build
        uses: ./.github/actions/jvs-snapshot@v1
        with:
          note: 'Before build'

      - name: Build
        run: make build

      - name: Snapshot after build
        uses: ./.github/actions/jvs-snapshot@v1
        with:
          note: 'After build'
          tags: 'ci,build'
```

### 4. Integration Testing

```yaml
jobs:
  integration-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup environment
        run: ./setup-env.sh

      - name: Run integration tests
        run: ./integration-tests.sh

      - name: Snapshot test results
        uses: ./.github/actions/jvs-snapshot@v1
        with:
          path: 'test-results/'
          note: 'Integration test results - ${{ github.sha }}'
```

## Requirements

- The action automatically installs JVS if not present
- Works on Ubuntu, macOS, and Windows runners
- For optimal performance, use JuiceFS-mounted workspaces

## Troubleshooting

### "Not in a JVS worktree" Error

The action will attempt to initialize a JVS repository automatically. If you encounter issues:

1. Ensure you have write permissions to the workspace
2. Check that `.jvs/` directory exists at the repository root
3. For custom paths, ensure the path is within the repository

### Permission Issues

The action uses `sudo` to install JVS to `/usr/local/bin/`. If your runner has restrictions:

```yaml
- name: Install JVS manually
  run: |
    curl -LO https://github.com/jvs-project/jvs/releases/latest/download/jvs-linux-amd64
    chmod +x jvs-linux-amd64
    mkdir -p ~/bin
    mv jvs-linux-amd64 ~/bin/jvs
    echo "$HOME/bin" >> $GITHUB_PATH

- name: Snapshot with custom install
  uses: ./.github/actions/jvs-snapshot@v1
  with:
    install-only: 'true'

- name: Take snapshot manually
  run: ~/bin/jvs snapshot "Manual snapshot"
```

## License

This action is distributed under the same license as JVS (MIT License).

## Contributing

Contributions are welcome! Please open an issue or pull request in the main JVS repository.

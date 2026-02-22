# JVS Homebrew Tap

This is the official Homebrew tap for [JVS (Juicy Versioned Workspaces)](https://github.com/jvs-project/jvs).

## Installation

### Install from Tap

```bash
brew tap jvs-project/jvs
brew install jvs
```

### Upgrade

```bash
brew upgrade jvs
```

### Uninstall

```bash
brew uninstall jvs
brew untap jvs-project/jvs
```

## Usage

After installation, the `jvs` command will be available:

```bash
# Check version
jvs version

# Initialize a new repository
jvs init my-workspace
cd my-workspace/main

# Create a snapshot
jvs snapshot "Initial state"

# View history
jvs history

# Restore a snapshot
jvs restore <snapshot-id>
```

## Documentation

For full documentation, see the [JVS GitHub repository](https://github.com/jvs-project/jvs).

## Development

### Building the Formula

To test changes to the formula locally:

```bash
# Clone this tap
git clone https://github.com/jvs-project/homebrew-jvs.git
cd homebrew-jvs

# Install from local file
brew install --build-from-source Formula/jvs.rb
```

### Updating for a New Release

When releasing a new version of JVS:

1. Build release binaries for all platforms
2. Calculate SHA256 checksums: `shasum -a 256 *.tar.gz`
3. Update the formula:
   - Update the `url` to point to the new release tag
   - Update the source `sha256`
   - Update bottle `sha256` values for each platform
4. Submit a pull request or commit the changes

## License

MIT License - See [LICENSE](https://github.com/jvs-project/jvs/blob/main/LICENSE) for details.

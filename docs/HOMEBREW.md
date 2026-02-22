# Installing JVS with Homebrew

JVS can be easily installed on macOS and Linux using Homebrew.

## Quick Install

```bash
brew tap jvs-project/jvs
brew install jvs
```

## Requirements

- macOS (Intel or Apple Silicon) or Linux
- [Homebrew](https://brew.sh/) installed
- Go 1.21+ (for building from source)

## Installation Methods

### Method 1: Install from Tap (Recommended)

This is the easiest way to install JVS. The tap provides pre-built binaries (bottles) for your platform.

```bash
brew tap jvs-project/jvs
brew install jvs
```

### Method 2: Build from Source

If you prefer to build from source or bottles are not available for your platform:

```bash
brew tap jvs-project/jvs
brew install --build-from-source jvs
```

## Verifying Installation

After installation, verify that JVS is working:

```bash
jvs version
jvs --help
```

## Upgrading

To upgrade to the latest version:

```bash
brew upgrade jvs
```

## Uninstalling

To remove JVS from your system:

```bash
brew uninstall jvs
brew untap jvs-project/jvs
```

## Shell Completions

The Homebrew formula automatically installs shell completions for bash, zsh, and fish.

### Bash

Completions are automatically installed. If they don't work, add to your `~/.bashrc`:

```bash
[[ -r "/usr/local/etc/profile.d/bash_completion.sh" ]] && . "/usr/local/etc/profile.d/bash_completion.sh"
```

### Zsh

Completions are automatically installed. If they don't work, add to your `~/.zshrc`:

```bash
# If you're using oh-my-zsh, completions should work automatically
# Otherwise, add:
fpath=(/usr/local/share/zsh/site-functions $fpath)
```

### Fish

Completions are automatically installed to `~/.config/fish/completions/`.

## Troubleshooting

### Command Not Found

If you get `command not found: jvs`, ensure Homebrew's bin directory is in your PATH:

```bash
# For macOS Intel
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.zshrc

# For macOS Apple Silicon
echo 'export PATH="/opt/homebrew/bin:$PATH"' >> ~/.zshrc

# For Linux
echo 'export PATH="/home/linuxbrew/.linuxbrew/bin:$PATH"' >> ~/.bashrc
```

### Permission Issues

If you encounter permission errors, try:

```bash
brew doctor
brew fix-perms
```

### Outdated Version

If you're not getting the latest version:

```bash
brew update
brew upgrade jvs
```

## Development

### Testing Formula Changes

To test local changes to the formula:

```bash
# Clone the tap repository
git clone https://github.com/jvs-project/homebrew-jvs.git
cd homebrew-jvs

# Install from local formula file
brew install --build-from-source Formula/jvs.rb
```

### Auditing the Formula

To ensure the formula follows Homebrew best practices:

```bash
brew audit Formula/jvs.rb --strict --online
```

## Next Steps

After installing JVS:

1. Read the [Quick Start Guide](QUICKSTART.md)
2. Explore [Configuration options](CONFIGURATION.md)
3. Check out [Examples](EXAMPLES.md)

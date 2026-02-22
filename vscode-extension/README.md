# JVS VS Code Extension

Visual Studio Code extension for [JVS (Juicy Versioned Workspaces)](https://github.com/jvs-project/jvs) - workspace versioning and snapshot management for JuiceFS.

## Features

- **Snapshot Management**: Create, restore, and delete snapshots directly from VS Code
- **Visual History**: Browse snapshot history with notes, tags, and timestamps
- **Worktree Support**: Fork and manage worktrees for parallel development
- **Status Bar Integration**: Quick access to repository info and snapshot count
- **Automatic Detection**: Extension activates when opening a JVS workspace
- **File System Watching**: Auto-refreshes when snapshots or worktrees change

## Requirements

- [JVS CLI](https://github.com/jvs-project/jvs) installed and available in your PATH
- A JVS repository (initialized with `jvs init <name>`)
- VS Code 1.80.0 or higher

## Installation

### From Marketplace (Coming Soon)

Search for "JVS" in the VS Code Extensions marketplace.

### From Source

1. Clone this repository
2. Install dependencies: `npm install`
3. Compile: `npm run compile`
4. Package: `vsce package`
5. Install the `.vsix` file in VS Code

## Usage

### Getting Started

1. Open a workspace that contains a `.jvs` directory
2. The JVS sidebar will appear automatically

### Commands

| Command | Description |
|---------|-------------|
| `JVS: Create Snapshot` | Create a new snapshot with optional note and tags |
| `JVS: Restore Snapshot...` | Choose a snapshot to restore from a quick pick |
| `JVS: Return to HEAD` | Restore to the latest state |
| `JVS: Show History` | Focus the snapshot history view |
| `JVS: Show Repository Info` | Display repository information |
| `JVS: Refresh` | Refresh all data |
| `JVS: Fork Worktree...` | Create a new worktree from current state |
| `JVS: Verify All Snapshots` | Run verification on all snapshots |

### Sidebar Views

#### Snapshots View
- Lists all snapshots in chronological order
- Shows snapshot ID (short), note, tags, and creation time
- Right-click to restore or delete

#### Worktrees View
- Lists all worktrees in the repository
- Shows current worktree with a special icon
- Click to open a worktree in a new VS Code window

### Configuration

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `jvs.autoSnapshot` | boolean | `false` | Automatically create snapshot before changes |
| `jvs.defaultEngine` | enum | `"copy"` | Default snapshot engine (`copy`, `reflink-copy`, `juicefs-clone`) |
| `jvs.historyLimit` | number | `50` | Number of snapshots to show in history (1-500) |
| `jvs.executablePath` | string | `"jvs"` | Path to JVS executable |

## Screenshots

### Sidebar
![JVS Sidebar showing snapshot history](images/sidebar.png)

### Create Snapshot Dialog
![Create snapshot dialog with note and tags input](images/create-snapshot.png)

### Status Bar
![Status bar showing JVS icon and snapshot count](images/status-bar.png)

## Keyboard Shortcuts

You can customize keyboard shortcuts in your VS Code `keybindings.json`:

```json
[
  {
    "key": "ctrl+shift+s",
    "command": "jvs.snapshot",
    "when": "jvs:hasRepo"
  },
  {
    "key": "ctrl+shift+r",
    "command": "jvs.restore",
    "when": "jvs:hasRepo"
  }
]
```

## Troubleshooting

### Extension doesn't activate

Make sure:
1. JVS CLI is installed and in your PATH
2. Your workspace contains a `.jvs` directory
3. You're inside a JVS worktree (e.g., `repo/main/`)

### Commands fail

Check the **JVS** output channel for error logs:
1. Open the Output panel (`Ctrl+Shift+U` / `Cmd+Shift+U`)
2. Select "JVS" from the dropdown

### JVS not found

If JVS is installed but not found:
1. Set the `jvs.executablePath` setting to the full path
2. Or add JVS to your system PATH

## Development

```bash
# Install dependencies
npm install

# Compile TypeScript
npm run compile

# Watch for changes
npm run watch

# Run linting
npm run lint

# Run tests
npm run test

# Package extension
vsce package
```

## Contributing

Contributions are welcome! Please open an issue or pull request in the [main JVS repository](https://github.com/jvs-project/jvs).

## License

MIT License - See [LICENSE](LICENSE) for details.

## Related

- [JVS CLI](https://github.com/jvs-project/jvs) - Core command-line tool
- [JuiceFS](https://github.com/juicedata/juicefs) - Distributed file system
- [JVS GitHub Action](https://github.com/jvs-project/jvs/tree/main/.github/actions/jvs-snapshot) - CI/CD integration

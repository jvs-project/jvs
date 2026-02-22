import * as vscode from 'vscode';
import * as path from 'path';
import { exec } from 'child_process';

interface JVSRepoInfo {
  root: string;
  engine: string;
  totalSnapshots: number;
  totalWorktrees: number;
}

interface JVSSnapshot {
  snapshot_id: string;
  note: string;
  tags: string[];
  created_at: string;
}

interface JVSWorktree {
  name: string;
  path: string;
  head_id: string | null;
  latest_id: string | null;
  is_current: boolean;
  is_detached: boolean;
}

class JVSExtension {
  private context: vscode.ExtensionContext;
  private repoInfo: JVSRepoInfo | null = null;
  private statusBarItem: vscode.StatusBarItem;
  private historyProvider: HistoryProvider;
  private worktreeProvider: WorktreeProvider;
  private outputChannel: vscode.OutputChannel;

  constructor(context: vscode.ExtensionContext) {
    this.context = context;
    this.outputChannel = vscode.window.createOutputChannel('JVS');
    this.statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    this.statusBarItem.command = 'jvs.info';

    this.historyProvider = new HistoryProvider(context, this.outputChannel);
    this.worktreeProvider = new WorktreeProvider(context, this.outputChannel);

    this.initialize();
  }

  private async initialize() {
    // Check if we're in a JVS repository
    await this.detectRepo();

    // Register tree data providers
    vscode.window.registerTreeDataProvider('jvs.historyTree', this.historyProvider);
    vscode.window.registerTreeDataProvider('jvs.worktreeTree', this.worktreeProvider);

    // Register commands
    this.registerCommands();

    // Update status bar
    this.updateStatusBar();

    // Watch for changes
    this.watchForChanges();
  }

  private async detectRepo() {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
      return;
    }

    try {
      const info = await this.execJVS(['info', '--json'], workspaceFolder.uri.fsPath);
      this.repoInfo = JSON.parse(info);

      vscode.commands.executeCommand('setContext', 'jvs:hasRepo', true);
      await this.refreshData();
    } catch (error) {
      // Not in a JVS repo or JVS not installed
      this.repoInfo = null;
      vscode.commands.executeCommand('setContext', 'jvs:hasRepo', false);
    }
  }

  private registerCommands() {
    const commands = [
      { id: 'jvs.snapshot', handler: this.createSnapshot.bind(this) },
      { id: 'jvs.restore', handler: this.restoreSnapshot.bind(this) },
      { id: 'jvs.restoreHead', handler: this.restoreHead.bind(this) },
      { id: 'jvs.history', handler: this.showHistory.bind(this) },
      { id: 'jvs.info', handler: this.showInfo.bind(this) },
      { id: 'jvs.refresh', handler: this.refresh.bind(this) },
      { id: 'jvs.worktree.fork', handler: this.forkWorktree.bind(this) },
      { id: 'jvs.verify', handler: this.verify.bind(this) },
      { id: 'jvs.snapshot.delete', handler: this.deleteSnapshot.bind(this) },
      { id: 'jvs.snapshot.restore', handler: this.restoreSnapshotFromTree.bind(this) }
    ];

    commands.forEach(({ id, handler }) => {
      const disposable = vscode.commands.registerCommand(id, handler);
      this.context.subscriptions.push(disposable);
    });
  }

  private async createSnapshot() {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
      vscode.window.showErrorMessage('No workspace folder found');
      return;
    }

    const note = await vscode.window.showInputBox({
      prompt: 'Enter snapshot note (optional):',
      placeHolder: 'e.g., "Before implementing feature X"'
    });

    if (note === undefined) {
      // User cancelled
      return;
    }

    const tagsInput = await vscode.window.showInputBox({
      prompt: 'Enter tags (comma-separated, optional):',
      placeHolder: 'e.g., "feature-x,experimental"'
    });

    if (tagsInput === undefined) {
      // User cancelled
      return;
    }

    try {
      this.outputChannel.appendLine(`Creating snapshot: ${note}`);
      const tags = tagsInput ? tagsInput.split(',').map(t => t.trim()).filter(t => t) : [];
      const tagArgs = tags.flatMap(t => ['--tag', t]);

      const output = await this.execJVS(
        ['snapshot', ...(note ? [note] : []), ...tagArgs],
        workspaceFolder.uri.fsPath
      );
      this.outputChannel.appendLine(output);

      const selection = await vscode.window.showInformationMessage('Snapshot created!', 'View History', 'OK');
      if (selection === 'View History') {
        await this.showHistory();
      }

      await this.refreshData();
    } catch (error: any) {
      vscode.window.showErrorMessage(`Failed to create snapshot: ${error.message}`);
    }
  }

  private async restoreSnapshot() {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
      vscode.window.showErrorMessage('No workspace folder found');
      return;
    }

    // Show quick pick with recent snapshots
    const snapshots = await this.historyProvider.getSnapshots();
    if (snapshots.length === 0) {
      vscode.window.showInformationMessage('No snapshots found. Create one first!');
      return;
    }

    const items = snapshots.slice(0, 10).map((s) => ({
      label: `${s.shortId} - ${s.note || '(no note)'}`,
      description: `${s.tags.join(', ') || 'no tags'} - ${formatDate(s.createdAt)}`,
      snapshotId: s.id
    }));

    const selected = await vscode.window.showQuickPick(items, {
      placeHolder: 'Select snapshot to restore'
    });

    if (selected) {
      await this.doRestore(selected.snapshotId, workspaceFolder.uri.fsPath);
    }
  }

  private async restoreSnapshotFromTree(snapshotId: string) {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
      vscode.window.showErrorMessage('No workspace folder found');
      return;
    }

    await this.doRestore(snapshotId, workspaceFolder.uri.fsPath);
  }

  private async doRestore(snapshotId: string, workspacePath: string) {
    const confirm = await vscode.window.showWarningMessage(
      `Restore to ${snapshotId.substring(0, 8)}? This will modify your workspace.`,
      { modal: true },
      'Restore',
      'Cancel'
    );

    if (confirm === 'Restore') {
      try {
        this.outputChannel.show();
        this.outputChannel.appendLine(`Restoring to ${snapshotId}...`);
        await this.execJVS(['restore', snapshotId], workspacePath);
        this.outputChannel.appendLine('Restore complete!');
        vscode.window.showInformationMessage(`Restored to ${snapshotId.substring(0, 8)}`);
        await this.refreshData();
      } catch (error: any) {
        vscode.window.showErrorMessage(`Failed to restore: ${error.message}`);
      }
    }
  }

  private async deleteSnapshot(snapshotId: string) {
    const confirm = await vscode.window.showWarningMessage(
      `Delete snapshot ${snapshotId.substring(0, 8)}? This cannot be undone.`,
      { modal: true },
      'Delete',
      'Cancel'
    );

    if (confirm === 'Delete') {
      const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
      if (!workspaceFolder) {
        return;
      }

      try {
        await this.execJVS(['snapshot', 'delete', snapshotId], workspaceFolder.uri.fsPath);
        vscode.window.showInformationMessage(`Snapshot ${snapshotId.substring(0, 8)} deleted`);
        await this.refreshData();
      } catch (error: any) {
        vscode.window.showErrorMessage(`Failed to delete snapshot: ${error.message}`);
      }
    }
  }

  private async restoreHead() {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
      vscode.window.showErrorMessage('No workspace folder found');
      return;
    }

    try {
      await this.execJVS(['restore', 'HEAD'], workspaceFolder.uri.fsPath);
      vscode.window.showInformationMessage('Returned to HEAD state');
      await this.refreshData();
    } catch (error: any) {
      vscode.window.showErrorMessage(`Failed to return to HEAD: ${error.message}`);
    }
  }

  private async showInfo() {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
      vscode.window.showErrorMessage('No workspace folder found');
      return;
    }

    try {
      const info = await this.execJVS(['info', '--json'], workspaceFolder.uri.fsPath);
      const repoInfo: JVSRepoInfo = JSON.parse(info);

      const message = [
        `**Repository:** ${repoInfo.root}`,
        `**Engine:** ${repoInfo.engine}`,
        `**Snapshots:** ${repoInfo.totalSnapshots}`,
        `**Worktrees:** ${repoInfo.totalWorktrees}`
      ].join('\n\n');

      vscode.window.showInformationMessage(message, 'OK');
    } catch (error: any) {
      vscode.window.showErrorMessage(`Failed to get info: ${error.message}`);
    }
  }

  private async showHistory() {
    await vscode.commands.executeCommand('jvs.historyView.focus');
  }

  private async refresh() {
    await this.detectRepo();
    await this.refreshData();
    vscode.window.showInformationMessage('JVS data refreshed');
  }

  private async forkWorktree() {
    const name = await vscode.window.showInputBox({
      prompt: 'Enter worktree name:',
      placeHolder: 'feature-branch',
      validateInput: (value) => {
        if (!value || !value.match(/^[a-zA-Z0-9-_]+$/)) {
          return 'Name must contain only alphanumeric characters, dashes, and underscores';
        }
        return null;
      }
    });

    if (!name) {
      return;
    }

    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
      vscode.window.showErrorMessage('No workspace folder found');
      return;
    }

    try {
      this.outputChannel.appendLine(`Forking worktree: ${name}`);
      const output = await this.execJVS(['worktree', 'fork', name], workspaceFolder.uri.fsPath);
      this.outputChannel.appendLine(output);

      const selection = await vscode.window.showInformationMessage(
        `Worktree "${name}" created`,
        'Open in New Window',
        'OK'
      );

      if (selection === 'Open in New Window') {
        // Parse the worktree path from output or construct it
        const worktreePath = path.join(workspaceFolder.uri.fsPath, '..', 'worktrees', name);
        const uri = vscode.Uri.file(worktreePath);
        await vscode.commands.executeCommand('vscode.openFolder', uri, { forceNewWindow: true });
      }

      await this.refreshData();
    } catch (error: any) {
      vscode.window.showErrorMessage(`Failed to fork worktree: ${error.message}`);
    }
  }

  private async verify() {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
      vscode.window.showErrorMessage('No workspace folder found');
      return;
    }

    try {
      this.outputChannel.show();
      this.outputChannel.appendLine('Verifying all snapshots...');
      const output = await this.execJVS(['verify', '--all'], workspaceFolder.uri.fsPath);
      this.outputChannel.appendLine(output);
      this.outputChannel.appendLine('Verification complete!');
      vscode.window.showInformationMessage('All snapshots verified successfully');
    } catch (error: any) {
      this.outputChannel.appendLine(`Verification failed: ${error.message}`);
      vscode.window.showErrorMessage(`Verification failed: ${error.message}`);
    }
  }

  private async refreshData() {
    await Promise.all([
      this.historyProvider.refresh(),
      this.worktreeProvider.refresh()
    ]);
    this.updateStatusBar();
  }

  private updateStatusBar() {
    if (this.repoInfo) {
      this.statusBarItem.text = `$(database) JVS $(chevron-down)`;
      this.statusBarItem.tooltip = new vscode.MarkdownString(
        `**JVS Repository**\n\n` +
        `- Root: \`${this.repoInfo.root}\`\n` +
        `- Engine: ${this.repoInfo.engine}\n` +
        `- Snapshots: ${this.repoInfo.totalSnapshots}\n` +
        `- Worktrees: ${this.repoInfo.totalWorktrees}`
      );
      this.statusBarItem.show();
    } else {
      this.statusBarItem.hide();
    }
  }

  private watchForChanges() {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
      return;
    }

    // Watch for .jvs directory changes
    const jvsWatcher = vscode.workspace.createFileSystemWatcher(
      path.join(workspaceFolder.uri.fsPath, '.jvs', '**'),
      false,
      true,
      false
    );

    const onJvsChange = () => {
      this.refreshData();
    };

    this.context.subscriptions.push(jvsWatcher.onDidChange(onJvsChange));
    this.context.subscriptions.push(jvsWatcher.onDidCreate(onJvsChange));
    this.context.subscriptions.push(jvsWatcher.onDidDelete(onJvsChange));
  }

  private async execJVS(args: string[], cwd: string): Promise<string> {
    return new Promise((resolve, reject) => {
      const cmd = `jvs ${args.map(a => a.includes(' ') ? `"${a}"` : a).join(' ')}`;
      exec(cmd, { cwd, env: process.env },
        (error, stdout, stderr) => {
          if (error) {
            reject(new Error(stderr || error.message));
          } else {
            resolve(stdout || stderr);
          }
        }
      );
    });
  }
}

interface SnapshotInfo {
  id: string;
  shortId: string;
  note: string;
  tags: string[];
  createdAt: Date;
}

class HistoryProvider implements vscode.TreeDataProvider<SnapshotTreeItem> {
  private _onDidChangeTreeData = new vscode.EventEmitter<SnapshotTreeItem | undefined | void>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  constructor(
    private context: vscode.ExtensionContext,
    private outputChannel: vscode.OutputChannel
  ) {}

  refresh(): Promise<void> {
    this._onDidChangeTreeData.fire();
    return Promise.resolve();
  }

  getTreeItem(element: SnapshotTreeItem): vscode.TreeItem {
    return element;
  }

  async getChildren(element?: SnapshotTreeItem): Promise<SnapshotTreeItem[]> {
    if (!element) {
      return await this.getSnapshots();
    }
    return [];
  }

  async getSnapshots(): Promise<SnapshotTreeItem[]> {
    try {
      const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
      if (!workspaceFolder) {
        return [];
      }

      const config = vscode.workspace.getConfiguration('jvs');
      const limit = config.get<number>('historyLimit', 50);

      const output = await this.execJVS(['history', '--json', '--limit', String(limit)], workspaceFolder.uri.fsPath);
      const data: JVSSnapshot[] = JSON.parse(output);

      return data.map((s) => {
        const item = new SnapshotTreeItem(
          s.snapshot_id,
          s.snapshot_id.substring(0, 8),
          s.note || '(no note)',
          s.tags,
          new Date(s.created_at)
        );
        return item;
      });
    } catch (error) {
      this.outputChannel.appendLine(`Failed to fetch snapshots: ${error}`);
      return [];
    }
  }

  async getSnapshotInfos(): Promise<SnapshotInfo[]> {
    const items = await this.getSnapshots();
    return items.map(item => ({
      id: item.snapshotId,
      shortId: item.shortId,
      note: item.note,
      tags: item.tags,
      createdAt: item.createdAt
    }));
  }

  private async execJVS(args: string[], cwd: string): Promise<string> {
    return new Promise((resolve, reject) => {
      const cmd = `jvs ${args.join(' ')}`;
      exec(cmd, { cwd, env: process.env },
        (error, stdout, stderr) => {
          if (error) {
            reject(new Error(stderr || error.message));
          } else {
            resolve(stdout || stderr);
          }
        }
      );
    });
  }
}

class SnapshotTreeItem extends vscode.TreeItem {
  constructor(
    public readonly snapshotId: string,
    public readonly shortId: string,
    public readonly note: string,
    public readonly tags: string[],
    public readonly createdAt: Date
  ) {
    super(`${shortId} - ${note}`, vscode.TreeItemCollapsibleState.None);

    this.description = tags.length > 0 ? tags.join(', ') : formatDate(createdAt);
    this.tooltip = new vscode.MarkdownString(
      `**Snapshot:** \`${shortId}\`\n\n` +
      `**Note:** ${note || '(no note)'}\n` +
      `**Tags:** ${tags.join(', ') || '(none)'}\n` +
      `**Created:** ${createdAt.toLocaleString()}\n\n` +
      `---\n\n` +
      `Click to restore`
    );
    this.contextValue = 'jvs.snapshot';
    this.iconPath = new vscode.ThemeIcon('camera');
  }
}

class WorktreeProvider implements vscode.TreeDataProvider<WorktreeTreeItem> {
  private _onDidChangeTreeData = new vscode.EventEmitter<WorktreeTreeItem | undefined | void>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  constructor(
    private context: vscode.ExtensionContext,
    private outputChannel: vscode.OutputChannel
  ) {}

  refresh(): Promise<void> {
    this._onDidChangeTreeData.fire();
    return Promise.resolve();
  }

  getTreeItem(element: WorktreeTreeItem): vscode.TreeItem {
    return element;
  }

  async getChildren(element?: WorktreeTreeItem): Promise<WorktreeTreeItem[]> {
    if (!element) {
      return await this.getWorktrees();
    }
    return [];
  }

  private async getWorktrees(): Promise<WorktreeTreeItem[]> {
    try {
      const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
      if (!workspaceFolder) {
        return [];
      }

      const output = await this.execJVS(['worktree', 'list', '--json'], workspaceFolder.uri.fsPath);
      const data: JVSWorktree[] = JSON.parse(output);

      return data.map(w => new WorktreeTreeItem(w));
    } catch (error) {
      this.outputChannel.appendLine(`Failed to fetch worktrees: ${error}`);
      return [];
    }
  }

  private async execJVS(args: string[], cwd: string): Promise<string> {
    return new Promise((resolve, reject) => {
      const cmd = `jvs ${args.join(' ')}`;
      exec(cmd, { cwd, env: process.env },
        (error, stdout, stderr) => {
          if (error) {
            reject(new Error(stderr || error.message));
          } else {
            resolve(stdout || stderr);
          }
        }
      );
    });
  }
}

class WorktreeTreeItem extends vscode.TreeItem {
  constructor(public readonly worktree: JVSWorktree) {
    super(worktree.name, vscode.TreeItemCollapsibleState.None);

    const status = worktree.is_detached ? '(detached)' : '';
    this.description = status || worktree.path;
    this.tooltip = new vscode.MarkdownString(
      `**Worktree:** ${worktree.name}\n\n` +
      `**Path:** ${worktree.path}\n` +
      `**Head:** ${worktree.head_id || '(none)'}\n` +
      `**Latest:** ${worktree.latest_id || '(none)'}\n` +
      `**Status:** ${worktree.is_detached ? 'detached' : 'normal'}`
    );

    if (worktree.is_current) {
      this.iconPath = new vscode.ThemeIcon('folder-active');
    } else if (worktree.is_detached) {
      this.iconPath = new vscode.ThemeIcon('folder', new vscode.ThemeColor('warningForeground'));
    } else {
      this.iconPath = new vscode.ThemeIcon('folder');
    }

    this.contextValue = 'jvs.worktree';
    this.command = {
      command: 'vscode.openFolder',
      title: 'Open Worktree',
      arguments: [vscode.Uri.file(worktree.path)]
    };
  }
}

function formatDate(date: Date): string {
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (diffDays === 0) {
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    if (diffHours === 0) {
      const diffMins = Math.floor(diffMs / (1000 * 60));
      return diffMins === 0 ? 'just now' : `${diffMins}m ago`;
    }
    return `${diffHours}h ago`;
  } else if (diffDays === 1) {
    return 'yesterday';
  } else if (diffDays < 7) {
    return `${diffDays}d ago`;
  }

  return date.toLocaleDateString();
}

export function activate(context: vscode.ExtensionContext) {
  const jvsExtension = new JVSExtension(context);
}

export function deactivate() {
  // Cleanup
}

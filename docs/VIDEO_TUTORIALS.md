# JVS Video Tutorial Outlines

**Version:** 1.0
**Last Updated:** 2026-02-23
**Target Audience:** Developers, Data Scientists, DevOps Engineers

---

## Video 1: Getting Started with JVS (5 minutes)

**Target:** New users who want to try JVS in 5 minutes

### Outline

| Time | Content | Visual |
|------|---------|--------|
| 0:00-0:30 | **Hook**: Problem with Git for large workspaces | Split screen: Git struggling with large files vs JVS instant snapshot |
| 0:30-1:00 | **What is JVS?**: 3 key points (snapshot-first, O(1), filesystem-native) | Diagram: .jvs/ + main/ structure |
| 1:00-1:30 | **Prerequisites**: Go, JuiceFS (optional) | Terminal: `go version`, `juicefs version` |
| 1:30-2:30 | **Installation & Init**: Clone, build, init repo | Terminal: `make build`, `jvs init myproject` |
| 2:30-3:30 | **First Snapshot**: Create files, snapshot, view history | Terminal: `jvs snapshot "initial"`, `jvs history` |
| 3:30-4:15 | **Restore Workflow**: Make changes, restore to previous | Terminal: `jvs restore abc123`, show files revert |
| 4:15-5:00 | **Next Steps**: Quick Start guide link, GitHub repo | Screen: Links to docs |

### Script Key Phrases

> "Git is great for code, but struggles with large datasets. JVS handles entire workspaces instantly."
> "Notice: the repo root is NOT your workspace. `main/` is where you work."
> "Snapshots are O(1) with JuiceFS - that means instant, regardless of workspace size."

### Production Notes

- **Terminal visibility**: Use large font, high contrast colors
- **Pacing**: Allow 1-2 seconds after each command for viewer to read output
- **Graphics**: Simple ASCII diagrams for architecture (avoids tool dependency)
- **Callouts**: On-screen text for key commands

---

## Video 2: Worktree Workflows (5 minutes)

**Target:** Users who want to understand branching and parallel work

### Outline

| Time | Content | Visual |
|------|---------|--------|
| 0:00-0:45 | **Recap**: Main worktree concept | Diagram: repo/main/ as primary workspace |
| 0:45-1:30 | **Why Fork?**: Parallel work without conflicts | Split screen: Two developers working separately |
| 1:30-2:30 | **Creating Forks**: `jvs worktree fork experiment` | Terminal: Creating and listing worktrees |
| 2:30-3:30 | **Switching Worktrees**: `cd` to different worktrees | Terminal: `cd ../worktrees/experiment`, `jvs history` |
| 3:30-4:15 | **Detached State**: What happens after restore? | Diagram: Main → restore → detached → fork to continue |
| 4:15-5:00 | **Cleanup**: Removing unused worktrees | Terminal: `jvs worktree remove experiment` |

### Script Key Phrases

> "Worktrees are real directories - no Git magic here. Just `cd` to switch."
> "After restoring to an old snapshot, you're in detached state. Fork to create a new branch."
> "Each worktree has its own config pointing to a snapshot. Completely independent."

### Production Notes

- **Visual metaphor**: Tree diagram showing main trunk + branches
- **Color coding**: Green for main, blue for feature worktrees
- **File browser**: Show actual directory structure during operations

---

## Video 3: Snapshot & Restore Basics (5 minutes)

**Target:** Users who want to understand snapshot internals and restore options

### Outline

| Time | Content | Visual |
|------|---------|--------|
| 0:00-0:45 | **Snapshot Anatomy**: What's in a snapshot? | Diagram: Descriptor + Payload + .READY marker |
| 0:45-1:30 | **Two-Layer Integrity**: Checksum + Payload Hash | Animation: Both layers must verify |
| 1:30-2:15 | **Tags and Notes**: Organizing snapshots | Terminal: `jvs snapshot "v1.0" --tag stable` |
| 2:15-3:00 | **Fuzzy Lookup**: Restore by ID, tag, or note | Terminal: Various restore commands |
| 3:00-4:00 | **Detached State Deep Dive**: When and why | Diagram: HEAD pointer, lineage chain |
| 4:00-5:00 | **Best Practices**: When to snapshot, naming conventions | Bullet points on screen |

### Script Key Phrases

> "Every snapshot has two integrity checks: descriptor checksum and payload hash. Both must pass."
> "You can restore by full ID, short prefix (8 chars), tag name, or even note prefix."
> "Detached state is safe - you can't accidentally lose work. Fork to create a new branch."

### Production Notes

- **Animation**: Show hash chain for audit trail
- **Zoom**: On descriptor JSON to show fields clearly
- **Comparison**: Side-by-side Git restore vs JVS restore

---

## Video 4: Advanced - GC and Doctor (5 minutes)

**Target:** Operations-focused users who maintain JVS repositories

### Outline

| Time | Content | Visual |
|------|---------|--------|
| 0:00-0:30 | **Problem**: Disk usage grows with snapshots | Graph: Storage over time |
| 0:30-1:30 | **GC Plan Preview**: What will be deleted? | Terminal: `jvs gc plan --keep-daily 7` |
| 1:30-2:15 | **GC Execution**: Running the plan safely | Terminal: `jvs gc run --plan-id XXX` |
| 2:15-3:30 | **Doctor Health Checks**: What gets checked? | Terminal: `jvs doctor --strict` |
| 3:30-4:15 | **Repair Actions**: Fixing common issues | Terminal: `jvs doctor --strict --repair-runtime` |
| 4:15-5:00 | **Ops Best Practices**: Daily/weekly maintenance | Checklist on screen |

### Script Key Phrases

> "GC is two-phase: plan first, review, then execute. No surprises."
> "Doctor checks layout, lineage, runtime state, and audit chain integrity."
> "Run `jvs doctor --strict` before any major operation to ensure repository health."

### Production Notes

- **Table view**: Show plan output in clean format
- **Warnings**: Use red/yellow highlights for doctor findings
- **Before/After**: Show storage metrics before/after GC

---

## Production Checklist

### Equipment
- [ ] 4K or 1080p resolution
- [ ] Clear microphone (USB condenser recommended)
- [ ] Clean terminal background (solid color)
- [ ] Large monospace font (Fira Code, JetBrains Mono, 14pt+)
- [ ] High contrast color scheme (Solarized Light, Dracula, One Light)

### Software
- [ ] Terminal with good rendering (iTerm2, WezTerm, GNOME Terminal)
- [ ] Screen recording: OBS Studio, CleanShot X, or similar
- [ ] Video editing: DaVinci Resolve (free) or Final Cut Pro
- [ ] Diagram tool: Keynote, PowerPoint, or Excalidraw (export as images)

### Assets Needed
- [ ] JVS logo (SVG/PNG)
- [ ] Architecture diagram (SVG for scaling)
- [ ] CNCF logo (if approved, for outro)
- [ ] Background music (optional, royalty-free)

### Distribution
- [ ] YouTube channel setup
- [ ] Thumbnail templates (1920x1080)
- [ ] Show notes template with commands
- [ ] Transcript for accessibility

---

## Style Guide

### Voice and Tone
- **Conversational but professional**
- **Avoid jargon where possible, explain when necessary**
- **Enthusiastic but not over-the-top**
- **Pacing**: 130-150 words per minute

### On-Screen Text
- **Commands**: Yellow/orange on dark background
- **File paths**: Cyan/blue
- **Key terms**: Bold, highlighted
- **Warnings**: Red with icon
- **Success messages**: Green with checkmark

### Transitions
- **Cut jumps** for immediate command results (no typing real-time)
- **Cross-fade** for scene changes
- **Zoom in** for important output
- **Callout circles** for cursor focus

---

## Related Resources

- [QUICKSTART.md](QUICKSTART.md) - Written quick start guide
- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture details
- [13_OPERATION_RUNBOOK.md](13_OPERATION_RUNBOOK.md) - Operations guide

---

*These outlines are living documents. Update based on user feedback and evolving features.*

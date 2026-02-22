package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/engine"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/model"
)

var (
	worktreeCreateFrom string
	worktreeForce      bool
)

var worktreeCmd = &cobra.Command{
	Use:     "worktree",
	Short:   "Manage worktrees",
	Aliases: []string{"wt"},
}

var worktreeCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new worktree",
	Long: `Create a new worktree.

If --from is specified, the worktree is created from an existing snapshot,
otherwise an empty worktree is created.

Examples:
  jvs worktree create feature-x                    # Create empty worktree
  jvs worktree create hotfix --from v1.0           # Create from tag
  jvs worktree create feature-y --from 1771589-abc # Create from snapshot`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()
		name := args[0]

		mgr := worktree.NewManager(r.Root)

		// If --from is specified, create from snapshot
		if worktreeCreateFrom != "" {
			snapshotID, err := resolveSnapshotID(r.Root, worktreeCreateFrom)
			if err != nil {
				fmtErr("resolve snapshot: %v", err)
				os.Exit(1)
			}

			// Verify snapshot exists and is valid
			if err := snapshot.VerifySnapshot(r.Root, snapshotID, false); err != nil {
				fmtErr("verify snapshot: %v", err)
				os.Exit(1)
			}

			// Create engine for cloning
			eng := engine.NewEngine(detectEngine(r.Root))

			cfg, err := mgr.CreateFromSnapshot(name, snapshotID, func(src, dst string) error {
				_, err := eng.Clone(src, dst)
				return err
			})
			if err != nil {
				fmtErr("create worktree from snapshot: %v", err)
				os.Exit(1)
			}

			if jsonOutput {
				outputJSON(cfg)
			} else {
				fmt.Printf("Created worktree '%s' from snapshot %s\n", name, snapshotID)
				fmt.Printf("Path: %s\n", mgr.Path(name))
			}
			return
		}

		// Create empty worktree
		cfg, err := mgr.Create(name, nil)
		if err != nil {
			fmtErr("create worktree: %v", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(cfg)
		} else {
			fmt.Printf("Created worktree '%s' at %s\n", name, mgr.Path(name))
		}
	},
}

// resolveSnapshotID resolves a snapshot reference to a full snapshot ID.
func resolveSnapshotID(repoRoot, ref string) (model.SnapshotID, error) {
	// Try exact match first
	testID := model.SnapshotID(ref)
	_, err := snapshot.LoadDescriptor(repoRoot, testID)
	if err == nil {
		return testID, nil
	}

	// Try fuzzy match
	desc, err := snapshot.FindOne(repoRoot, ref)
	if err != nil {
		return "", fmt.Errorf("snapshot not found: %s", ref)
	}
	return desc.SnapshotID, nil
}

var worktreeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all worktrees",
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()

		mgr := worktree.NewManager(r.Root)
		list, err := mgr.List()
		if err != nil {
			fmtErr("list worktrees: %v", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(list)
			return
		}

		for _, cfg := range list {
			head := string(cfg.HeadSnapshotID)
			if head == "" {
				head = "(none)"
			} else if len(head) > 16 {
				head = head[:16] + "..."
			}
			fmt.Printf("%-20s  %s\n", cfg.Name, head)
		}
	},
}

var worktreePathCmd = &cobra.Command{
	Use:   "path [<name>]",
	Short: "Print the path to a worktree",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()

		name := ""
		if len(args) > 0 {
			name = args[0]
		} else {
			_, name = requireWorktree()
		}

		mgr := worktree.NewManager(r.Root)
		path := mgr.Path(name)
		fmt.Println(path)
	},
}

var worktreeRenameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename a worktree",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()
		oldName := args[0]
		newName := args[1]

		mgr := worktree.NewManager(r.Root)
		if err := mgr.Rename(oldName, newName); err != nil {
			fmtErr("rename worktree: %v", err)
			os.Exit(1)
		}

		if !jsonOutput {
			fmt.Printf("Renamed worktree '%s' to '%s'\n", oldName, newName)
		}
	},
}

var worktreeRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a worktree",
	Long: `Remove a worktree.

The worktree payload and metadata are deleted, but all snapshots remain.
Use --force to remove a worktree that is in detached state.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()
		name := args[0]

		mgr := worktree.NewManager(r.Root)

		// Check for detached state unless --force
		if !worktreeForce {
			cfg, err := mgr.Get(name)
			if err == nil && cfg.IsDetached() {
				fmtErr("worktree '%s' is in detached state", name)
				fmt.Println()
				fmt.Printf("Current position: %s\n", cfg.HeadSnapshotID)
				fmt.Printf("Latest snapshot: %s\n", cfg.LatestSnapshotID)
				fmt.Println()
				fmt.Println("To remove anyway, use: jvs worktree remove --force " + name)
				os.Exit(1)
			}
		}

		if err := mgr.Remove(name); err != nil {
			fmtErr("remove worktree: %v", err)
			os.Exit(1)
		}

		if !jsonOutput {
			fmt.Printf("Removed worktree '%s'\n", name)
		}
	},
}

var worktreeForkCmd = &cobra.Command{
	Use:   "fork [snapshot-id] [name]",
	Short: "Create a new worktree from a snapshot",
	Long: `Create a new worktree from a snapshot.

The new worktree will be at HEAD state - you can create snapshots immediately.

If snapshot-id is omitted, uses the current worktree's position.
If name is omitted, auto-generates a name.

The snapshot-id can be:
  - A full snapshot ID
  - A short ID prefix
  - A tag name
  - A note prefix (fuzzy match)

Examples:
  jvs worktree fork                           # Fork from current position, auto-name
  jvs worktree fork feature-x                 # Fork from current position with name
  jvs worktree fork v1.0 hotfix               # Fork from tag v1.0, name hotfix
  jvs worktree fork 1771589-abc feature-y     # Fork from specific snapshot`,
	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		r, wtName := requireWorktree()

		var snapshotID model.SnapshotID
		var name string

		// Parse arguments
		switch len(args) {
		case 0:
			// No args: use current position, auto-generate name
			mgr := worktree.NewManager(r.Root)
			cfg, err := mgr.Get(wtName)
			if err != nil {
				fmtErr("get current worktree: %v", err)
				os.Exit(1)
			}
			if cfg.HeadSnapshotID == "" {
				fmtErr("current worktree has no snapshots to fork from")
				os.Exit(1)
			}
			snapshotID = cfg.HeadSnapshotID
			name = "" // auto-generate

		case 1:
			// One arg: could be snapshot-id or name
			// Try to interpret as snapshot-id first
			arg := args[0]
			testID := model.SnapshotID(arg)

			// Check if it looks like a valid snapshot ID or can be resolved
			_, err := snapshot.LoadDescriptor(r.Root, testID)
			if err != nil {
				// Try fuzzy match
				desc, fuzzyErr := snapshot.FindOne(r.Root, arg)
				if fuzzyErr == nil {
					// It's a snapshot reference
					snapshotID = desc.SnapshotID
					name = "" // auto-generate
				} else {
					// Not a snapshot, treat as name, use current position
					mgr := worktree.NewManager(r.Root)
					cfg, err := mgr.Get(wtName)
					if err != nil {
						fmtErr("get current worktree: %v", err)
						os.Exit(1)
					}
					if cfg.HeadSnapshotID == "" {
						fmtErr("current worktree has no snapshots to fork from")
						os.Exit(1)
					}
					snapshotID = cfg.HeadSnapshotID
					name = arg
				}
			} else {
				// Valid snapshot ID
				snapshotID = testID
				name = "" // auto-generate
			}

		case 2:
			// Two args: snapshot-id and name
			snapshotArg := args[0]
			name = args[1]

			testID := model.SnapshotID(snapshotArg)
			_, err := snapshot.LoadDescriptor(r.Root, testID)
			if err != nil {
				// Try fuzzy match
				desc, fuzzyErr := snapshot.FindOne(r.Root, snapshotArg)
				if fuzzyErr != nil {
					fmtErr("snapshot not found: %v (fuzzy search: %v)", err, fuzzyErr)
					os.Exit(1)
				}
				snapshotID = desc.SnapshotID
			} else {
				snapshotID = testID
			}
		}

		// Auto-generate name if not provided
		if name == "" {
			name = fmt.Sprintf("fork-%s", snapshotID.ShortID())
		}

		// Verify snapshot exists and is valid
		if err := snapshot.VerifySnapshot(r.Root, snapshotID, false); err != nil {
			fmtErr("verify snapshot: %v", err)
			os.Exit(1)
		}

		// Create engine for cloning (use copy engine as default)
		eng := engine.NewEngine(model.EngineCopy)

		// Fork the worktree
		mgr := worktree.NewManager(r.Root)
		cfg, err := mgr.Fork(snapshotID, name, func(src, dst string) error {
			_, err := eng.Clone(src, dst)
			return err
		})
		if err != nil {
			fmtErr("fork worktree: %v", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(cfg)
		} else {
			fmt.Printf("Created worktree '%s' from snapshot %s\n", name, snapshotID)
			fmt.Printf("Path: %s\n", mgr.Path(name))
			fmt.Println("Worktree is at HEAD state - you can create snapshots.")
		}
	},
}

func init() {
	worktreeCreateCmd.Flags().StringVar(&worktreeCreateFrom, "from", "", "create from snapshot (ID, tag, or note prefix)")
	worktreeRemoveCmd.Flags().BoolVarP(&worktreeForce, "force", "f", false, "force removal even if in detached state")
	worktreeCmd.AddCommand(worktreeCreateCmd)
	worktreeCmd.AddCommand(worktreeListCmd)
	worktreeCmd.AddCommand(worktreePathCmd)
	worktreeCmd.AddCommand(worktreeRenameCmd)
	worktreeCmd.AddCommand(worktreeRemoveCmd)
	worktreeCmd.AddCommand(worktreeForkCmd)
	rootCmd.AddCommand(worktreeCmd)
}

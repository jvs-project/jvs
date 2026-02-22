package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/restore"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/model"
)

var restoreCmd = &cobra.Command{
	Use:   "restore <snapshot-id>",
	Short: "Restore worktree to a historical snapshot",
	Long: `Restore worktree to a historical snapshot.

This replaces the current worktree content with the specified snapshot.
After restore, the worktree enters "detached" state - you cannot create
new snapshots until you either:

  - Create a new worktree from this point: jvs worktree fork <name>
  - Return to the latest state: jvs restore HEAD

The snapshot-id can be:
  - A full snapshot ID
  - A short ID prefix
  - A tag name
  - A note prefix (fuzzy match)
  - "HEAD" to restore to the latest snapshot

Examples:
  jvs restore 1771589366482-abc12345   # Restore to specific snapshot
  jvs restore v1.0                      # Restore by tag
  jvs restore HEAD                      # Restore to latest (exit detached)`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r, wtName := requireWorktree()
		snapshotArg := args[0]

		var snapshotID model.SnapshotID

		// Handle special "HEAD" case
		if snapshotArg == "HEAD" {
			restorer := restore.NewRestorer(r.Root, detectEngine(r.Root))
			if err := restorer.RestoreToLatest(wtName); err != nil {
				fmtErr("restore to latest: %v", err)
				os.Exit(1)
			}

			// Load config to get the snapshot ID for output
			wtMgr := worktree.NewManager(r.Root)
			cfg, _ := wtMgr.Get(wtName)

			if jsonOutput {
				outputJSON(map[string]string{
					"status":      "restored",
					"snapshot_id": string(cfg.HeadSnapshotID),
					"detached":    "false",
				})
			} else {
				fmt.Printf("Restored to latest snapshot %s\n", cfg.HeadSnapshotID)
				fmt.Println("Worktree is now at HEAD state.")
			}
			return
		}

		// Try to resolve the snapshot ID
		snapshotID = model.SnapshotID(snapshotArg)

		// Check if it's a valid snapshot ID (exists directly)
		_, err := snapshot.LoadDescriptor(r.Root, snapshotID)
		if err != nil {
			// Try fuzzy match by note/tag/ID prefix
			desc, fuzzyErr := snapshot.FindOne(r.Root, snapshotArg)
			if fuzzyErr != nil {
				fmtErr("snapshot not found: %v (fuzzy search: %v)", err, fuzzyErr)
				os.Exit(1)
			}
			snapshotID = desc.SnapshotID
		}

		// Perform restore
		restorer := restore.NewRestorer(r.Root, detectEngine(r.Root))
		if err := restorer.Restore(wtName, snapshotID); err != nil {
			fmtErr("restore: %v", err)
			os.Exit(1)
		}

		// Check if we're now detached
		wtMgr := worktree.NewManager(r.Root)
		cfg, _ := wtMgr.Get(wtName)
		isDetached := cfg.IsDetached()

		if jsonOutput {
			outputJSON(map[string]interface{}{
				"status":      "restored",
				"snapshot_id": string(snapshotID),
				"detached":    isDetached,
			})
		} else {
			fmt.Printf("Restored to snapshot %s\n", snapshotID)
			if isDetached {
				fmt.Println("Worktree is now in DETACHED state.")
				fmt.Println("To continue working from here: jvs worktree fork <name>")
				fmt.Println("To return to latest: jvs restore HEAD")
			} else {
				fmt.Println("Worktree is now at HEAD state.")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}

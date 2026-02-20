package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/restore"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/pkg/model"
)

var (
	restoreInplace   bool
	restoreForce     bool
	restoreReason    string
	restoreLatestTag string
)

var restoreCmd = &cobra.Command{
	Use:   "restore <snapshot-id>",
	Short: "Restore a snapshot",
	Long: `Restore a snapshot.

By default, creates a new worktree from the snapshot (safe restore).
Use --inplace --force --reason <text> to overwrite the current worktree.

The snapshot-id can be:
- A full snapshot ID
- A short ID prefix
- A tag name (with --latest-tag or as fuzzy match)
- A note prefix (fuzzy match)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r, _ := requireWorktree()

		var snapshotID model.SnapshotID

		if restoreLatestTag != "" {
			// Find latest snapshot with tag
			desc, err := snapshot.FindByTag(r.Root, restoreLatestTag)
			if err != nil {
				fmtErr("find snapshot by tag: %v", err)
				os.Exit(1)
			}
			snapshotID = desc.SnapshotID
		} else if len(args) == 0 {
			fmtErr("snapshot-id or --latest-tag required")
			os.Exit(1)
		} else {
			// Try to resolve the snapshot ID
			snapshotID = model.SnapshotID(args[0])

			// Check if it's a valid snapshot ID (exists directly)
			_, err := snapshot.LoadDescriptor(r.Root, snapshotID)
			if err != nil {
				// Try fuzzy match by note/tag/ID prefix
				desc, fuzzyErr := snapshot.FindOne(r.Root, args[0])
				if fuzzyErr != nil {
					fmtErr("snapshot not found: %v (fuzzy search: %v)", err, fuzzyErr)
					os.Exit(1)
				}
				snapshotID = desc.SnapshotID
			}
		}

		restorer := restore.NewRestorer(r.Root, model.EngineCopy)

		if restoreInplace {
			if !restoreForce {
				fmtErr("--inplace requires --force")
				os.Exit(1)
			}
			if restoreReason == "" {
				fmtErr("--inplace requires --reason")
				os.Exit(1)
			}

			if err := restorer.InplaceRestore(snapshotID, restoreReason); err != nil {
				fmtErr("inplace restore: %v", err)
				os.Exit(1)
			}

			if !jsonOutput {
				fmt.Printf("Restored snapshot %s (inplace)\n", snapshotID)
			}
		} else {
			// Safe restore - create new worktree
			cfg, err := restorer.SafeRestore(snapshotID, "", nil)
			if err != nil {
				fmtErr("restore: %v", err)
				os.Exit(1)
			}

			if jsonOutput {
				outputJSON(cfg)
			} else {
				fmt.Printf("Restored snapshot %s to new worktree '%s'\n", snapshotID, cfg.Name)
			}
		}
	},
}

func init() {
	restoreCmd.Flags().BoolVar(&restoreInplace, "inplace", false, "overwrite current worktree (dangerous)")
	restoreCmd.Flags().BoolVar(&restoreForce, "force", false, "force dangerous operation")
	restoreCmd.Flags().StringVar(&restoreReason, "reason", "", "reason for inplace restore (required with --inplace)")
	restoreCmd.Flags().StringVar(&restoreLatestTag, "latest-tag", "", "restore latest snapshot with this tag")
	rootCmd.AddCommand(restoreCmd)
}

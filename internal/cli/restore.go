package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/lock"
	"github.com/jvs-project/jvs/internal/restore"
	"github.com/jvs-project/jvs/pkg/model"
)

var (
	restoreInplace bool
	restoreForce   bool
	restoreReason  string
)

var restoreCmd = &cobra.Command{
	Use:   "restore <snapshot-id>",
	Short: "Restore a snapshot",
	Long: `Restore a snapshot.

By default, creates a new worktree from the snapshot (safe restore).
Use --inplace --force --reason <text> to overwrite the current worktree.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r, wtName := requireWorktree()
		snapshotID := model.SnapshotID(args[0])

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

			// Get fencing token from current lock
			lockMgr := lock.NewManager(r.Root, model.LockPolicy{})
			sess, err := lockMgr.LoadSession(wtName)
			if err != nil {
				fmtErr("no active lock session (run 'jvs lock acquire' first)")
				os.Exit(1)
			}
			state, rec, err := lockMgr.Status(wtName)
			if err != nil || state != model.LockStateHeld || rec.HolderNonce != sess.HolderNonce {
				fmtErr("must hold lock for inplace restore")
				os.Exit(1)
			}

			if err := restorer.InplaceRestore(wtName, snapshotID, rec.FencingToken, restoreReason); err != nil {
				fmtErr("inplace restore: %v", err)
				os.Exit(1)
			}

			if !jsonOutput {
				fmt.Printf("Restored snapshot %s to worktree '%s' (inplace)\n", snapshotID, wtName)
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
	rootCmd.AddCommand(restoreCmd)
}

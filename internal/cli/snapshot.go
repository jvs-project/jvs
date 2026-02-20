package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/lock"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/pkg/model"
)

var (
	snapshotConsistency string
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot [note]",
	Short: "Create a snapshot of the current worktree",
	Long: `Create a snapshot of the current worktree.

The worktree must be locked before creating a snapshot.
Use --consistency to specify the consistency level (quiesced or best_effort).`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r, wtName := requireWorktree()

		note := ""
		if len(args) > 0 {
			note = args[0]
		}

		consistency := model.ConsistencyQuiesced
		if snapshotConsistency == "best_effort" {
			consistency = model.ConsistencyBestEffort
		}

		// Check lock and get fencing token
		lockMgr := lock.NewManager(r.Root, model.LockPolicy{})
		state, rec, err := lockMgr.Status(wtName)
		if err != nil {
			fmtErr("check lock status: %v", err)
			os.Exit(1)
		}
		if state != model.LockStateHeld {
			fmtErr("worktree %s is not locked (run 'jvs lock acquire' first)", wtName)
			os.Exit(1)
		}

		// Load session to get nonce
		sess, err := lockMgr.LoadSession(wtName)
		if err != nil || sess.HolderNonce != rec.HolderNonce {
			fmtErr("lock session mismatch (run 'jvs lock acquire' from this terminal)")
			os.Exit(1)
		}

		creator := snapshot.NewCreator(r.Root, model.EngineCopy)
		desc, err := creator.Create(wtName, note, consistency, rec.FencingToken)
		if err != nil {
			fmtErr("create snapshot: %v", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(desc)
		} else {
			fmt.Printf("Created snapshot %s\n", desc.SnapshotID)
		}
	},
}

func init() {
	snapshotCmd.Flags().StringVar(&snapshotConsistency, "consistency", "quiesced", "consistency level (quiesced|best_effort)")
	rootCmd.AddCommand(snapshotCmd)
}

package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/lock"
	"github.com/jvs-project/jvs/pkg/model"
)

var (
	lockPurpose string
)

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Manage worktree locks",
}

var lockAcquireCmd = &cobra.Command{
	Use:   "acquire",
	Short: "Acquire an exclusive lock on the current worktree",
	Run: func(cmd *cobra.Command, args []string) {
		r, wtName := requireWorktree()

		mgr := lock.NewManager(r.Root, model.LockPolicy{DefaultLeaseTTL: 5 * time.Minute})
		rec, err := mgr.Acquire(wtName, lockPurpose)
		if err != nil {
			fmtErr("acquire lock: %v", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(rec)
		} else {
			fmt.Printf("Lock acquired on worktree '%s'\n", wtName)
			fmt.Printf("  Session ID: %s\n", rec.SessionID)
			fmt.Printf("  Expires: %s\n", rec.ExpiresAt.Format(time.RFC3339))
		}
	},
}

var lockStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show lock status for the current worktree",
	Run: func(cmd *cobra.Command, args []string) {
		r, wtName := requireWorktree()

		mgr := lock.NewManager(r.Root, model.LockPolicy{})
		state, rec, err := mgr.Status(wtName)
		if err != nil {
			fmtErr("check lock status: %v", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(map[string]any{
				"worktree": wtName,
				"state":    state,
				"lock":     rec,
			})
		} else {
			fmt.Printf("Worktree: %s\n", wtName)
			fmt.Printf("Lock state: %s\n", state)
			if rec != nil {
				fmt.Printf("  Holder: %s\n", rec.HolderNonce[:8]+"...")
				fmt.Printf("  Acquired: %s\n", rec.AcquiredAt.Format(time.RFC3339))
				fmt.Printf("  Expires: %s\n", rec.ExpiresAt.Format(time.RFC3339))
			}
		}
	},
}

var lockRenewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Renew the lock on the current worktree",
	Run: func(cmd *cobra.Command, args []string) {
		r, wtName := requireWorktree()

		mgr := lock.NewManager(r.Root, model.LockPolicy{})
		sess, err := mgr.LoadSession(wtName)
		if err != nil {
			fmtErr("no active lock session (run 'jvs lock acquire' first)")
			os.Exit(1)
		}

		rec, err := mgr.Renew(wtName, sess.HolderNonce)
		if err != nil {
			fmtErr("renew lock: %v", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(rec)
		} else {
			fmt.Printf("Lock renewed, expires: %s\n", rec.ExpiresAt.Format(time.RFC3339))
		}
	},
}

var lockReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release the lock on the current worktree",
	Run: func(cmd *cobra.Command, args []string) {
		r, wtName := requireWorktree()

		mgr := lock.NewManager(r.Root, model.LockPolicy{})
		sess, err := mgr.LoadSession(wtName)
		if err != nil {
			fmtErr("no active lock session")
			os.Exit(1)
		}

		if err := mgr.Release(wtName, sess.HolderNonce); err != nil {
			fmtErr("release lock: %v", err)
			os.Exit(1)
		}

		if !jsonOutput {
			fmt.Printf("Lock released on worktree '%s'\n", wtName)
		}
	},
}

func init() {
	lockAcquireCmd.Flags().StringVar(&lockPurpose, "purpose", "", "purpose for acquiring lock")
	lockCmd.AddCommand(lockAcquireCmd)
	lockCmd.AddCommand(lockStatusCmd)
	lockCmd.AddCommand(lockRenewCmd)
	lockCmd.AddCommand(lockReleaseCmd)
	rootCmd.AddCommand(lockCmd)
}

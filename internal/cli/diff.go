package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/diff"
	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/model"
)

var (
	diffStatOnly  bool
)

var diffCmd = &cobra.Command{
	Use:   "diff [<from> [<to>]]",
	Short: "Show differences between snapshots",
	Long: `Show differences between two snapshots.

If only one argument is provided, compares that snapshot with the current worktree state.
If no arguments are provided, compares the two most recent snapshots.

Arguments can be:
- Full snapshot ID
- Short ID prefix (must be unique)
- Tag name
- HEAD (latest snapshot of current worktree)`,
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()

		// Parse arguments
		var fromID, toID model.SnapshotID

		switch len(args) {
		case 0:
			// Compare two most recent snapshots
			snapshots, err := snapshot.ListAll(r.Root)
			if err != nil {
				fmtErr("list snapshots: %v", err)
				os.Exit(1)
			}
			if len(snapshots) < 2 {
				fmtErr("need at least 2 snapshots to diff")
				os.Exit(1)
			}
			// ListAll returns newest first
			toID = snapshots[0].SnapshotID
			fromID = snapshots[1].SnapshotID

		case 1:
			// Compare given snapshot with current worktree
			snapID, err := resolveSnapshot(r.Root, args[0])
			if err != nil {
				fmtErr("resolve snapshot: %v", err)
				os.Exit(1)
			}
			fromID = snapID
			// For now, toID is empty means compare against "nothing"
			// In future, we could compare against current worktree state
			toID = fromID // Same snapshot = no diff
			fmt.Println("Note: Comparing snapshot against itself (worktree comparison not yet implemented)")

		case 2:
			// Compare two specific snapshots
			from, err := resolveSnapshot(r.Root, args[0])
			if err != nil {
				fmtErr("resolve from snapshot: %v", err)
				os.Exit(1)
			}
			to, err := resolveSnapshot(r.Root, args[1])
			if err != nil {
				fmtErr("resolve to snapshot: %v", err)
				os.Exit(1)
			}
			fromID = from
			toID = to

		default:
			fmtErr("too many arguments")
			os.Exit(1)
		}

		// Load descriptors for timestamps
		var fromTime, toTime time.Time
		if fromID != "" {
			fromDesc, err := snapshot.LoadDescriptor(r.Root, fromID)
			if err == nil {
				fromTime = fromDesc.CreatedAt
			}
		}
		if toID != "" {
			toDesc, err := snapshot.LoadDescriptor(r.Root, toID)
			if err == nil {
				toTime = toDesc.CreatedAt
			}
		}

		// Compute diff
		differ := diff.NewDiffer(r.Root)
		result, err := differ.Diff(fromID, toID)
		if err != nil {
			fmtErr("compute diff: %v", err)
			os.Exit(1)
		}

		// Set timestamps
		result.SetTimes(fromTime, toTime)

		if jsonOutput {
			outputJSON(result)
			return
		}

		if diffStatOnly {
			// Print summary only
			fmt.Printf("Added: %d, Removed: %d, Modified: %d\n",
				result.TotalAdded, result.TotalRemoved, result.TotalModified)
		} else {
			// Print full diff
			fmt.Print(result.FormatHuman())
		}
	},
}

// resolveSnapshot resolves a snapshot reference to a full snapshot ID.
func resolveSnapshot(repoRoot string, ref string) (model.SnapshotID, error) {
	// Handle HEAD specially
	if ref == "HEAD" {
		wtMgr := worktree.NewManager(repoRoot)
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get cwd: %w", err)
		}
		_, wtName, err := repo.DiscoverWorktree(cwd)
		if err != nil {
			return "", fmt.Errorf("discover worktree: %w", err)
		}
		if wtName == "" {
			return "", fmt.Errorf("not inside a worktree")
		}
		cfg, err := wtMgr.Get(wtName)
		if err != nil {
			return "", fmt.Errorf("get worktree: %w", err)
		}
		if cfg.HeadSnapshotID == "" {
			return "", fmt.Errorf("no snapshots in current worktree")
		}
		return cfg.HeadSnapshotID, nil
	}

	// Check if it's a tag
	desc, err := snapshot.FindByTag(repoRoot, ref)
	if err == nil {
		return desc.SnapshotID, nil
	}

	// Try fuzzy match by ID or note
	desc, err = snapshot.FindOne(repoRoot, ref)
	if err != nil {
		return "", err
	}
	return desc.SnapshotID, nil
}

func init() {
	diffCmd.Flags().BoolVar(&diffStatOnly, "stat", false, "show summary only")
	rootCmd.AddCommand(diffCmd)
}

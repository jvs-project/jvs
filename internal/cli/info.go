package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/engine"
	"github.com/jvs-project/jvs/internal/worktree"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show repository information",
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()

		// Count worktrees and snapshots
		wtMgr := worktree.NewManager(r.Root)
		wtList, _ := wtMgr.List()

		snapshotsDir := r.Root + "/.jvs/snapshots"
		entries, _ := os.ReadDir(snapshotsDir)
		snapshotCount := len(entries)

		eng, _ := engine.DetectEngine(r.Root)
		snapshotEngine := string(eng.Name())

		info := map[string]any{
			"repo_root":       r.Root,
			"repo_id":         r.RepoID,
			"format_version":  r.FormatVersion,
			"snapshot_engine": snapshotEngine,
			"total_worktrees": len(wtList),
			"total_snapshots": snapshotCount,
		}

		if jsonOutput {
			outputJSON(info)
			return
		}

		fmt.Printf("Repository: %s\n", r.Root)
		fmt.Printf("  Repo ID: %s\n", r.RepoID)
		fmt.Printf("  Format version: %d\n", r.FormatVersion)
		fmt.Printf("  Snapshot engine: %s\n", snapshotEngine)
		fmt.Printf("  Worktrees: %d\n", len(wtList))
		fmt.Printf("  Snapshots: %d\n", snapshotCount)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

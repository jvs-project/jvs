package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/jvs-project/jvs/pkg/pathutil"
)

var (
	snapshotTags  []string
	snapshotPaths []string
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot [note] [-- <paths>...]",
	Short: "Create a snapshot of the current worktree",
	Long: `Create a snapshot of the current worktree.

Captures the current state of the worktree at a point in time.
Use --tag to attach one or more tags to the snapshot.

For partial snapshots of specific paths, use -- followed by paths:
  jvs snapshot "models update" -- models/ data/

NOTE: Cannot create snapshots in detached state. Use 'jvs worktree fork'
to create a new worktree from the current position first.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r, wtName := requireWorktree()

		// Check if worktree is in detached state
		wtMgr := worktree.NewManager(r.Root)
		cfg, err := wtMgr.Get(wtName)
		if err != nil {
			fmtErr("get worktree: %v", err)
			os.Exit(1)
		}

		if cfg.IsDetached() {
			fmtErr("cannot create snapshot in detached state")
			fmt.Println()
			fmt.Printf("You are currently at snapshot '%s' (historical).\n", cfg.HeadSnapshotID)
			fmt.Println("To continue working from this point:")
			fmt.Println()
			fmt.Printf("    jvs worktree fork %s <new-worktree-name>\n", cfg.HeadSnapshotID.ShortID())
			fmt.Println()
			fmt.Println("Or return to the latest state:")
			fmt.Println()
			fmt.Println("    jvs restore HEAD")
			os.Exit(1)
		}

		note := ""
		if len(args) > 0 {
			note = args[0]
		}

		// Validate tags
		for _, tag := range snapshotTags {
			if err := pathutil.ValidateTag(tag); err != nil {
				fmtErr("invalid tag %q: %v", tag, err)
				os.Exit(1)
			}
		}

		creator := snapshot.NewCreator(r.Root, detectEngine(r.Root))
		var desc *model.Descriptor

		if len(snapshotPaths) > 0 {
			// Partial snapshot
			desc, err = creator.CreatePartial(wtName, note, snapshotTags, snapshotPaths)
		} else {
			// Full snapshot
			desc, err = creator.Create(wtName, note, snapshotTags)
		}

		if err != nil {
			fmtErr("create snapshot: %v", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(desc)
		} else {
			if len(snapshotPaths) > 0 {
				fmt.Printf("Created partial snapshot %s (%d paths)\n", desc.SnapshotID, len(snapshotPaths))
			} else {
				fmt.Printf("Created snapshot %s\n", desc.SnapshotID)
			}
		}
	},
}

func init() {
	snapshotCmd.Flags().StringSliceVar(&snapshotTags, "tag", []string{}, "tag for this snapshot (can be repeated)")
	snapshotCmd.Flags().StringSliceVar(&snapshotPaths, "paths", []string{}, "paths to include in partial snapshot")
	rootCmd.AddCommand(snapshotCmd)
}

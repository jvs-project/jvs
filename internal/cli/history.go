package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/pkg/model"
)

var (
	historyLimit int
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show snapshot history for the current worktree",
	Long: `Show snapshot history for the current worktree.

Traverses the lineage chain from head backwards.`,
	Run: func(cmd *cobra.Command, args []string) {
		r, wtName := requireWorktree()

		cfg, err := repo.LoadWorktreeConfig(r.Root, wtName)
		if err != nil {
			fmtErr("load worktree config: %v", err)
			os.Exit(1)
		}

		if cfg.HeadSnapshotID == "" {
			if jsonOutput {
				outputJSON([]any{})
			} else {
				fmt.Println("No snapshots yet.")
			}
			return
		}

		var history []*model.Descriptor
		currentID := &cfg.HeadSnapshotID
		count := 0

		for currentID != nil && (historyLimit == 0 || count < historyLimit) {
			desc, err := snapshot.LoadDescriptor(r.Root, *currentID)
			if err != nil {
				break
			}
			history = append(history, desc)
			currentID = desc.ParentID
			count++
		}

		if jsonOutput {
			outputJSON(history)
			return
		}

		for _, desc := range history {
			note := desc.Note
			if note == "" {
				note = "(no note)"
			}
			fmt.Printf("%s  %s  %s\n", desc.SnapshotID, desc.CreatedAt.Format("2006-01-02 15:04"), note)
		}
	},
}

func init() {
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "n", 0, "limit number of entries (0 = all)")
	rootCmd.AddCommand(historyCmd)
}

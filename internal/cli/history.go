package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/pkg/model"
)

var (
	historyLimit      int
	historyNoteFilter string
	historyTagFilter  string
	historyAll        bool
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show snapshot history for the current worktree",
	Long: `Show snapshot history for the current worktree.

Traverses the lineage chain from head backwards.
Use --all to show all snapshots in the repository, not just the current worktree's lineage.`,
	Run: func(cmd *cobra.Command, args []string) {
		r, wtName := requireWorktree()

		var history []*model.Descriptor

		if historyAll {
			// Show all snapshots with optional filtering
			opts := snapshot.FilterOptions{
				NoteContains: historyNoteFilter,
				HasTag:       historyTagFilter,
			}
			var err error
			history, err = snapshot.Find(r.Root, opts)
			if err != nil {
				fmtErr("list snapshots: %v", err)
				os.Exit(1)
			}
		} else {
			// Show lineage for current worktree
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

			currentID := &cfg.HeadSnapshotID
			count := 0

			for currentID != nil && (historyLimit == 0 || count < historyLimit) {
				desc, err := snapshot.LoadDescriptor(r.Root, *currentID)
				if err != nil {
					break
				}

				// Apply filters
				if historyNoteFilter != "" && !strings.Contains(desc.Note, historyNoteFilter) {
					currentID = desc.ParentID
					continue
				}
				if historyTagFilter != "" && !hasTag(desc, historyTagFilter) {
					currentID = desc.ParentID
					continue
				}

				history = append(history, desc)
				currentID = desc.ParentID
				count++
			}
		}

		if jsonOutput {
			outputJSON(history)
			return
		}

		if len(history) == 0 {
			fmt.Println("No snapshots found.")
			return
		}

		for _, desc := range history {
			note := desc.Note
			if note == "" {
				note = "(no note)"
			}
			tagsStr := ""
			if len(desc.Tags) > 0 {
				tagsStr = "  [" + strings.Join(desc.Tags, ",") + "]"
			}
			fmt.Printf("%s  %s  %s%s\n", desc.SnapshotID.ShortID(), desc.CreatedAt.Format("2006-01-02 15:04"), note, tagsStr)
		}
	},
}

func hasTag(desc *model.Descriptor, tag string) bool {
	for _, t := range desc.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func init() {
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "n", 0, "limit number of entries (0 = all)")
	historyCmd.Flags().StringVarP(&historyNoteFilter, "grep", "g", "", "filter by note substring")
	historyCmd.Flags().StringVar(&historyTagFilter, "tag", "", "filter by tag")
	historyCmd.Flags().BoolVar(&historyAll, "all", false, "show all snapshots (not just current worktree)")
	rootCmd.AddCommand(historyCmd)
}

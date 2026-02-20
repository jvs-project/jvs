package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/ref"
	"github.com/jvs-project/jvs/pkg/model"
)

var refCmd = &cobra.Command{
	Use:   "ref",
	Short: "Manage named snapshot references",
}

var refCreateCmd = &cobra.Command{
	Use:   "create <name> <snapshot-id>",
	Short: "Create a named reference to a snapshot",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()
		name := args[0]
		snapshotID := model.SnapshotID(args[1])

		mgr := ref.NewManager(r.Root)
		rec, err := mgr.Create(name, snapshotID, "")
		if err != nil {
			fmtErr("create ref: %v", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(rec)
		} else {
			fmt.Printf("Created ref '%s' -> %s\n", name, snapshotID)
		}
	},
}

var refListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all refs",
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()

		mgr := ref.NewManager(r.Root)
		list, err := mgr.List()
		if err != nil {
			fmtErr("list refs: %v", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(list)
			return
		}

		for _, rec := range list {
			desc := rec.Description
			if desc == "" {
				desc = "(no description)"
			}
			fmt.Printf("%-20s  %s  %s\n", rec.Name, rec.TargetID, desc)
		}
	},
}

var refDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a ref",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()
		name := args[0]

		mgr := ref.NewManager(r.Root)
		if err := mgr.Delete(name); err != nil {
			fmtErr("delete ref: %v", err)
			os.Exit(1)
		}

		if !jsonOutput {
			fmt.Printf("Deleted ref '%s'\n", name)
		}
	},
}

func init() {
	refCmd.AddCommand(refCreateCmd)
	refCmd.AddCommand(refListCmd)
	refCmd.AddCommand(refDeleteCmd)
	rootCmd.AddCommand(refCmd)
}

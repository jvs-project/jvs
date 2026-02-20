package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/worktree"
)

var worktreeCmd = &cobra.Command{
	Use:     "worktree",
	Short:   "Manage worktrees",
	Aliases: []string{"wt"},
}

var worktreeCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new worktree",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()
		name := args[0]

		mgr := worktree.NewManager(r.Root)
		cfg, err := mgr.Create(name, nil)
		if err != nil {
			fmtErr("create worktree: %v", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(cfg)
		} else {
			fmt.Printf("Created worktree '%s' at %s\n", name, mgr.Path(name))
		}
	},
}

var worktreeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all worktrees",
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()

		mgr := worktree.NewManager(r.Root)
		list, err := mgr.List()
		if err != nil {
			fmtErr("list worktrees: %v", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(list)
			return
		}

		for _, cfg := range list {
			head := string(cfg.HeadSnapshotID)
			if head == "" {
				head = "(none)"
			} else if len(head) > 16 {
				head = head[:16] + "..."
			}
			fmt.Printf("%-20s  %s\n", cfg.Name, head)
		}
	},
}

var worktreePathCmd = &cobra.Command{
	Use:   "path [<name>]",
	Short: "Print the path to a worktree",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()

		name := ""
		if len(args) > 0 {
			name = args[0]
		} else {
			_, name = requireWorktree()
		}

		mgr := worktree.NewManager(r.Root)
		path := mgr.Path(name)
		fmt.Println(path)
	},
}

var worktreeRenameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename a worktree",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()
		oldName := args[0]
		newName := args[1]

		mgr := worktree.NewManager(r.Root)
		if err := mgr.Rename(oldName, newName); err != nil {
			fmtErr("rename worktree: %v", err)
			os.Exit(1)
		}

		if !jsonOutput {
			fmt.Printf("Renamed worktree '%s' to '%s'\n", oldName, newName)
		}
	},
}

var worktreeRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a worktree",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()
		name := args[0]

		mgr := worktree.NewManager(r.Root)
		if err := mgr.Remove(name); err != nil {
			fmtErr("remove worktree: %v", err)
			os.Exit(1)
		}

		if !jsonOutput {
			fmt.Printf("Removed worktree '%s'\n", name)
		}
	},
}

func init() {
	worktreeCmd.AddCommand(worktreeCreateCmd)
	worktreeCmd.AddCommand(worktreeListCmd)
	worktreeCmd.AddCommand(worktreePathCmd)
	worktreeCmd.AddCommand(worktreeRenameCmd)
	worktreeCmd.AddCommand(worktreeRemoveCmd)
	rootCmd.AddCommand(worktreeCmd)
}

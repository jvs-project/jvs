package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/pkg/color"
	"github.com/jvs-project/jvs/pkg/pathutil"
)

var initCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Initialize a new JVS repository",
	Long: `Initialize a new JVS repository in a directory named <name>.

This creates:
  - .jvs/ directory with all metadata structures
  - main/ worktree as the primary payload directory
  - format_version file (version 1)`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if err := pathutil.ValidateName(name); err != nil {
			fmtErr("%v", err)
			os.Exit(1)
		}

		cwd, _ := os.Getwd()
		repoPath := filepath.Join(cwd, name)

		r, err := repo.Init(repoPath, name)
		if err != nil {
			fmtErr("failed to initialize repository: %v", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(map[string]any{
				"repo_root":      r.Root,
				"format_version": r.FormatVersion,
				"repo_id":        r.RepoID,
			})
		} else {
			fmt.Printf("Initialized JVS repository in %s\n", color.Success(repoPath))
			fmt.Printf("  Main worktree: %s/main\n", color.Highlight(repoPath))
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

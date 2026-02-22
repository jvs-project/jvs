package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/pkg/logging"
)

var (
	jsonOutput  bool
	debugOutput bool
	rootCmd     = &cobra.Command{
		Use:   "jvs",
		Short: "JVS - Juicy Versioned Workspaces",
		Long: `JVS is a snapshot-first, filesystem-native workspace versioning system
built on JuiceFS. It provides atomic snapshots, detached state navigation,
and exclusive-mode worktree isolation.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Configure logging based on debug flag
			if debugOutput {
				logging.SetGlobal(logging.NewLogger(logging.LevelDebug))
			}
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&debugOutput, "debug", false, "enable debug logging")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// outputJSON prints v as JSON if --json flag is set, otherwise does nothing.
func outputJSON(v any) error {
	if !jsonOutput {
		return nil
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// outputJSONOrError prints v as JSON if --json flag is set, or prints error.
func outputJSONOrError(v any, err error) error {
	if err != nil {
		return err
	}
	return outputJSON(v)
}

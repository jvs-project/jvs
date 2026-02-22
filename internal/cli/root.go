package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/pkg/color"
	"github.com/jvs-project/jvs/pkg/logging"
	"github.com/jvs-project/jvs/pkg/progress"
)

var (
	jsonOutput   bool
	debugOutput  bool
	noProgress   bool
	noColor      bool
	rootCmd      = &cobra.Command{
		Use:   "jvs",
		Short: "JVS - Juicy Versioned Workspaces",
		Long: `JVS is a snapshot-first, filesystem-native workspace versioning system
built on JuiceFS. It provides atomic snapshots, detached state navigation,
and exclusive-mode worktree isolation.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Configure color output first (before any output)
			color.Init(noColor)

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
	rootCmd.PersistentFlags().BoolVar(&noProgress, "no-progress", false, "disable progress bars")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output (also respects NO_COLOR env var)")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// progressEnabled returns whether progress bars should be shown.
func progressEnabled() bool {
	return !noProgress && !jsonOutput
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

// newProgressCallback creates a progress callback if progress is enabled.
func newProgressCallback(op string, total int) progress.Callback {
	if !progressEnabled() {
		return progress.Noop
	}
	term := progress.NewTerminal(op, total, true)
	cb := term.Callback()
	// Return a callback that wraps the terminal callback
	return func(op string, current, total int, message string) {
		cb(op, current, total, message)
		if current == total {
			term.Done(message)
		}
	}
}

// newCountingProgress creates a counting progress bar for operations with unknown total.
func newCountingProgress(op string) *progress.CountingTerminal {
	return progress.NewCountingTerminal(op, progressEnabled())
}

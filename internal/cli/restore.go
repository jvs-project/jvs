package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/restore"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/model"
)

var (
	restoreInteractive bool
)

var restoreCmd = &cobra.Command{
	Use:   "restore <snapshot-id>",
	Short: "Restore worktree to a historical snapshot",
	Long: `Restore worktree to a historical snapshot.

This replaces the current worktree content with the specified snapshot.
After restore, the worktree enters "detached" state - you cannot create
new snapshots until you either:

  - Create a new worktree from this point: jvs worktree fork <name>
  - Return to the latest state: jvs restore HEAD

The snapshot-id can be:
  - A full snapshot ID
  - A short ID prefix
  - A tag name
  - A note prefix (fuzzy match)
  - "HEAD" to restore to the latest snapshot

With --interactive, shows matching snapshots and prompts for confirmation.

Examples:
  jvs restore 1771589366482-abc12345   # Restore to specific snapshot
  jvs restore v1.0                      # Restore by tag
  jvs restore HEAD                      # Restore to latest (exit detached)
  jvs restore --interactive 177         # Interactive fuzzy match`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r, wtName := requireWorktree()
		snapshotArg := args[0]

		var snapshotID model.SnapshotID

		// Handle special "HEAD" case
		if snapshotArg == "HEAD" {
			restorer := restore.NewRestorer(r.Root, detectEngine(r.Root))
			if err := restorer.RestoreToLatest(wtName); err != nil {
				fmtErr("restore to latest: %v", err)
				os.Exit(1)
			}

			// Load config to get the snapshot ID for output
			wtMgr := worktree.NewManager(r.Root)
			cfg, _ := wtMgr.Get(wtName)

			if jsonOutput {
				outputJSON(map[string]string{
					"status":      "restored",
					"snapshot_id": string(cfg.HeadSnapshotID),
					"detached":    "false",
				})
			} else {
				fmt.Printf("Restored to latest snapshot %s\n", cfg.HeadSnapshotID)
				fmt.Println("Worktree is now at HEAD state.")
			}
			return
		}

		// Try to resolve the snapshot ID
		snapshotID = model.SnapshotID(snapshotArg)

		// Check if it's a valid snapshot ID (exists directly)
		_, err := snapshot.LoadDescriptor(r.Root, snapshotID)
		if err != nil {
			// In interactive mode, show fuzzy matches
			if restoreInteractive && !jsonOutput {
				matches, fuzzyErr := snapshot.FindMultiple(r.Root, snapshotArg, 10)
				if fuzzyErr != nil {
					fmtErr("search failed: %v", fuzzyErr)
					os.Exit(1)
				}
				if len(matches) == 0 {
					fmtErr("no snapshots found matching %q", snapshotArg)
					os.Exit(1)
				}

				// Show matches and prompt for selection
				fmt.Println(snapshot.FormatMatchList(matches))

				// If only one match and it's a good match, confirm directly
				if len(matches) == 1 && matches[0].Score >= 700 {
					selected := matches[0]
					fmt.Printf("\nRestore to %s? (%s) [y/N]: ",
						selected.Desc.SnapshotID.ShortID(),
						selected.Desc.Note)
					if !confirm() {
						fmt.Println("Restore cancelled.")
						os.Exit(0)
					}
					snapshotID = selected.Desc.SnapshotID
				} else {
					// Prompt for selection
					fmt.Printf("\nSelect snapshot to restore [1-%d]: ", len(matches))
					choice := readInt(len(matches))

					if choice == 0 {
						fmt.Println("Restore cancelled.")
						os.Exit(0)
					}

					selected := matches[choice-1]
					fmt.Printf("\nRestore to %s? (%s) [y/N]: ",
						selected.Desc.SnapshotID.ShortID(),
						selected.Desc.Note)
					if !confirm() {
						fmt.Println("Restore cancelled.")
						os.Exit(0)
					}
					snapshotID = selected.Desc.SnapshotID
				}
			} else {
				// Non-interactive: try single fuzzy match
				desc, fuzzyErr := snapshot.FindOne(r.Root, snapshotArg)
				if fuzzyErr != nil {
					fmtErr("snapshot not found: %v (fuzzy search: %v)", err, fuzzyErr)
					os.Exit(1)
				}
				snapshotID = desc.SnapshotID
			}
		} else if restoreInteractive && !jsonOutput {
			// Snapshot ID exists, but still confirm in interactive mode
			desc, _ := snapshot.LoadDescriptor(r.Root, snapshotID)
			note := desc.Note
			if note == "" {
				note = "(no note)"
			}
			fmt.Printf("\nRestore to %s? (%s) [y/N]: ", snapshotID, note)
			if !confirm() {
				fmt.Println("Restore cancelled.")
				os.Exit(0)
			}
		}

		// Perform restore
		restorer := restore.NewRestorer(r.Root, detectEngine(r.Root))
		if err := restorer.Restore(wtName, snapshotID); err != nil {
			fmtErr("restore: %v", err)
			os.Exit(1)
		}

		// Check if we're now detached
		wtMgr := worktree.NewManager(r.Root)
		cfg, _ := wtMgr.Get(wtName)
		isDetached := cfg.IsDetached()

		if jsonOutput {
			outputJSON(map[string]interface{}{
				"status":      "restored",
				"snapshot_id": string(snapshotID),
				"detached":    isDetached,
			})
		} else {
			fmt.Printf("\nRestored to snapshot %s\n", snapshotID)
			if isDetached {
				fmt.Println("Worktree is now in DETACHED state.")
				fmt.Println("To continue working from here: jvs worktree fork <name>")
				fmt.Println("To return to latest: jvs restore HEAD")
			} else {
				fmt.Println("Worktree is now at HEAD state.")
			}
		}
	},
}

func init() {
	restoreCmd.Flags().BoolVarP(&restoreInteractive, "interactive", "i", false, "interactive mode with fuzzy matching and confirmation")
	rootCmd.AddCommand(restoreCmd)
}

// confirm prompts the user for yes/no confirmation.
func confirm() bool {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

// readInt reads an integer from stdin within range [1, max].
// Returns 0 if user wants to cancel.
func readInt(max int) int {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)

	if line == "" || line == "0" || strings.ToLower(line) == "cancel" {
		return 0
	}

	var choice int
	if _, err := fmt.Sscanf(line, "%d", &choice); err != nil {
		return 0
	}

	if choice < 1 || choice > max {
		return 0
	}
	return choice
}

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/gc"
)

var (
	gcPlanID string
)

var gcCmd = &cobra.Command{
	Use:   "gc",
	Short: "Garbage collection",
}

var gcPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Create a GC plan",
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()

		collector := gc.NewCollector(r.Root)
		plan, err := collector.Plan()
		if err != nil {
			fmtErr("create gc plan: %v", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(plan)
			return
		}

		fmt.Printf("GC Plan: %s\n", plan.PlanID)
		fmt.Printf("  Protected: %d snapshots\n", len(plan.ProtectedSet))
		fmt.Printf("  To delete: %d snapshots\n", len(plan.ToDelete))
		fmt.Printf("  Estimated reclaim: ~%d MB\n", plan.EstimatedBytes/1024/1024)
		fmt.Println()
		fmt.Printf("Run: jvs gc run --plan-id %s\n", plan.PlanID)
	},
}

var gcRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute a GC plan",
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()

		if gcPlanID == "" {
			fmtErr("--plan-id is required")
			os.Exit(1)
		}

		collector := gc.NewCollector(r.Root)
		if err := collector.Run(gcPlanID); err != nil {
			fmtErr("run gc: %v", err)
			os.Exit(1)
		}

		if !jsonOutput {
			fmt.Println("GC completed successfully.")
		}
	},
}

func init() {
	gcRunCmd.Flags().StringVar(&gcPlanID, "plan-id", "", "plan ID to execute")
	gcCmd.AddCommand(gcPlanCmd)
	gcCmd.AddCommand(gcRunCmd)
	rootCmd.AddCommand(gcCmd)
}

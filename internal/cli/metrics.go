package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/pkg/metrics"
)

var (
	metricsAddr string
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Start Prometheus metrics server",
	Long: `Start a Prometheus metrics server for JVS operations.

This exposes a /metrics endpoint with Prometheus-format metrics about:
- Snapshot operations (count, duration, size)
- Restore operations (count, duration)
- Garbage collection (count, bytes reclaimed)
- Repository statistics (snapshots, worktrees, engine)
- Verification operations

The metrics server runs in the foreground until interrupted.

Examples:
  jvs metrics                    # Start on default port :2112
  jvs metrics --addr :9090       # Start on custom port`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize metrics if not already enabled
		if !metrics.Enabled() {
			metrics.Init()
		}

		fmt.Printf("Starting Prometheus metrics server on %s\n", metricsAddr)
		fmt.Println("Available metrics:")
		fmt.Println("  - jvs_snapshot_total")
		fmt.Println("  - jvs_snapshot_failed_total")
		fmt.Println("  - jvs_snapshot_duration_seconds")
		fmt.Println("  - jvs_snapshot_size_bytes")
		fmt.Println("  - jvs_restore_total")
		fmt.Println("  - jvs_restore_failed_total")
		fmt.Println("  - jvs_restore_duration_seconds")
		fmt.Println("  - jvs_gc_run_total")
		fmt.Println("  - jvs_gc_candidates_deleted_total")
		fmt.Println("  - jvs_gc_bytes_reclaimed")
		fmt.Println("  - jvs_snapshots_total")
		fmt.Println("  - jvs_worktrees_total")
		fmt.Println("  - jvs_engine")
		fmt.Println()
		fmt.Printf("Metrics available at http://%s/metrics\n", metricsAddr)
		fmt.Println("Press Ctrl+C to stop")

		if err := metrics.StartServer(metricsAddr); err != nil {
			fmtErr("metrics server: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	metricsCmd.Flags().StringVarP(&metricsAddr, "addr", "a", ":2112", "address to listen on")
	rootCmd.AddCommand(metricsCmd)
}

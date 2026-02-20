package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/doctor"
)

var (
	doctorStrict bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check repository health",
	Long: `Check repository health.

Runs diagnostic checks on the repository and reports any issues.
Use --strict to include full snapshot integrity verification.`,
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()

		doc := doctor.NewDoctor(r.Root)
		result, err := doc.Check(doctorStrict)
		if err != nil {
			fmtErr("doctor: %v", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(result)
			return
		}

		if len(result.Findings) == 0 {
			fmt.Println("Repository is healthy.")
			return
		}

		fmt.Printf("Findings (%d):\n", len(result.Findings))
		for _, f := range result.Findings {
			fmt.Printf("  [%s] %s: %s\n", f.Severity, f.Category, f.Description)
		}

		if !result.Healthy {
			os.Exit(1)
		}
	},
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorStrict, "strict", false, "include full integrity verification")
	rootCmd.AddCommand(doctorCmd)
}

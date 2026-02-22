package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/pkg/config"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage snapshot templates",
	Long:  `Manage snapshot templates for pre-configured snapshot patterns.`,
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available snapshot templates",
	Long:  `List all available snapshot templates.

Shows both user-defined templates from .jvs/config.yaml and built-in templates.`,
	Run: func(cmd *cobra.Command, args []string) {
		r := requireRepo()

		templates := config.ListTemplates(r.Root)

		if jsonOutput {
			outputJSON(map[string]interface{}{
				"templates": templates,
			})
			return
		}

		if len(templates) == 0 {
			fmt.Println("No templates available.")
			return
		}

		fmt.Println("Available snapshot templates:")
		fmt.Println()

		jvsCfg, _ := config.Load(r.Root)

		for _, name := range templates {
			tmpl := config.ResolveTemplate(r.Root, name)
			if tmpl == nil {
				continue
			}

			// Show if it's user-defined or built-in
			source := "built-in"
			if jvsCfg.GetSnapshotTemplate(name) != nil {
				source = "user-defined"
			}

			fmt.Printf("  %s (%s)\n", name, source)

			// Show note template
			if tmpl.Note != "" {
				fmt.Printf("    Note: %s\n", tmpl.Note)
			}

			// Show tags
			if len(tmpl.Tags) > 0 {
				fmt.Printf("    Tags: %v\n", tmpl.Tags)
			}

			// Show compression
			if tmpl.Compression != "" {
				fmt.Printf("    Compression: %s\n", tmpl.Compression)
			}

			// Show paths if partial snapshot
			if len(tmpl.Paths) > 0 {
				fmt.Printf("    Paths: %v\n", tmpl.Paths)
			}

			fmt.Println()
		}
	},
}

var templateShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show snapshot template details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r, err := repo.Find()
		if err != nil {
			fmtErr("find repository: %v", err)
			os.Exit(1)
		}

		name := args[0]
		tmpl := config.ResolveTemplate(r.Root, name)
		if tmpl == nil {
			fmtErr("template not found: %s", name)
			fmt.Println()
			fmt.Println("Available templates:")
			for _, n := range config.ListTemplates(r.Root) {
				fmt.Printf("  - %s\n", n)
			}
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(tmpl)
			return
		}

		fmt.Printf("Template: %s\n", name)
		fmt.Println()

		if tmpl.Note != "" {
			fmt.Printf("Note: %s\n", tmpl.Note)
		}

		if len(tmpl.Tags) > 0 {
			fmt.Printf("Tags: %v\n", tmpl.Tags)
		}

		if tmpl.Compression != "" {
			fmt.Printf("Compression: %s\n", tmpl.Compression)
		}

		if len(tmpl.Paths) > 0 {
			fmt.Printf("Paths: %v\n", tmpl.Paths)
		} else {
			fmt.Println("Paths: (full snapshot)")
		}
	},
}

func init() {
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateShowCmd)
	rootCmd.AddCommand(templateCmd)
}

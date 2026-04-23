package cmd

import (
	"os"

	"github.com/paulmeller/qwilr-cli/internal/output"
	"github.com/spf13/cobra"
)

var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage Qwilr templates",
}

var templatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		var result []map[string]interface{}
		if err := client.Get("/templates", &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		headers := []string{"ID", "NAME"}
		var rows [][]string
		for _, t := range result {
			rows = append(rows, []string{
				stringVal(t, "id"),
				stringVal(t, "name"),
			})
		}
		output.PrintTable(os.Stdout, headers, rows)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(templatesCmd)
	templatesCmd.AddCommand(templatesListCmd)
}

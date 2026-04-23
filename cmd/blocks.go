package cmd

import (
	"os"

	"github.com/paulmeller/qwilr-cli/internal/output"
	"github.com/spf13/cobra"
)

var blocksCmd = &cobra.Command{
	Use:   "blocks",
	Short: "Manage Qwilr blocks",
}

var blocksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved blocks",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		var result []map[string]interface{}
		if err := client.Get("/blocks", &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		headers := []string{"ID", "NAME", "TYPE"}
		var rows [][]string
		for _, b := range result {
			rows = append(rows, []string{
				stringVal(b, "id"),
				stringVal(b, "name"),
				stringVal(b, "type"),
			})
		}
		output.PrintTable(os.Stdout, headers, rows)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(blocksCmd)
	blocksCmd.AddCommand(blocksListCmd)
}

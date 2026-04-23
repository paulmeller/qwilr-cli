package cmd

import (
	"fmt"
	"os"

	"github.com/paulmeller/qwilr-cli/internal/output"
	"github.com/spf13/cobra"
)

var webhooksCmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Manage webhook subscriptions",
}

var webhooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhook subscriptions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()
		var result []map[string]interface{}
		if err := client.Get(ctx, "/webhooks", &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		headers := []string{"ID", "URL", "EVENT"}
		var rows [][]string
		for _, w := range result {
			rows = append(rows, []string{
				stringVal(w, "id"),
				stringVal(w, "url"),
				stringVal(w, "event"),
			})
		}
		output.PrintTable(os.Stdout, headers, rows)
		return nil
	},
}

var webhooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a webhook subscription",
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := cmd.Flags().GetString("url")
		event, _ := cmd.Flags().GetString("event")

		if url == "" {
			return fmt.Errorf("--url is required")
		}
		if event == "" {
			return fmt.Errorf("--event is required")
		}

		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()

		body := map[string]string{
			"url":   url,
			"event": event,
		}

		var result map[string]interface{}
		if err := client.Post(ctx, "/webhooks", body, &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		fmt.Fprintf(os.Stderr, "Webhook created: %s\n", stringVal(result, "id"))
		fields := []output.KeyValue{
			{Key: "ID", Value: stringVal(result, "id")},
			{Key: "URL", Value: stringVal(result, "url")},
			{Key: "Event", Value: stringVal(result, "event")},
		}
		output.PrintDetail(os.Stdout, fields)
		return nil
	},
}

var webhooksDeleteCmd = &cobra.Command{
	Use:   "delete <webhookId>",
	Short: "Delete a webhook subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()
		if err := client.Delete(ctx, "/webhooks/"+args[0]); err != nil {
			return err
		}
		if !jsonOutput {
			fmt.Fprintf(os.Stderr, "Webhook deleted: %s\n", args[0])
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(webhooksCmd)

	webhooksCreateCmd.Flags().String("url", "", "Webhook callback URL")
	webhooksCreateCmd.Flags().String("event", "", "Event type (pageAccepted, pagePreviewAccepted, pageViewed, pageFirstViewed)")

	webhooksCmd.AddCommand(webhooksListCmd)
	webhooksCmd.AddCommand(webhooksCreateCmd)
	webhooksCmd.AddCommand(webhooksDeleteCmd)
}

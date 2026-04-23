package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/paulmeller/qwilr-cli/internal/api"
	"github.com/paulmeller/qwilr-cli/internal/config"
	"github.com/paulmeller/qwilr-cli/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var pagesCmd = &cobra.Command{
	Use:   "pages",
	Short: "Manage Qwilr pages",
}

var pagesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pages",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		path := fmt.Sprintf("/pages?limit=%d&offset=%d", limit, offset)
		var result []map[string]interface{}
		if err := client.Get(ctx, path, &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		headers := []string{"ID", "NAME", "STATUS", "CREATED"}
		var rows [][]string
		for _, p := range result {
			rows = append(rows, []string{
				stringVal(p, "id"),
				stringVal(p, "name"),
				stringVal(p, "status"),
				stringVal(p, "createdAt"),
			})
		}
		output.PrintTable(os.Stdout, headers, rows)
		return nil
	},
}

var pagesGetCmd = &cobra.Command{
	Use:   "get <pageId>",
	Short: "Get page details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()
		var result map[string]interface{}
		if err := client.Get(ctx, "/pages/"+args[0], &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		fields := []output.KeyValue{
			{Key: "Name", Value: stringVal(result, "name")},
			{Key: "ID", Value: stringVal(result, "id")},
			{Key: "Status", Value: stringVal(result, "status")},
			{Key: "Created", Value: stringVal(result, "createdAt")},
			{Key: "URL", Value: stringVal(result, "url")},
		}
		output.PrintDetail(os.Stdout, fields)
		return nil
	},
}

type pageCreateRequest struct {
	Name      string            `json:"name,omitempty" yaml:"name"`
	Template  string            `json:"template,omitempty" yaml:"template"`
	Published bool              `json:"published,omitempty" yaml:"published"`
	Variables map[string]string `json:"variables,omitempty" yaml:"variables"`
}

var pagesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new page",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()

		var req pageCreateRequest

		fromFile, _ := cmd.Flags().GetString("from")
		if fromFile != "" {
			data, err := os.ReadFile(fromFile)
			if err != nil {
				return fmt.Errorf("reading file %s: %w", fromFile, err)
			}
			if err := yaml.Unmarshal(data, &req); err != nil {
				return fmt.Errorf("parsing file %s: %w", fromFile, err)
			}
		}

		// Flags override file values
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			req.Name = name
		}
		if tmpl, _ := cmd.Flags().GetString("template"); tmpl != "" {
			req.Template = tmpl
		}
		if cmd.Flags().Changed("publish") {
			pub, _ := cmd.Flags().GetBool("publish")
			req.Published = pub
		}

		var result map[string]interface{}
		if err := client.Post(ctx, "/pages", req, &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		fmt.Fprintf(os.Stderr, "Page created: %s\n", stringVal(result, "id"))
		fields := []output.KeyValue{
			{Key: "ID", Value: stringVal(result, "id")},
			{Key: "Name", Value: stringVal(result, "name")},
			{Key: "URL", Value: stringVal(result, "url")},
		}
		output.PrintDetail(os.Stdout, fields)
		return nil
	},
}

var pagesUpdateCmd = &cobra.Command{
	Use:   "update <pageId>",
	Short: "Update a page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()

		body := make(map[string]interface{})
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			body["name"] = name
		}
		if content, _ := cmd.Flags().GetString("content"); content != "" {
			body["content"] = content
		}

		var result map[string]interface{}
		if err := client.Put(ctx, "/pages/"+args[0], body, &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		fmt.Fprintf(os.Stderr, "Page updated: %s\n", args[0])
		return nil
	},
}

var pagesDeleteCmd = &cobra.Command{
	Use:   "delete <pageId>",
	Short: "Delete a page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()
		if err := client.Delete(ctx, "/pages/"+args[0]); err != nil {
			return err
		}
		if !jsonOutput {
			fmt.Fprintf(os.Stderr, "Page deleted: %s\n", args[0])
		}
		return nil
	},
}

var pagesVariablesCmd = &cobra.Command{
	Use:   "variables <pageId>",
	Short: "Set token/substitution values on a page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := cmdContext()
		defer cancel()

		vars := make(map[string]string)

		fromFile, _ := cmd.Flags().GetString("from")
		if fromFile != "" {
			data, err := os.ReadFile(fromFile)
			if err != nil {
				return fmt.Errorf("reading file %s: %w", fromFile, err)
			}
			if err := yaml.Unmarshal(data, &vars); err != nil {
				return fmt.Errorf("parsing file %s: %w", fromFile, err)
			}
		}

		setVals, _ := cmd.Flags().GetStringSlice("set")
		for _, kv := range setVals {
			parts := splitKeyValue(kv)
			if parts == nil {
				return fmt.Errorf("invalid --set value %q — expected key=value", kv)
			}
			vars[parts[0]] = parts[1]
		}

		var result map[string]interface{}
		if err := client.Put(ctx, "/pages/"+args[0]+"/variables", vars, &result); err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(os.Stdout, result)
		}

		fmt.Fprintf(os.Stderr, "Variables updated on page %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pagesCmd)

	pagesListCmd.Flags().Int("limit", 20, "Maximum number of pages to return")
	pagesListCmd.Flags().Int("offset", 0, "Number of pages to skip")

	pagesCreateCmd.Flags().String("name", "", "Page name")
	pagesCreateCmd.Flags().String("template", "", "Template ID to use")
	pagesCreateCmd.Flags().String("from", "", "YAML/JSON file with page definition")
	pagesCreateCmd.Flags().Bool("publish", false, "Publish the page immediately")

	pagesUpdateCmd.Flags().String("name", "", "New page name")
	pagesUpdateCmd.Flags().String("content", "", "New page content")

	pagesVariablesCmd.Flags().String("from", "", "YAML/JSON file with variable values")
	pagesVariablesCmd.Flags().StringSlice("set", nil, "Set variable as key=value")

	pagesCmd.AddCommand(pagesListCmd)
	pagesCmd.AddCommand(pagesGetCmd)
	pagesCmd.AddCommand(pagesCreateCmd)
	pagesCmd.AddCommand(pagesUpdateCmd)
	pagesCmd.AddCommand(pagesDeleteCmd)
	pagesCmd.AddCommand(pagesVariablesCmd)
}

func newClient() (*api.Client, error) {
	token := config.ResolveToken(config.DefaultPath())
	if token == "" {
		return nil, fmt.Errorf("no API token configured — run \"qwilr configure\" or set QWILR_API_TOKEN")
	}
	return api.NewClient(token, ""), nil
}

func stringVal(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case json.Number:
		return val.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

func splitKeyValue(s string) []string {
	for i, c := range s {
		if c == '=' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return nil
}

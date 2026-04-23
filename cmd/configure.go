package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/paulmeller/qwilr-cli/internal/config"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Set up API token and preferences",
	Long:  "Interactively configure your Qwilr API token and default output preferences.",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Qwilr API Token: ")
		token, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading token: %w", err)
		}
		token = strings.TrimSpace(token)
		if token == "" {
			return fmt.Errorf("API token cannot be empty")
		}

		fmt.Print("Default output format (text/json) [text]: ")
		format, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading format: %w", err)
		}
		format = strings.TrimSpace(format)
		if format == "" {
			format = "text"
		}
		if format != "text" && format != "json" {
			return fmt.Errorf("invalid output format %q — must be \"text\" or \"json\"", format)
		}

		cfg := &config.Config{
			APIToken:      token,
			DefaultOutput: format,
		}
		path := config.DefaultPath()
		if err := config.Save(cfg, path); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Configuration saved to %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}

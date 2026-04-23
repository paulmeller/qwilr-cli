package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var jsonOutput bool
var timeout time.Duration

var rootCmd = &cobra.Command{
	Use:           "qwilr",
	Short:         "CLI for the Qwilr API",
	Long:          "Manage Qwilr pages, templates, blocks, and webhooks from the command line.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		if isValidationError(err) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func isValidationError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "is required") ||
		strings.Contains(msg, "cannot be empty") ||
		strings.Contains(msg, "invalid") ||
		strings.Contains(msg, "no API token configured")
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second, "API request timeout")
}

func cmdContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	home := os.Getenv("HOME")
	viper.AddConfigPath(strings.Join([]string{home, ".config", "qwilr"}, "/"))
	viper.SetEnvPrefix("QWILR")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}

package cli

import (
	"fmt"
	"os"

	"github.com/gaskaj/OpenAgentFramework/internal/version"
	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd is the base command for the CLI.
var rootCmd = &cobra.Command{
	Use:     "agentctl",
	Short:   "Autonomous development agent controller",
	Long:    "agentctl manages autonomous development agents that monitor GitHub issues, write code, and create pull requests.",
	Version: version.Version,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (required)")
}

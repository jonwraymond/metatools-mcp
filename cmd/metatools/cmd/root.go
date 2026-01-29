package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// Version information (set at build time)
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// NewRootCmd creates the root command for metatools-mcp.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "metatools",
		Short: "MCP server for progressive tool discovery and execution",
		Long: `metatools is the MCP server that exposes the tool stack via a small,
progressive-disclosure tool surface. It composes toolmodel, toolindex, tooldocs,
toolrun, and optionally toolcode/toolruntime.

Use subcommands to start the server or manage configuration.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

// Execute runs the root command.
func Execute() error {
	return NewRootCmd().Execute()
}

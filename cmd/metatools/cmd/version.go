package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

type versionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

func newVersionCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print the version, git commit, build date, and Go version.",
		RunE: func(cmd *cobra.Command, args []string) error {
			info := versionInfo{
				Version:   Version,
				GitCommit: GitCommit,
				BuildDate: BuildDate,
				GoVersion: runtime.Version(),
				Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			}

			if jsonOutput {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "metatools %s\n", info.Version)
			fmt.Fprintf(cmd.OutOrStdout(), "  Git commit: %s\n", info.GitCommit)
			fmt.Fprintf(cmd.OutOrStdout(), "  Build date: %s\n", info.BuildDate)
			fmt.Fprintf(cmd.OutOrStdout(), "  Go version: %s\n", info.GoVersion)
			fmt.Fprintf(cmd.OutOrStdout(), "  Platform:   %s\n", info.Platform)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output version as JSON")

	return cmd
}

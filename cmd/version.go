package cmd

import (
	"fmt"

	"github.com/mindovermachine-dev/gh-docs/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	var (
		configFile string
		flags      version.Config
	)

	c := &cobra.Command{
		Use:   "version",
		Short: "Write a build provenance file",
		Long: `Write the repository, ref and commit of the current build to a text file.

Each value is taken from the first source that defines it:
  1. BUILD_REPO, BUILD_REF, BUILD_SHA
  2. GitHub Actions: GITHUB_SERVER_URL + GITHUB_REPOSITORY, GITHUB_REF_NAME, GITHUB_SHA
  3. git: remote "origin", current branch, HEAD

Precedence for the output path: flags > config file (.docs.yml) > built-in fallback.
Config keys, under a "version" node: file.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := version.LoadConfig(configFile, cmd.Flags().Changed("config"))
			if err != nil {
				return err
			}
			path, err := version.Run(version.Merge(cfg, flags))
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s\n", path)
			return nil
		},
	}

	f := c.Flags()
	f.StringVarP(&configFile, "config", "c", ".docs.yml", "Path to the config file")
	f.StringVarP(&flags.File, "file", "f", "", "Output file path (default dist/version.txt)")
	return c
}

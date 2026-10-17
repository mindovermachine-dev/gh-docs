package cmd

import (
	"fmt"

	"github.com/mindovermachine-dev/gh-docs/internal/index"
	"github.com/spf13/cobra"
)

func newIndexCmd() *cobra.Command {
	var (
		configFile string
		flags      index.Config
	)

	c := &cobra.Command{
		Use:   "index",
		Short: "Generate an index of documentation repositories",
		Long: `List repositories of a user/organization that carry a given topic and
write their name, description and homepage URL to a YAML, JSON, HTML or Starlight sidebar file.

With --file the extension decides the format; --file and --format are mutually
exclusive. Missing parent directories are created.

Precedence: flags > config file (.docs.yml) > built-in fallbacks.
Config keys, under an "index" node: owner, visibility, topic, format, file, limit.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := index.LoadConfig(configFile, cmd.Flags().Changed("config"))
			if err != nil {
				return err
			}
			path, err := index.Run(index.Merge(cfg, flags))
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s\n", path)
			return nil
		},
	}

	f := c.Flags()
	f.StringVarP(&configFile, "config", "c", ".docs.yml", "Path to the config file")
	f.StringVarP(&flags.Owner, "owner", "o", "", "User or organization to list repositories for")
	f.StringVarP(&flags.Visibility, "visibility", "v", "", "Repository visibility: public, internal or private")
	f.StringVarP(&flags.Topic, "topic", "t", "", "Repository topic to filter on")
	f.StringVarP(&flags.Format, "format", "F", "", "Output format: yaml, json, html or sidebar (not valid with --file)")
	f.StringVarP(&flags.File, "file", "f", "", "Output file path (.yml, .yaml, .json, .html or .msj); extension sets the format")
	f.IntVarP(&flags.Limit, "limit", "l", 0, "Maximum number of repositories to fetch")
	c.MarkFlagsMutuallyExclusive("file", "format")
	return c
}

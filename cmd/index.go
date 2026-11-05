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
write their name, description and homepage URL to a YAML, JSON or HTML file.
The sidebar format writes repository pages from a Markdown template and a paired
Starlight sidebar module.

With --file the extension decides the format; --file and --format are mutually
exclusive. Missing parent directories are created.

Precedence: flags > config file (.docs.yml) > built-in fallbacks.
Config keys, under an "index" node: owner, visibility, topic, format, file, pages-dir, base-url, badge-url, limit.
Sidebar mode requires the user-authored docs/repository-template.md file; see "gh docs template".`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := index.LoadConfig(configFile, cmd.Flags().Changed("config"))
			if err != nil {
				return err
			}
			merged := index.Merge(cfg, flags)
			path, err := index.Run(merged)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s\n", path)
			if merged.Format == "sidebar" {
				fmt.Fprintf(cmd.OutOrStdout(), "Wrote pages in %s\n", merged.PagesDir)
			}
			return nil
		},
	}

	f := c.Flags()
	f.StringVarP(&configFile, "config", "c", ".docs.yml", "Path to the config file")
	f.StringVarP(&flags.Owner, "owner", "o", "", "User or organization to list repositories for")
	f.StringVarP(&flags.Visibility, "visibility", "v", "", "Repository visibility: public, internal or private")
	f.StringVarP(&flags.Topic, "topic", "t", "", "Repository topic to filter on")
	f.StringVarP(&flags.Format, "format", "F", "", "Output format: yaml, json, html or sidebar (not valid with --file)")
	f.StringVarP(&flags.File, "file", "f", "", "Output file path (.yml, .yaml, .json, .html or .mjs); extension sets the format")
	f.StringVar(&flags.PagesDir, "pages-dir", "", "Directory for generated Markdown repository pages (sidebar format)")
	f.StringVar(&flags.BaseURL, "base-url", "", "Base URL for generated local sidebar links")
	f.StringVar(&flags.BadgeURL, "badge-url", "", "Optional status badge URL for repository page templates")
	f.IntVarP(&flags.Limit, "limit", "l", 0, "Maximum number of repositories to fetch")
	c.MarkFlagsMutuallyExclusive("file", "format")
	return c
}

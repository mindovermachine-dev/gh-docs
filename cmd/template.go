package cmd

import (
	"errors"
	"fmt"

	"github.com/mindovermachine-dev/gh-docs/internal/template"
	"github.com/spf13/cobra"
)

func newTemplateCmd() *cobra.Command {
	var (
		configFile string
		flags      template.Config
		jekyll     bool
		force      bool
	)

	c := &cobra.Command{
		Use:   "template",
		Short: "Generate starter files for the documentation site",
		Long: `Generate a minimal GitHub Pages site that renders the list produced by
"index". Files are plain Jekyll sources; GitHub Pages builds them itself.

Precedence: flags > config file (.docs.yml) > built-in fallbacks.
Config keys, under a "template" node: output.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !jekyll {
				return errors.New("a template type is required: use --jekyll")
			}
			cfg, err := template.LoadConfig(configFile, cmd.Flags().Changed("config"))
			if err != nil {
				return err
			}
			paths, err := template.Jekyll(template.Merge(cfg, flags), force)
			if err != nil {
				return err
			}
			for _, p := range paths {
				fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s\n", p)
			}
			return nil
		},
	}

	f := c.Flags()
	f.BoolVar(&jekyll, "jekyll", false, "Generate a Jekyll (GitHub Pages) index.md and _config.yml")
	f.StringVarP(&configFile, "config", "c", ".docs.yml", "Path to the config file")
	f.StringVarP(&flags.Output, "output", "o", "", "Output folder")
	f.BoolVar(&force, "force", false, "Overwrite existing files")
	return c
}

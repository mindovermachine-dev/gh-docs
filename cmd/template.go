package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/mindovermachine-dev/gh-docs/internal/index"
	"github.com/mindovermachine-dev/gh-docs/internal/template"
	"github.com/spf13/cobra"
)

func newTemplateCmd() *cobra.Command {
	var (
		configFile string
		flags      template.Config
		jekyll     bool
		starlight  bool
		force      bool
	)

	c := &cobra.Command{
		Use:   "template",
		Short: "Generate starter files for the documentation site",
		Long: `Generate starter files for the documentation site.

Use --jekyll to generate a GitHub Pages index, or --starlight to copy a Markdown
repository page template used by "index --format sidebar".

Precedence: flags > config file (.docs.yml) > built-in fallbacks.
Config keys, under a "template" node: output.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !jekyll && !starlight {
				return errors.New("a template type is required: use --jekyll or --starlight")
			}
			cfg, err := template.LoadConfig(configFile, cmd.Flags().Changed("config"))
			if err != nil {
				return err
			}
			merged := template.Merge(cfg, flags)
			if starlight {
				destination := filepath.Join(merged.Output, "repository-template.md")
				if err := index.StarlightTemplate(destination, force); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s\n", destination)
				return nil
			}
			paths, err := template.Jekyll(merged, force)
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
	f.BoolVar(&starlight, "starlight", false, "Copy a Starlight repository page template into the output folder")
	f.StringVarP(&configFile, "config", "c", ".docs.yml", "Path to the config file")
	f.StringVarP(&flags.Output, "output", "o", "", "Output folder")
	f.BoolVar(&force, "force", false, "Overwrite existing files")
	c.MarkFlagsMutuallyExclusive("jekyll", "starlight")
	return c
}

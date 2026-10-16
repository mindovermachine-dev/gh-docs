package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "docs",
		Short: "GitHub utilities extension",
		Long: `A collection of utilities for helping and automating the documentations pages.`,
	}
	root.CompletionOptions.HiddenDefaultCmd = true
	root.AddCommand(newIndexCmd(), newTemplateCmd())
	return root
}

// Execute runs the root command and exits non-zero on error.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

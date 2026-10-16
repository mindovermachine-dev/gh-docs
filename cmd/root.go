package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "godo",
		Short: "GitHub utilities extension",
		Long: `A collection of utilities for GitHub workflows and automation.
This is a GitHub CLI extension that provides additional commands
to enhance your GitHub workflow.`,
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

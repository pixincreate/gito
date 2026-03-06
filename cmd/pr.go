package cmd

import (
	"github.com/spf13/cobra"
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Manage pull requests",
	Long:  "Commands for creating and managing GitHub pull requests",
}

func init() {
	rootCmd.AddCommand(prCmd)
}

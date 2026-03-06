package cmd

import (
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull operations from GitHub",
	Long:  "Commands for pulling data from GitHub repositories",
}

func init() {
	rootCmd.AddCommand(pullCmd)
}

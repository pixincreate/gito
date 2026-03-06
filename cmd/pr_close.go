package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type prCloseFlags struct {
	repo string
}

var closeFlags prCloseFlags

var prCloseCmd = &cobra.Command{
	Use:   "close <pr_number>",
	Short: "Close a pull request without merging",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := closeFlags.repo
		if repo == "" {
			r, err := detectRepo()
			if err != nil {
				return fmt.Errorf("failed to detect repository: %w", err)
			}
			repo = r
		}

		client, err := resolveClient()
		if err != nil {
			return err
		}

		pr, err := client.ClosePullRequest(repo, args[0])
		if err != nil {
			return fmt.Errorf("failed to close PR: %w", err)
		}

		fmt.Printf("Closed PR #%d: %s\n", pr.Number, pr.Title)
		fmt.Printf("URL: %s\n", pr.HTMLURL)

		return nil
	},
}

func init() {
	prCmd.AddCommand(prCloseCmd)
	prCloseCmd.Flags().StringVar(&closeFlags.repo, "repo", "", "repository in owner/repo format")
}

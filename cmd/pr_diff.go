package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type prDiffFlags struct {
	repo string
}

var diffFlags prDiffFlags

var prDiffCmd = &cobra.Command{
	Use:   "diff <pr_number>",
	Short: "Get pull request diff",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := diffFlags.repo
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

		diff, err := client.GetPullRequestDiff(repo, args[0])
		if err != nil {
			return fmt.Errorf("failed to get PR diff: %w", err)
		}

		fmt.Print(diff)
		return nil
	},
}

func init() {
	prCmd.AddCommand(prDiffCmd)
	prDiffCmd.Flags().StringVar(&diffFlags.repo, "repo", "", "repository in owner/repo format")
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type prReviewersFlags struct {
	repo string
}

var reviewersFlags prReviewersFlags

var prReviewersCmd = &cobra.Command{
	Use:   "reviewers <pr_number>",
	Short: "List requested reviewers for a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := reviewersFlags.repo
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

		reviewers, err := client.GetPullRequestReviewers(repo, args[0])
		if err != nil {
			return fmt.Errorf("failed to get reviewers: %w", err)
		}

		if len(reviewers) == 0 {
			fmt.Println("No requested reviewers")
			return nil
		}

		for _, reviewer := range reviewers {
			fmt.Println(reviewer)
		}

		return nil
	},
}

func init() {
	prCmd.AddCommand(prReviewersCmd)
	prReviewersCmd.Flags().StringVar(&reviewersFlags.repo, "repo", "", "repository in owner/repo format")
}

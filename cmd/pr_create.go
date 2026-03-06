package cmd

import (
	"fmt"

	"github.com/pixincreate/gito/internal/github"
	"github.com/spf13/cobra"
)

type createPRFlags struct {
	title     string
	body      string
	base      string
	head      string
	labels    []string
	assignees []string
	reviewers []string
	repo      string
}

var prCreateFlags createPRFlags

var prCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prCreateFlags.title = args[0]
		repo := prCreateFlags.repo
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

		pr, err := client.CreatePullRequest(repo, github.PullRequest{
			Title:     prCreateFlags.title,
			Body:      prCreateFlags.body,
			Base:      github.Branch{Ref: prCreateFlags.base},
			Head:      github.Branch{Ref: prCreateFlags.head},
			Labels:    prCreateFlags.labels,
			Assignees: prCreateFlags.assignees,
			Reviewers: prCreateFlags.reviewers,
		})
		if err != nil {
			return fmt.Errorf("failed to create PR: %w", err)
		}

		fmt.Printf("PR created successfully: %s\n", pr.HTMLURL)
		return nil
	},
}

func init() {
	prCmd.AddCommand(prCreateCmd)
	prCreateCmd.Flags().StringVar(&prCreateFlags.body, "body", "", "PR body/description")
	prCreateCmd.Flags().StringVar(&prCreateFlags.base, "base", "main", "base branch")
	prCreateCmd.Flags().StringVar(&prCreateFlags.head, "head", "", "head branch")
	prCreateCmd.Flags().StringSliceVar(&prCreateFlags.labels, "label", nil, "labels (can be repeated)")
	prCreateCmd.Flags().StringSliceVar(&prCreateFlags.assignees, "assignee", nil, "assignees (can be repeated)")
	prCreateCmd.Flags().StringSliceVar(&prCreateFlags.reviewers, "reviewer", nil, "reviewers (can be repeated)")
	prCreateCmd.Flags().StringVar(&prCreateFlags.repo, "repo", "", "repository in owner/repo format")
}

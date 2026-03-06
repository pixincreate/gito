package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type prMergeFlags struct {
	repo         string
	method       string
	deleteBranch bool
}

var mergeFlags prMergeFlags

var prMergeCmd = &cobra.Command{
	Use:   "merge <pr_number>",
	Short: "Merge a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if mergeFlags.method != "merge" && mergeFlags.method != "squash" && mergeFlags.method != "rebase" {
			return fmt.Errorf("invalid --method value %q, expected merge|squash|rebase", mergeFlags.method)
		}

		repo := mergeFlags.repo
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

		pr, err := client.MergePullRequest(repo, args[0], mergeFlags.method, mergeFlags.deleteBranch)
		if err != nil {
			return fmt.Errorf("failed to merge PR: %w", err)
		}

		fmt.Printf("Merged PR #%d: %s\n", pr.Number, pr.Title)
		fmt.Printf("URL: %s\n", pr.HTMLURL)
		if mergeFlags.deleteBranch {
			fmt.Println("Head branch deleted")
		}

		return nil
	},
}

func init() {
	prCmd.AddCommand(prMergeCmd)
	prMergeCmd.Flags().StringVar(&mergeFlags.repo, "repo", "", "repository in owner/repo format")
	prMergeCmd.Flags().StringVar(&mergeFlags.method, "method", "merge", "merge method: merge, squash, or rebase")
	prMergeCmd.Flags().BoolVar(&mergeFlags.deleteBranch, "delete-branch", false, "delete source branch after merge")
}

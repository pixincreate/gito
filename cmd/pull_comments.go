package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type pullCommentsFlags struct {
	output string
	repo   string
}

var pullComments pullCommentsFlags

var pullCommentsCmd = &cobra.Command{
	Use:   "comments <pr_number>",
	Short: "Fetch review comments from a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := pullComments.repo
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

		comments, err := client.GetPullRequestComments(repo, args[0])
		if err != nil {
			return fmt.Errorf("failed to get comments: %w", err)
		}

		var output *os.File
		if pullComments.output != "" {
			f, err := os.Create(pullComments.output)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer f.Close()
			output = f
		} else {
			output = os.Stdout
		}

		for _, comment := range comments {
			fmt.Fprintln(output, "Author:", comment.Author)
			fmt.Fprintln(output, "PR Number:", args[0])
			fmt.Fprintln(output, "Diff:", comment.DiffHunk)
			fmt.Fprintln(output, "Review comment:", comment.Body)
			fmt.Fprintln(output, "URL:", comment.HTMLURL)
			fmt.Fprintln(output, "Created At:", comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
			fmt.Fprintln(output, "Author Association:", comment.AuthorAssociation)
			fmt.Fprintln(output, strings.Repeat("-", 80))
		}

		return nil
	},
}

func init() {
	pullCmd.AddCommand(pullCommentsCmd)
	pullCommentsCmd.Flags().StringVar(&pullComments.output, "output", "", "output file path (optional)")
	pullCommentsCmd.Flags().StringVar(&pullComments.repo, "repo", "", "repository in owner/repo format")
}

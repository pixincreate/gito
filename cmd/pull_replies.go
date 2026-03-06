package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type pullRepliesFlags struct {
	repo string
	body string
}

var repliesFlags pullRepliesFlags

var pullRepliesCmd = &cobra.Command{
	Use:   "replies <pr_number> <comment_id>",
	Short: "Reply to a pull request review comment",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if repliesFlags.body == "" {
			return fmt.Errorf("--body is required")
		}

		repo := repliesFlags.repo
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

		reply, err := client.ReplyToPullRequestComment(repo, args[0], args[1], repliesFlags.body)
		if err != nil {
			return fmt.Errorf("failed to reply to comment: %w", err)
		}

		fmt.Printf("Reply created: %s\n", reply.HTMLURL)
		return nil
	},
}

func init() {
	pullCmd.AddCommand(pullRepliesCmd)
	pullRepliesCmd.Flags().StringVar(&repliesFlags.repo, "repo", "", "repository in owner/repo format")
	pullRepliesCmd.Flags().StringVar(&repliesFlags.body, "body", "", "reply body")
	_ = pullRepliesCmd.MarkFlagRequired("body")
}

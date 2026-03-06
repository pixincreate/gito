package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type prViewFlags struct {
	repo string
}

var viewFlags prViewFlags

var prViewCmd = &cobra.Command{
	Use:   "view <pr_number>",
	Short: "View pull request details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := viewFlags.repo
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

		pr, err := client.GetPullRequest(repo, args[0])
		if err != nil {
			return fmt.Errorf("failed to fetch PR: %w", err)
		}

		fmt.Printf("Number: %d\n", pr.Number)
		fmt.Printf("Title: %s\n", pr.Title)
		fmt.Printf("Author: %s\n", pr.Author)
		fmt.Printf("State: %s\n", pr.State)
		fmt.Printf("Branch: %s ← %s\n", pr.Base.Ref, pr.Head.Ref)
		fmt.Printf("Body: %s\n", pr.Body)
		fmt.Printf("URL: %s\n", pr.HTMLURL)
		fmt.Printf("Created: %s\n", pr.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
		fmt.Printf("Labels: %s\n", strings.Join(pr.Labels, ", "))
		fmt.Printf("Assignees: %s\n", strings.Join(pr.Assignees, ", "))

		return nil
	},
}

func init() {
	prCmd.AddCommand(prViewCmd)
	prViewCmd.Flags().StringVar(&viewFlags.repo, "repo", "", "repository in owner/repo format")
}

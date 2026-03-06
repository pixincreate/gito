package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type prListFlags struct {
	repo  string
	state string
	limit int
}

var listFlags prListFlags

var prListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pull requests",
	RunE: func(cmd *cobra.Command, args []string) error {
		if listFlags.state != "open" && listFlags.state != "closed" && listFlags.state != "all" {
			return fmt.Errorf("invalid --state value %q, expected open|closed|all", listFlags.state)
		}
		if listFlags.limit <= 0 {
			return fmt.Errorf("--limit must be greater than 0")
		}

		repo := listFlags.repo
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

		prs, err := client.ListPullRequests(repo, listFlags.state, listFlags.limit)
		if err != nil {
			return fmt.Errorf("failed to list PRs: %w", err)
		}

		fmt.Printf("%-8s %-60s %-20s %-20s\n", "#NUMBER", "TITLE", "AUTHOR", "CREATED")
		for _, pr := range prs {
			title := pr.Title
			if len([]rune(title)) > 60 {
				title = string([]rune(title)[:57]) + "..."
			}
			title = strings.ReplaceAll(title, "\n", " ")
			fmt.Printf("%-8d %-60s %-20s %-20s\n", pr.Number, title, pr.Author, pr.CreatedAt.Format("2006-01-02 15:04"))
		}

		return nil
	},
}

func init() {
	prCmd.AddCommand(prListCmd)
	prListCmd.Flags().StringVar(&listFlags.repo, "repo", "", "repository in owner/repo format")
	prListCmd.Flags().StringVar(&listFlags.state, "state", "open", "PR state: open, closed, or all")
	prListCmd.Flags().IntVar(&listFlags.limit, "limit", 30, "maximum number of PRs to list")
}

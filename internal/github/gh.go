package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type GHClient struct{}

type ghPullRequest struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	HTMLURL   string    `json:"html_url"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	User      ghUser    `json:"user"`
	Base      struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"base"`
	Head struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"head"`
	Labels    []ghLabel `json:"labels"`
	Assignees []ghUser  `json:"assignees"`
}

type ghLabel struct {
	Name string `json:"name"`
}

type ghUser struct {
	Login string `json:"login"`
}

type ghTeam struct {
	Name string `json:"name"`
}

type ghComment struct {
	Body              string    `json:"body"`
	DiffHunk          string    `json:"diff_hunk"`
	Path              string    `json:"path"`
	Line              int       `json:"line"`
	HTMLURL           string    `json:"html_url"`
	CreatedAt         time.Time `json:"created_at"`
	AuthorAssociation string    `json:"author_association"`
	User              ghUser    `json:"user"`
}

type ghRequestedReviewers struct {
	Users []ghUser `json:"users"`
	Teams []ghTeam `json:"teams"`
}

func (c *GHClient) ghCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("gh", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gh command failed: %w\n%s", err, string(out))
	}
	return out, nil
}

func (c *GHClient) apiRequest(method, endpoint string, fields map[string]string, extraHeaders ...string) ([]byte, error) {
	args := []string{"api", "-X", method}
	args = append(args,
		"-H", "Accept: application/vnd.github+json",
		"-H", "X-GitHub-Api-Version: 2022-11-28",
	)
	for _, header := range extraHeaders {
		args = append(args, "-H", header)
	}
	for key, value := range fields {
		args = append(args, "-f", fmt.Sprintf("%s=%s", key, value))
	}
	args = append(args, endpoint)

	return c.ghCommand(args...)
}

func mapGHPullRequest(pr ghPullRequest) PullRequest {
	labels := make([]string, 0, len(pr.Labels))
	for _, label := range pr.Labels {
		labels = append(labels, label.Name)
	}

	assignees := make([]string, 0, len(pr.Assignees))
	for _, assignee := range pr.Assignees {
		assignees = append(assignees, assignee.Login)
	}

	return PullRequest{
		Number:    pr.Number,
		Title:     pr.Title,
		Body:      pr.Body,
		HTMLURL:   pr.HTMLURL,
		State:     pr.State,
		Author:    pr.User.Login,
		CreatedAt: pr.CreatedAt,
		Base:      Branch{Ref: pr.Base.Ref, SHA: pr.Base.SHA},
		Head:      Branch{Ref: pr.Head.Ref, SHA: pr.Head.SHA},
		Labels:    labels,
		Assignees: assignees,
	}
}

func (c *GHClient) CreatePullRequest(repo string, pr PullRequest) (*PullRequest, error) {
	args := []string{"pr", "create", "--repo", repo, "--title", pr.Title}
	if pr.Body != "" {
		args = append(args, "--body", pr.Body)
	}
	if pr.Base != (Branch{}) && pr.Base.Ref != "" {
		args = append(args, "--base", pr.Base.Ref)
	}
	if pr.Head != (Branch{}) && pr.Head.Ref != "" {
		args = append(args, "--head", pr.Head.Ref)
	}
	for _, label := range pr.Labels {
		args = append(args, "--label", label)
	}
	for _, assignee := range pr.Assignees {
		args = append(args, "--assignee", assignee)
	}
	for _, reviewer := range pr.Reviewers {
		args = append(args, "--reviewer", reviewer)
	}

	out, err := c.ghCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("gh pr create failed: %w", err)
	}

	output := strings.TrimSpace(string(out))
	result := &PullRequest{
		HTMLURL: output,
	}
	if strings.HasPrefix(output, "http") {
		return result, nil
	}

	return result, nil
}

func (c *GHClient) ListPullRequests(repo, state string, limit int) ([]PullRequest, error) {
	endpoint := fmt.Sprintf("/repos/%s/pulls?state=%s&per_page=%d", repo, state, limit)
	out, err := c.apiRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var prs []ghPullRequest
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := make([]PullRequest, 0, len(prs))
	for _, pr := range prs {
		result = append(result, mapGHPullRequest(pr))
	}

	return result, nil
}

func (c *GHClient) GetPullRequest(repo, prNumber string) (*PullRequest, error) {
	out, err := c.apiRequest("GET", fmt.Sprintf("/repos/%s/pulls/%s", repo, prNumber), nil)
	if err != nil {
		return nil, err
	}

	var pr ghPullRequest
	if err := json.Unmarshal(out, &pr); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := mapGHPullRequest(pr)
	return &result, nil
}

func (c *GHClient) MergePullRequest(repo, prNumber, method string, deleteBranch bool) (*PullRequest, error) {
	pr, err := c.GetPullRequest(repo, prNumber)
	if err != nil {
		return nil, err
	}

	_, err = c.apiRequest("PUT", fmt.Sprintf("/repos/%s/pulls/%s/merge", repo, prNumber), map[string]string{
		"merge_method": method,
	})
	if err != nil {
		return nil, err
	}

	if deleteBranch && pr.Head.Ref != "" {
		_, err = c.apiRequest("DELETE", fmt.Sprintf("/repos/%s/git/refs/heads/%s", repo, pr.Head.Ref), nil)
		if err != nil {
			return nil, err
		}
	}

	pr.State = "closed"
	return pr, nil
}

func (c *GHClient) ClosePullRequest(repo, prNumber string) (*PullRequest, error) {
	out, err := c.apiRequest("PATCH", fmt.Sprintf("/repos/%s/pulls/%s", repo, prNumber), map[string]string{
		"state": "closed",
	})
	if err != nil {
		return nil, err
	}

	var pr ghPullRequest
	if err := json.Unmarshal(out, &pr); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := mapGHPullRequest(pr)
	return &result, nil
}

func (c *GHClient) GetPullRequestComments(repo, prNumber string) ([]ReviewComment, error) {
	out, err := c.apiRequest("GET", fmt.Sprintf("/repos/%s/pulls/%s/comments", repo, prNumber), nil)
	if err != nil {
		return nil, err
	}

	var comments []ghComment
	if err := json.Unmarshal(out, &comments); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var result []ReviewComment
	for _, c := range comments {
		comment := ReviewComment{
			Author:            c.User.Login,
			AuthorAssociation: c.AuthorAssociation,
			Body:              c.Body,
			DiffHunk:          c.DiffHunk,
			Path:              c.Path,
			Line:              c.Line,
			HTMLURL:           c.HTMLURL,
			CreatedAt:         c.CreatedAt,
		}
		result = append(result, comment)
	}

	return result, nil
}

func (c *GHClient) ReplyToPullRequestComment(repo, prNumber, commentID, body string) (*ReviewComment, error) {
	out, err := c.apiRequest("POST", fmt.Sprintf("/repos/%s/pulls/%s/comments/%s/replies", repo, prNumber, commentID), map[string]string{
		"body": body,
	})
	if err != nil {
		return nil, err
	}

	var comment ghComment
	if err := json.Unmarshal(out, &comment); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := &ReviewComment{
		Author:            comment.User.Login,
		AuthorAssociation: comment.AuthorAssociation,
		Body:              comment.Body,
		DiffHunk:          comment.DiffHunk,
		Path:              comment.Path,
		Line:              comment.Line,
		HTMLURL:           comment.HTMLURL,
		CreatedAt:         comment.CreatedAt,
	}

	return result, nil
}

func (c *GHClient) GetPullRequestDiff(repo, prNumber string) (string, error) {
	out, err := c.apiRequest("GET", fmt.Sprintf("/repos/%s/pulls/%s", repo, prNumber), nil, "Accept: application/vnd.github.diff")
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (c *GHClient) GetPullRequestReviewers(repo, prNumber string) ([]string, error) {
	out, err := c.apiRequest("GET", fmt.Sprintf("/repos/%s/pulls/%s/requested_reviewers", repo, prNumber), nil)
	if err != nil {
		return nil, err
	}

	var reviewers ghRequestedReviewers
	if err := json.Unmarshal(out, &reviewers); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := make([]string, 0, len(reviewers.Users)+len(reviewers.Teams))
	for _, user := range reviewers.Users {
		result = append(result, user.Login)
	}
	for _, team := range reviewers.Teams {
		result = append(result, team.Name)
	}

	return result, nil
}

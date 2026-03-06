package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

const (
	apiVersion = "2022-11-28"
	apiURL     = "https://api.github.com"
)

type CurlClient struct {
	Token string
}

type apiPullRequest struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	HTMLURL   string    `json:"html_url"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	Base      struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"base"`
	Head struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"head"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
	Assignees []struct {
		Login string `json:"login"`
	} `json:"assignees"`
	User struct {
		Login string `json:"login"`
	} `json:"user"`
}

type apiPullRequestCreate struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Base  string `json:"base"`
	Head  string `json:"head"`
}

type apiComment struct {
	Body              string    `json:"body"`
	DiffHunk          string    `json:"diff_hunk"`
	Path              string    `json:"path"`
	Line              int       `json:"line"`
	HTMLURL           string    `json:"html_url"`
	CreatedAt         time.Time `json:"created_at"`
	AuthorAssociation string    `json:"author_association"`
	User              struct {
		Login string `json:"login"`
	} `json:"user"`
}

type apiMergeRequest struct {
	MergeMethod string `json:"merge_method"`
}

type apiReplyRequest struct {
	Body string `json:"body"`
}

type apiReviewersResponse struct {
	Users []struct {
		Login string `json:"login"`
	} `json:"users"`
	Teams []struct {
		Name string `json:"name"`
	} `json:"teams"`
}

func splitRepo(repo string) (string, string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo format: %s (expected owner/repo)", repo)
	}

	return parts[0], parts[1], nil
}

func mapAPIPullRequest(pr apiPullRequest) PullRequest {
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

func (c *CurlClient) doRequest(method, endpoint string, body []byte) ([]byte, error) {
	return c.doRequestWithAccept(method, endpoint, "application/vnd.github+json", body)
}

func (c *CurlClient) doRequestWithAccept(method, endpoint, accept string, body []byte) ([]byte, error) {
	args := []string{
		"-L",
		"-s",
		"-w", "\n%{http_code}",
		"-X", method,
		"-H", fmt.Sprintf("Accept: %s", accept),
		"-H", fmt.Sprintf("Authorization: Bearer %s", c.Token),
		"-H", fmt.Sprintf("X-GitHub-Api-Version: %s", apiVersion),
	}

	if body != nil {
		args = append(args, "-d", string(body))
	}

	args = append(args, apiURL+endpoint)

	cmd := exec.Command("curl", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("curl failed: %w\n%s", err, string(out))
	}

	output := string(out)
	lines := strings.Split(output, "\n")
	statusCode := strings.TrimSpace(lines[len(lines)-1])
	responseBody := strings.TrimSpace(strings.Join(lines[:len(lines)-1], "\n"))

	if statusCode >= "400" {
		return nil, fmt.Errorf("API request failed (HTTP %s): %s", statusCode, responseBody)
	}

	return []byte(responseBody), nil
}

func (c *CurlClient) CreatePullRequest(repo string, pr PullRequest) (*PullRequest, error) {
	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}

	apiPr := apiPullRequestCreate{
		Title: pr.Title,
		Body:  pr.Body,
	}
	if pr.Base.Ref != "" {
		apiPr.Base = pr.Base.Ref
	}
	if pr.Head.Ref != "" {
		apiPr.Head = pr.Head.Ref
	}

	body, err := json.Marshal(apiPr)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	out, err := c.doRequest("POST", fmt.Sprintf("/repos/%s/%s/pulls", owner, repoName), body)
	if err != nil {
		return nil, err
	}

	var result apiPullRequest
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	mapped := mapAPIPullRequest(result)
	return &mapped, nil
}

func (c *CurlClient) ListPullRequests(repo, state string, limit int) ([]PullRequest, error) {
	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}

	out, err := c.doRequest("GET", fmt.Sprintf("/repos/%s/%s/pulls?state=%s&per_page=%d", owner, repoName, state, limit), nil)
	if err != nil {
		return nil, err
	}

	var prs []apiPullRequest
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := make([]PullRequest, 0, len(prs))
	for _, pr := range prs {
		result = append(result, mapAPIPullRequest(pr))
	}

	return result, nil
}

func (c *CurlClient) GetPullRequest(repo, prNumber string) (*PullRequest, error) {
	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}

	out, err := c.doRequest("GET", fmt.Sprintf("/repos/%s/%s/pulls/%s", owner, repoName, prNumber), nil)
	if err != nil {
		return nil, err
	}

	var pr apiPullRequest
	if err := json.Unmarshal(out, &pr); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	mapped := mapAPIPullRequest(pr)
	return &mapped, nil
}

func (c *CurlClient) MergePullRequest(repo, prNumber, method string, deleteBranch bool) (*PullRequest, error) {
	pr, err := c.GetPullRequest(repo, prNumber)
	if err != nil {
		return nil, err
	}

	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(apiMergeRequest{MergeMethod: method})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	_, err = c.doRequest("PUT", fmt.Sprintf("/repos/%s/%s/pulls/%s/merge", owner, repoName, prNumber), body)
	if err != nil {
		return nil, err
	}

	if deleteBranch && pr.Head.Ref != "" {
		_, err = c.doRequest("DELETE", fmt.Sprintf("/repos/%s/%s/git/refs/heads/%s", owner, repoName, pr.Head.Ref), nil)
		if err != nil {
			return nil, err
		}
	}

	pr.State = "closed"
	return pr, nil
}

func (c *CurlClient) ClosePullRequest(repo, prNumber string) (*PullRequest, error) {
	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(map[string]string{"state": "closed"})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	out, err := c.doRequest("PATCH", fmt.Sprintf("/repos/%s/%s/pulls/%s", owner, repoName, prNumber), body)
	if err != nil {
		return nil, err
	}

	var pr apiPullRequest
	if err := json.Unmarshal(out, &pr); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	mapped := mapAPIPullRequest(pr)
	return &mapped, nil
}

func (c *CurlClient) GetPullRequestComments(repo, prNumber string) ([]ReviewComment, error) {
	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}

	out, err := c.doRequest("GET", fmt.Sprintf("/repos/%s/%s/pulls/%s/comments", owner, repoName, prNumber), nil)
	if err != nil {
		return nil, err
	}

	var comments []apiComment
	if err := json.Unmarshal(out, &comments); err != nil {
		reader := bytes.NewReader(out)
		data, _ := io.ReadAll(reader)
		return nil, fmt.Errorf("failed to parse response: %w, response: %s", err, string(data))
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

func (c *CurlClient) ReplyToPullRequestComment(repo, prNumber, commentID, bodyText string) (*ReviewComment, error) {
	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(apiReplyRequest{Body: bodyText})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	out, err := c.doRequest("POST", fmt.Sprintf("/repos/%s/%s/pulls/%s/comments/%s/replies", owner, repoName, prNumber, commentID), body)
	if err != nil {
		return nil, err
	}

	var comment apiComment
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

func (c *CurlClient) GetPullRequestDiff(repo, prNumber string) (string, error) {
	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return "", err
	}

	out, err := c.doRequestWithAccept("GET", fmt.Sprintf("/repos/%s/%s/pulls/%s", owner, repoName, prNumber), "application/vnd.github.diff", nil)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (c *CurlClient) GetPullRequestReviewers(repo, prNumber string) ([]string, error) {
	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}

	out, err := c.doRequest("GET", fmt.Sprintf("/repos/%s/%s/pulls/%s/requested_reviewers", owner, repoName, prNumber), nil)
	if err != nil {
		return nil, err
	}

	var reviewers apiReviewersResponse
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

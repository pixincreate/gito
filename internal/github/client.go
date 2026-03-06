package github

import "time"

type PullRequest struct {
	Number    int
	Title     string
	Body      string
	HTMLURL   string
	State     string
	Author    string
	CreatedAt time.Time
	Base      Branch
	Head      Branch
	Labels    []string
	Assignees []string
	Reviewers []string
}

type Branch struct {
	Ref string
	SHA string
}

type ReviewComment struct {
	Author            string
	AuthorAssociation string
	Body              string
	DiffHunk          string
	Path              string
	Line              int
	HTMLURL           string
	CreatedAt         time.Time
}

type Client interface {
	CreatePullRequest(repo string, pr PullRequest) (*PullRequest, error)
	ListPullRequests(repo, state string, limit int) ([]PullRequest, error)
	GetPullRequest(repo, prNumber string) (*PullRequest, error)
	MergePullRequest(repo, prNumber, method string, deleteBranch bool) (*PullRequest, error)
	ClosePullRequest(repo, prNumber string) (*PullRequest, error)
	GetPullRequestComments(repo, prNumber string) ([]ReviewComment, error)
	ReplyToPullRequestComment(repo, prNumber, commentID, body string) (*ReviewComment, error)
	GetPullRequestDiff(repo, prNumber string) (string, error)
	GetPullRequestReviewers(repo, prNumber string) ([]string, error)
}

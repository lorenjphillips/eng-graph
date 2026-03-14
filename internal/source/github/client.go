package github

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v68/github"
)

type PRReview struct {
	ID          int64
	PRID        int
	PRTitle     string
	PRNumber    int
	PRURL       string
	Owner       string
	Repo        string
	Body        string
	State       string
	SubmittedAt time.Time
	User        string
}

type PRReviewComment struct {
	ID        int64
	PRID      int
	PRTitle   string
	PRNumber  int
	PRURL     string
	Owner     string
	Repo      string
	Body      string
	Path      string
	DiffHunk  string
	CreatedAt time.Time
	User      string
}

type PR struct {
	ID        int64
	Number    int
	Title     string
	Body      string
	URL       string
	Owner     string
	Repo      string
	CreatedAt time.Time
	User      string
}

type Client struct {
	gh *github.Client
}

func NewClient(token string) *Client {
	return &Client{
		gh: github.NewClient(nil).WithAuthToken(token),
	}
}

func (c *Client) GetAuthenticatedUser(ctx context.Context) (string, error) {
	user, _, err := c.gh.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	return user.GetLogin(), nil
}

func (c *Client) ListPRReviews(ctx context.Context, owner, repo, user string, since time.Time) ([]PRReview, error) {
	prs, err := c.listAllPRs(ctx, owner, repo, since)
	if err != nil {
		return nil, err
	}

	var reviews []PRReview
	for _, pr := range prs {
		opts := &github.ListOptions{PerPage: 100}
		for {
			page, resp, err := c.gh.PullRequests.ListReviews(ctx, owner, repo, pr.GetNumber(), opts)
			if err != nil {
				return nil, err
			}
			for _, r := range page {
				if r.GetUser().GetLogin() != user {
					continue
				}
				submitted := r.GetSubmittedAt().Time
				if submitted.Before(since) {
					continue
				}
				reviews = append(reviews, PRReview{
					ID:          r.GetID(),
					PRID:        pr.GetNumber(),
					PRTitle:     pr.GetTitle(),
					PRNumber:    pr.GetNumber(),
					PRURL:       pr.GetHTMLURL(),
					Owner:       owner,
					Repo:        repo,
					Body:        r.GetBody(),
					State:       r.GetState(),
					SubmittedAt: submitted,
					User:        user,
				})
			}
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
	}
	return reviews, nil
}

func (c *Client) ListPRReviewComments(ctx context.Context, owner, repo, user string, since time.Time) ([]PRReviewComment, error) {
	prMap, err := c.prTitleMap(ctx, owner, repo, since)
	if err != nil {
		return nil, err
	}

	opts := &github.PullRequestListCommentsOptions{
		Sort:      "created",
		Direction: "desc",
		Since:     since,
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var comments []PRReviewComment
	for {
		page, resp, err := c.gh.PullRequests.ListComments(ctx, owner, repo, 0, opts)
		if err != nil {
			return nil, err
		}
		for _, cm := range page {
			if cm.GetUser().GetLogin() != user {
				continue
			}
			if cm.GetCreatedAt().Time.Before(since) {
				continue
			}
			prNumber := prNumberFromURL(cm.GetPullRequestURL())
			title := prMap[prNumber]
			prURL := "https://github.com/" + owner + "/" + repo + "/pull/" + strconv.Itoa(prNumber)
			comments = append(comments, PRReviewComment{
				ID:        cm.GetID(),
				PRID:      prNumber,
				PRTitle:   title,
				PRNumber:  prNumber,
				PRURL:     prURL,
				Owner:     owner,
				Repo:      repo,
				Body:      cm.GetBody(),
				Path:      cm.GetPath(),
				DiffHunk:  cm.GetDiffHunk(),
				CreatedAt: cm.GetCreatedAt().Time,
				User:      user,
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return comments, nil
}

func (c *Client) ListPRs(ctx context.Context, owner, repo, user string, since time.Time) ([]PR, error) {
	prs, err := c.listAllPRs(ctx, owner, repo, since)
	if err != nil {
		return nil, err
	}

	var out []PR
	for _, pr := range prs {
		if pr.GetUser().GetLogin() != user {
			continue
		}
		out = append(out, PR{
			ID:        int64(pr.GetNumber()),
			Number:    pr.GetNumber(),
			Title:     pr.GetTitle(),
			Body:      pr.GetBody(),
			URL:       pr.GetHTMLURL(),
			Owner:     owner,
			Repo:      repo,
			CreatedAt: pr.GetCreatedAt().Time,
			User:      user,
		})
	}
	return out, nil
}

func (c *Client) listAllPRs(ctx context.Context, owner, repo string, since time.Time) ([]*github.PullRequest, error) {
	opts := &github.PullRequestListOptions{
		State:       "all",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var all []*github.PullRequest
	for {
		prs, resp, err := c.gh.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		for _, pr := range prs {
			if pr.GetUpdatedAt().Time.Before(since) {
				return all, nil
			}
			all = append(all, pr)
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return all, nil
}

func (c *Client) prTitleMap(ctx context.Context, owner, repo string, since time.Time) (map[int]string, error) {
	prs, err := c.listAllPRs(ctx, owner, repo, since)
	if err != nil {
		return nil, err
	}
	m := make(map[int]string, len(prs))
	for _, pr := range prs {
		m[pr.GetNumber()] = pr.GetTitle()
	}
	return m, nil
}

func prNumberFromURL(url string) int {
	// URL format: https://api.github.com/repos/{owner}/{repo}/pulls/{number}
	if i := strings.LastIndex(url, "/"); i >= 0 {
		n, _ := strconv.Atoi(url[i+1:])
		return n
	}
	return 0
}

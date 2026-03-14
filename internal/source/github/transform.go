package github

import (
	"fmt"
	"strconv"

	"github.com/eng-graph/eng-graph/internal/source"
)

func TransformPRReview(r PRReview, sourceName string) source.DataPoint {
	return source.DataPoint{
		ID:        fmt.Sprintf("github:%s/%s:review:%d", r.Owner, r.Repo, r.ID),
		Source:    sourceName,
		Kind:      source.KindPRReview,
		Author:    r.User,
		Body:      r.Body,
		Timestamp: r.SubmittedAt,
		Context: map[string]string{
			"pr_title":     r.PRTitle,
			"pr_number":    strconv.Itoa(r.PRNumber),
			"pr_url":       r.PRURL,
			"review_state": r.State,
			"owner":        r.Owner,
			"repo":         r.Repo,
		},
	}
}

func TransformPRReviewComment(c PRReviewComment, sourceName string) source.DataPoint {
	return source.DataPoint{
		ID:        fmt.Sprintf("github:%s/%s:review_comment:%d", c.Owner, c.Repo, c.ID),
		Source:    sourceName,
		Kind:      source.KindPRReviewComment,
		Author:    c.User,
		Body:      c.Body,
		Timestamp: c.CreatedAt,
		Context: map[string]string{
			"pr_title":  c.PRTitle,
			"pr_number": strconv.Itoa(c.PRNumber),
			"pr_url":    c.PRURL,
			"file_path": c.Path,
			"diff_hunk": c.DiffHunk,
			"owner":     c.Owner,
			"repo":      c.Repo,
		},
	}
}

func TransformPR(pr PR, sourceName string) source.DataPoint {
	return source.DataPoint{
		ID:        fmt.Sprintf("github:%s/%s:pr:%d", pr.Owner, pr.Repo, pr.Number),
		Source:    sourceName,
		Kind:      source.KindPRDescription,
		Author:    pr.User,
		Body:      pr.Body,
		Timestamp: pr.CreatedAt,
		Context: map[string]string{
			"pr_title":  pr.Title,
			"pr_number": strconv.Itoa(pr.Number),
			"pr_url":    pr.URL,
			"owner":     pr.Owner,
			"repo":      pr.Repo,
		},
	}
}

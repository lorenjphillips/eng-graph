package review

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v68/github"
)

type PRDiff struct {
	Title       string
	Description string
	Files       []DiffFile
	URL         string
}

type DiffFile struct {
	Path   string
	Patch  string
	Status string // added, modified, removed
}

var prURLPattern = regexp.MustCompile(`https?://github\.com/([^/]+)/([^/]+)/pull/(\d+)`)

func ParsePRRef(ref string) (owner, repo string, number int, err error) {
	if m := prURLPattern.FindStringSubmatch(ref); m != nil {
		n, _ := strconv.Atoi(m[3])
		return m[1], m[2], n, nil
	}

	if idx := strings.Index(ref, "#"); idx > 0 {
		parts := strings.SplitN(ref[:idx], "/", 2)
		if len(parts) == 2 {
			n, err := strconv.Atoi(ref[idx+1:])
			if err != nil {
				return "", "", 0, fmt.Errorf("invalid PR number in %q", ref)
			}
			return parts[0], parts[1], n, nil
		}
	}

	if n, err := strconv.Atoi(ref); err == nil {
		return "", "", n, nil
	}

	return "", "", 0, fmt.Errorf("cannot parse PR reference %q: expected URL, owner/repo#number, or number", ref)
}

func FetchDiff(ctx context.Context, token, owner, repo string, number int) (*PRDiff, error) {
	gh := github.NewClient(nil).WithAuthToken(token)

	pr, _, err := gh.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("fetching PR: %w", err)
	}

	files, _, err := gh.PullRequests.ListFiles(ctx, owner, repo, number, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("fetching PR files: %w", err)
	}

	diff := &PRDiff{
		Title:       pr.GetTitle(),
		Description: pr.GetBody(),
		URL:         pr.GetHTMLURL(),
	}

	for _, f := range files {
		diff.Files = append(diff.Files, DiffFile{
			Path:   f.GetFilename(),
			Patch:  f.GetPatch(),
			Status: normalizeStatus(f.GetStatus()),
		})
	}

	return diff, nil
}

func normalizeStatus(s string) string {
	switch s {
	case "added":
		return "added"
	case "removed":
		return "removed"
	default:
		return "modified"
	}
}

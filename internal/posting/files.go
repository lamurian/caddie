package posting

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/lamurian/caddie/internal/workspace"
)

// postingDirs are where a posting can live during its lifecycle.
var postingDirs = []string{workspace.DirPostings, workspace.DirApplied, workspace.DirManual}

// findByURL looks for a saved posting with this URL. Files that do not
// parse are skipped here; the commands that own them report them.
func findByURL(l workspace.Layout, rawURL string) (Result, workspace.Posting, bool) {
	for _, dir := range postingDirs {
		paths, _ := filepath.Glob(filepath.Join(l.Dir(dir), "*.md"))
		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			p, _, err := workspace.ParsePosting(path, data)
			if err == nil && p.URL == rawURL {
				return Result{Slug: p.Slug, Path: path, Missing: []string{}}, p, true
			}
		}
	}
	return Result{}, workspace.Posting{}, false
}

// freeSlug returns base, or base-2, base-3… when a posting with that name
// already exists in postings/, applied/ or manual/.
func freeSlug(l workspace.Layout, base string) string {
	if base == "" {
		base = "posting"
	}
	for n := 1; ; n++ {
		slug := base
		if n > 1 {
			slug = fmt.Sprintf("%s-%d", base, n)
		}
		if !slugTaken(l, slug) {
			return slug
		}
	}
}

func slugTaken(l workspace.Layout, slug string) bool {
	for _, dir := range postingDirs {
		if _, err := os.Stat(filepath.Join(l.Dir(dir), slug+".md")); err == nil {
			return true
		}
	}
	return false
}

// findReview returns the body of reviews/<company-slug>.md, or of the
// review whose frontmatter names the company (case-insensitive), and the
// file it came from. No review is not an error.
func findReview(l workspace.Layout, company string) (string, string, error) {
	dir := l.Dir(workspace.DirReviews)
	if slug := workspace.Slug(company); slug != "" {
		path := filepath.Join(dir, slug+".md")
		if data, err := os.ReadFile(path); err == nil {
			_, body := splitReview(data)
			return body, path, nil
		}
	}
	paths, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", "", err
		}
		if name, body := splitReview(data); company != "" && strings.EqualFold(name, company) {
			return body, path, nil
		}
	}
	return "", "", nil
}

// splitReview returns a review's frontmatter company and its body.
func splitReview(data []byte) (company, body string) {
	text := string(bytes.TrimPrefix(data, []byte("\xef\xbb\xbf")))
	rest, ok := strings.CutPrefix(text, "---\n")
	if !ok {
		return "", strings.TrimSpace(text)
	}
	front, after, ok := strings.Cut(rest, "\n---")
	if !ok {
		return "", strings.TrimSpace(text)
	}
	var meta struct {
		Company string `yaml:"company"`
	}
	_ = yaml.Unmarshal([]byte(front), &meta)
	return strings.TrimSpace(meta.Company), strings.TrimSpace(after)
}

func reviewsSection(review, company string) string {
	if review == "" {
		review = fmt.Sprintf("_No reviews/%s.md at save time._", workspace.Slug(company))
	}
	return "## Reviews\n\n" + review + "\n"
}

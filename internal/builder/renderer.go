package builder

import (
	"embed"
	"os"
	"path/filepath"
	"text/template"

	"github.com/eng-graph/eng-graph/internal/profile"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

var funcMap = template.FuncMap{
	"add": func(a, b int) int { return a + b },
}

var templates = template.Must(
	template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.tmpl"),
)

func Render(p *profile.Profile, outputDir string) error {
	refsDir := filepath.Join(outputDir, "references")
	if err := os.MkdirAll(refsDir, 0755); err != nil {
		return err
	}

	files := []struct {
		tmpl string
		path string
	}{
		{"persona.md.tmpl", filepath.Join(outputDir, "persona.md")},
		{"review_patterns.md.tmpl", filepath.Join(refsDir, "review-patterns.md")},
		{"codebase_rules.md.tmpl", filepath.Join(refsDir, "codebase-rules.md")},
		{"voice.md.tmpl", filepath.Join(refsDir, "voice.md")},
	}

	for _, f := range files {
		if err := renderFile(f.tmpl, f.path, p); err != nil {
			return err
		}
	}
	return nil
}

func renderFile(tmpl, path string, p *profile.Profile) (retErr error) {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && retErr == nil {
			retErr = cerr
		}
	}()
	return templates.ExecuteTemplate(out, tmpl, p)
}

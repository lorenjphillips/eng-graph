package gitlab

import (
	"context"
	"fmt"

	"github.com/eng-graph/eng-graph/internal/source"
)

func init() {
	source.Register("gitlab", func(name string, cfg map[string]any) (source.SourceAdapter, error) {
		a := &Adapter{name: name}
		if v, ok := cfg["token_env"].(string); ok {
			a.tokenEnv = v
		}
		if v, ok := cfg["project"].(string); ok {
			a.project = v
		}
		return a, nil
	})
}

type Adapter struct {
	name     string
	tokenEnv string
	project  string
}

func (a *Adapter) Name() string { return a.name }
func (a *Adapter) Kind() string { return "gitlab" }

func (a *Adapter) Validate() error {
	if a.tokenEnv == "" {
		return fmt.Errorf("gitlab: token_env is required")
	}
	if a.project == "" {
		return fmt.Errorf("gitlab: project is required")
	}
	return nil
}

func (a *Adapter) TestConnection(ctx context.Context) error {
	return fmt.Errorf("gitlab: not yet implemented")
}

func (a *Adapter) Ingest(ctx context.Context, opts source.IngestOptions, out chan<- source.DataPoint, progress chan<- source.IngestProgress) error {
	return fmt.Errorf("gitlab: not yet implemented")
}

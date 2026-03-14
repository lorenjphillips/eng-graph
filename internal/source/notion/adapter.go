package notion

import (
	"context"
	"fmt"

	"github.com/eng-graph/eng-graph/internal/source"
)

func init() {
	source.Register("notion", func(name string, cfg map[string]any) (source.SourceAdapter, error) {
		a := &Adapter{name: name}
		if v, ok := cfg["token_env"].(string); ok {
			a.tokenEnv = v
		}
		if v, ok := cfg["database_ids"].([]any); ok {
			for _, id := range v {
				if s, ok := id.(string); ok {
					a.databaseIDs = append(a.databaseIDs, s)
				}
			}
		}
		return a, nil
	})
}

type Adapter struct {
	name        string
	tokenEnv    string
	databaseIDs []string
}

func (a *Adapter) Name() string { return a.name }
func (a *Adapter) Kind() string { return "notion" }

func (a *Adapter) Validate() error {
	if a.tokenEnv == "" {
		return fmt.Errorf("notion: token_env is required")
	}
	if len(a.databaseIDs) == 0 {
		return fmt.Errorf("notion: database_ids is required")
	}
	return nil
}

func (a *Adapter) TestConnection(ctx context.Context) error {
	return fmt.Errorf("notion: not yet implemented")
}

func (a *Adapter) Ingest(ctx context.Context, opts source.IngestOptions, out chan<- source.DataPoint, progress chan<- source.IngestProgress) error {
	return fmt.Errorf("notion: not yet implemented")
}

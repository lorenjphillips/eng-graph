package gdocs

import (
	"context"
	"fmt"

	"github.com/eng-graph/eng-graph/internal/source"
)

func init() {
	source.Register("gdocs", func(name string, cfg map[string]any) (source.SourceAdapter, error) {
		a := &Adapter{name: name}
		if v, ok := cfg["credentials_env"].(string); ok {
			a.credentialsEnv = v
		}
		if v, ok := cfg["document_ids"].([]any); ok {
			for _, id := range v {
				if s, ok := id.(string); ok {
					a.documentIDs = append(a.documentIDs, s)
				}
			}
		}
		return a, nil
	})
}

type Adapter struct {
	name           string
	credentialsEnv string
	documentIDs    []string
}

func (a *Adapter) Name() string { return a.name }
func (a *Adapter) Kind() string { return "gdocs" }

func (a *Adapter) Validate() error {
	if a.credentialsEnv == "" {
		return fmt.Errorf("gdocs: credentials_env is required")
	}
	if len(a.documentIDs) == 0 {
		return fmt.Errorf("gdocs: document_ids is required")
	}
	return nil
}

func (a *Adapter) TestConnection(ctx context.Context) error {
	return fmt.Errorf("gdocs: not yet implemented")
}

func (a *Adapter) Ingest(ctx context.Context, opts source.IngestOptions, out chan<- source.DataPoint, progress chan<- source.IngestProgress) error {
	return fmt.Errorf("gdocs: not yet implemented")
}

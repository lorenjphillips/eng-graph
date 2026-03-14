package confluence

import (
	"context"
	"fmt"

	"github.com/eng-graph/eng-graph/internal/source"
)

func init() {
	source.Register("confluence", func(name string, cfg map[string]any) (source.SourceAdapter, error) {
		a := &Adapter{name: name}
		if v, ok := cfg["base_url"].(string); ok {
			a.baseURL = v
		}
		if v, ok := cfg["token_env"].(string); ok {
			a.tokenEnv = v
		}
		if v, ok := cfg["space_keys"].([]any); ok {
			for _, k := range v {
				if s, ok := k.(string); ok {
					a.spaceKeys = append(a.spaceKeys, s)
				}
			}
		}
		return a, nil
	})
}

type Adapter struct {
	name      string
	baseURL   string
	tokenEnv  string
	spaceKeys []string
}

func (a *Adapter) Name() string { return a.name }
func (a *Adapter) Kind() string { return "confluence" }

func (a *Adapter) Validate() error {
	if a.baseURL == "" {
		return fmt.Errorf("confluence: base_url is required")
	}
	if a.tokenEnv == "" {
		return fmt.Errorf("confluence: token_env is required")
	}
	if len(a.spaceKeys) == 0 {
		return fmt.Errorf("confluence: space_keys is required")
	}
	return nil
}

func (a *Adapter) TestConnection(ctx context.Context) error {
	return fmt.Errorf("confluence: not yet implemented")
}

func (a *Adapter) Ingest(ctx context.Context, opts source.IngestOptions, out chan<- source.DataPoint, progress chan<- source.IngestProgress) error {
	return fmt.Errorf("confluence: not yet implemented")
}

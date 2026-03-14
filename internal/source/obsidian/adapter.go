package obsidian

import (
	"context"
	"fmt"

	"github.com/eng-graph/eng-graph/internal/source"
)

func init() {
	source.Register("obsidian", func(name string, cfg map[string]any) (source.SourceAdapter, error) {
		a := &Adapter{name: name}
		if v, ok := cfg["vault_path"].(string); ok {
			a.vaultPath = v
		}
		return a, nil
	})
}

type Adapter struct {
	name      string
	vaultPath string
}

func (a *Adapter) Name() string { return a.name }
func (a *Adapter) Kind() string { return "obsidian" }

func (a *Adapter) Validate() error {
	if a.vaultPath == "" {
		return fmt.Errorf("obsidian: vault_path is required")
	}
	return nil
}

func (a *Adapter) TestConnection(ctx context.Context) error {
	return fmt.Errorf("obsidian: not yet implemented")
}

func (a *Adapter) Ingest(ctx context.Context, opts source.IngestOptions, out chan<- source.DataPoint, progress chan<- source.IngestProgress) error {
	return fmt.Errorf("obsidian: not yet implemented")
}

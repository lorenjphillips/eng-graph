package ingest

import (
	"context"
	"fmt"
	"os"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/eng-graph/eng-graph/internal/source"
	"github.com/eng-graph/eng-graph/internal/storage"
)

type Pipeline struct {
	store *storage.SQLiteStore
	dedup *Deduplicator
}

func NewPipeline(store *storage.SQLiteStore) *Pipeline {
	return &Pipeline{
		store: store,
		dedup: NewDeduplicator(store),
	}
}

func (p *Pipeline) Run(ctx context.Context, profileName string, adapters []source.SourceAdapter, opts source.IngestOptions) (int, error) {
	dataCh := make(chan source.DataPoint, 256)
	progressCh := make(chan source.IngestProgress, 64)

	g, ctx := errgroup.WithContext(ctx)

	for _, a := range adapters {
		a := a
		g.Go(func() error {
			defer func() {
				progressCh <- source.IngestProgress{
					Source:  a.Name(),
					Message: "done",
				}
			}()
			return a.Ingest(ctx, opts, dataCh, progressCh)
		})
	}

	go func() {
		g.Wait()
		close(dataCh)
		close(progressCh)
	}()

	var progressWg sync.WaitGroup
	progressWg.Add(1)
	go func() {
		defer progressWg.Done()
		for prog := range progressCh {
			if prog.Message != "" {
				fmt.Fprintf(os.Stderr, "[%s] %s\n", prog.Source, prog.Message)
			} else {
				fmt.Fprintf(os.Stderr, "[%s] fetched %d\n", prog.Source, prog.Fetched)
			}
		}
	}()

	var total int
	var batch []source.DataPoint
	var consumeErr error
	const batchSize = 100

	for dp := range dataCh {
		if consumeErr != nil {
			continue // drain channel to unblock senders
		}
		if p.dedup.IsDuplicate(profileName, dp) {
			continue
		}
		batch = append(batch, dp)
		if len(batch) >= batchSize {
			if err := p.store.InsertBatch(profileName, batch); err != nil {
				consumeErr = err
				continue
			}
			total += len(batch)
			fmt.Fprintf(os.Stderr, "ingested %d data points\n", total)
			batch = batch[:0]
		}
	}

	if consumeErr != nil {
		progressWg.Wait()
		return total, consumeErr
	}

	if len(batch) > 0 {
		if err := p.store.InsertBatch(profileName, batch); err != nil {
			progressWg.Wait()
			return total, err
		}
		total += len(batch)
	}

	progressWg.Wait()

	if err := g.Wait(); err != nil {
		return total, err
	}

	fmt.Fprintf(os.Stderr, "total ingested: %d data points\n", total)
	return total, nil
}

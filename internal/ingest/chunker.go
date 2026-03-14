package ingest

import (
	"github.com/eng-graph/eng-graph/internal/source"
)

func ChunkDataPoints(dps []source.DataPoint, maxTokens int) [][]source.DataPoint {
	if maxTokens <= 0 {
		maxTokens = 100000
	}

	var chunks [][]source.DataPoint
	var current []source.DataPoint
	currentTokens := 0

	for _, dp := range dps {
		tokens := len(dp.Body) / 4
		if tokens == 0 {
			tokens = 1
		}

		if currentTokens+tokens > maxTokens && len(current) > 0 {
			chunks = append(chunks, current)
			current = nil
			currentTokens = 0
		}

		current = append(current, dp)
		currentTokens += tokens
	}

	if len(current) > 0 {
		chunks = append(chunks, current)
	}

	return chunks
}

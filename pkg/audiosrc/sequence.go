package audiosrc

import (
	"context"
	"github.com/gieseladev/elakshi/pkg/edb"
)

type searcherSequence []Searcher

func (s searcherSequence) Search(ctx context.Context, track edb.Track) ([]Result, error) {
	for _, searcher := range s {
		results, err := searcher.Search(ctx, track)
		if err != nil {
			return nil, err
		}

		if len(results) == 0 {
			continue
		}

		// TODO check whether results are good

		return results, err
	}

	return nil, nil
}

func CollectSearchers(searchers ...Searcher) Searcher {
	if len(searchers) == 1 {
		return searchers[0]
	}

	return searcherSequence(searchers)
}

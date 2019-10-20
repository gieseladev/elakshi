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

// CollectSearchers combines multiple Searchers into a single Searcher which
// uses them in sequence.
func CollectSearchers(searchers ...Searcher) Searcher {
	if len(searchers) == 1 {
		return searchers[0]
	}

	return searcherSequence(searchers)
}

// AppendSearcher adds searchers to a Searcher much like CollectSearchers does.
// The key difference is that when searcher is a sequence of searchers already,
// the new searchers are simply appended to that list.
func AppendSearcher(searcher Searcher, searchers ...Searcher) Searcher {
	if searcher == nil {
		return CollectSearchers(searchers...)
	}

	if searcher, ok := searcher.(searcherSequence); ok {
		return append(searcher, searchers...)
	}

	searchers = append(searchers, nil)
	copy(searchers[1:], searchers)
	searchers[0] = searcher

	return CollectSearchers(searchers...)
}

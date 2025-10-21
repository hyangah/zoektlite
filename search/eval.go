package search

import (
	"context"

	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/query"
)

// typeRepoSearcher evaluates all type:repo sub-queries before sending the query
// to the underlying searcher. We need to evaluate type:repo queries first
// since they need to do cross shard operations.
type typeRepoSearcher struct {
	zoekt.Streamer
}

func (s *typeRepoSearcher) Search(ctx context.Context, q query.Q, opts *zoekt.SearchOptions) (sr *zoekt.SearchResult, err error) {
	q, err = s.eval(ctx, q)
	if err != nil {
		return nil, err
	}

	return s.Streamer.Search(ctx, q, opts)
}

func (s *typeRepoSearcher) StreamSearch(ctx context.Context, q query.Q, opts *zoekt.SearchOptions, sender zoekt.Sender) (err error) {
	var stats zoekt.Stats
	q, err = s.eval(ctx, q)
	if err != nil {
		return err
	}

	return s.Streamer.StreamSearch(ctx, q, opts, zoekt.SenderFunc(func(event *zoekt.SearchResult) {
		stats.Add(event.Stats)
		sender.Send(event)
	}))
}

func (s *typeRepoSearcher) List(ctx context.Context, q query.Q, opts *zoekt.ListOptions) (rl *zoekt.RepoList, err error) {
	q, err = s.eval(ctx, q)
	if err != nil {
		return nil, err
	}

	return s.Streamer.List(ctx, q, opts)
}

func (s *typeRepoSearcher) eval(ctx context.Context, q query.Q) (query.Q, error) {
	var err error
	q = query.Map(q, func(q query.Q) query.Q {
		if err != nil {
			return nil
		}

		rq, ok := q.(*query.Type)
		if !ok || rq.Type != query.TypeRepo {
			return q
		}

		var rl *zoekt.RepoList
		rl, err = s.Streamer.List(ctx, rq.Child, nil)
		if err != nil {
			return nil
		}

		rs := &query.RepoSet{Set: make(map[string]bool, len(rl.Repos))}
		for _, r := range rl.Repos {
			rs.Set[r.Repository.Name] = true
		}

		return rs
	})
	return q, err
}

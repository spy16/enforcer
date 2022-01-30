package campaign

import (
	"context"
	"time"
)

// Store implementation provides storage for campaign data.
// Storage layer must ensure efficient query based on id, tags and
// start/end-at timestamps.
type Store interface {
	GetCampaign(ctx context.Context, id int) (*Campaign, error)
	ListCampaigns(ctx context.Context, q Query, p *Pager) ([]Campaign, error)
	CreateCampaign(ctx context.Context, camp Campaign) (int, error)
	UpdateCampaign(ctx context.Context, id int, updateFn UpdateFn) (*Campaign, error)
	DeleteCampaign(ctx context.Context, id int) error
}

// UpdateFn typed func value is used by campaign store to
// update an existing campaign atomically.
type UpdateFn func(ctx context.Context, actual *Campaign) error

// Pager represents pagination options.
type Pager struct {
	Offset  int `json:"offset,omitempty"`
	MaxSize int `json:"max_size,omitempty"`
}

// Query represents filtering options for the listing store.
// All criteria act in AND combination unless specified otherwise.
type Query struct {
	Include    []int    `json:"include,omitempty"`
	SearchIn   []int    `json:"search_in,omitempty"`
	OnlyActive bool     `json:"only_active,omitempty"`
	HavingTags []string `json:"having_tags,omitempty"`
}

func (q Query) filterCampaigns(arr []Campaign) []Campaign {
	searchSet := map[int]struct{}{}
	for _, id := range q.SearchIn {
		searchSet[id] = struct{}{}
	}

	var res []Campaign
	for _, camp := range arr {
		if q.matchQuery(camp) {
			res = append(res, camp)
		}
	}
	return res
}

func (q Query) matchQuery(c Campaign) bool {
	isMatch := !q.OnlyActive || c.IsActive(time.Now())
	if len(q.SearchIn) > 0 {
		found := false
		for _, id := range q.SearchIn {
			if id == c.ID {
				found = true
				break
			}
		}
		isMatch = isMatch && found
	}
	isMatch = isMatch && (len(q.HavingTags) == 0 || c.HasAllTags(q.HavingTags))
	return isMatch
}

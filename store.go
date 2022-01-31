package enforcer

import (
	"context"
	"time"
)

// Store implementation provides storage for campaign and enrolment data.
type Store interface {
	campaignStore
	enrolmentStore
}

type campaignStore interface {
	GetCampaign(ctx context.Context, id int) (*Campaign, error)
	ListCampaigns(ctx context.Context, q Query) ([]Campaign, error)
	CreateCampaign(ctx context.Context, camp Campaign) (int, error)
	UpdateCampaign(ctx context.Context, id int, updateFn UpdateCampaignFn) (*Campaign, error)
	DeleteCampaign(ctx context.Context, id int) error
}

type enrolmentStore interface {
	GetEnrolment(ctx context.Context, actorID string, campaignID int) (*Enrolment, error)
	ListEnrolments(ctx context.Context, actorID string, status []string) ([]Enrolment, error)
	CreateEnrolment(ctx context.Context, enrolment Enrolment) error
	UpdateEnrolment(ctx context.Context, actorID string, campaignID int, updateFn UpdateEnrolmentFn) (*Enrolment, error)
}

// UpdateCampaignFn typed func value is used by campaign store to
// update an existing campaign atomically.
type UpdateCampaignFn func(ctx context.Context, actual *Campaign) error

// UpdateEnrolmentFn is used by enrolment-store to perform updates atomically.
type UpdateEnrolmentFn func(ctx context.Context, enr *Enrolment) error

// Query represents filtering options for the listing store.
// All criteria act in AND combination unless specified otherwise.
type Query struct {
	Include    []int    `json:"include,omitempty"`
	SearchIn   []int    `json:"search_in,omitempty"`
	OnlyActive bool     `json:"only_active,omitempty"`
	HavingTags []string `json:"having_tags,omitempty"`

	// paging options.
	Offset  int `json:"offset,omitempty"`
	MaxSize int `json:"max_size,omitempty"`
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

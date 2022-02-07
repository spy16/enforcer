package enforcer

import (
	"context"
	"time"
)

// Store implementation provides storage layer for campaigns and enrolments.
type Store interface {
	CampaignStore
	EnrolmentStore
}

// CampaignStore implementation provides storage layer for campaigns.
type CampaignStore interface {
	GetCampaign(ctx context.Context, id string) (*Campaign, error)
	ListCampaigns(ctx context.Context, q Query) ([]Campaign, error)
	CreateCampaign(ctx context.Context, camp Campaign) error
	UpdateCampaign(ctx context.Context, id string, updateFn UpdateFn) (*Campaign, error)
	DeleteCampaign(ctx context.Context, id string) error
}

// EnrolmentStore implementation provides storage layer for enrolments.
type EnrolmentStore interface {
	GetEnrolment(ctx context.Context, actorID, campaignID string) (*Enrolment, error)
	ListEnrolments(ctx context.Context, actorID string) ([]Enrolment, error)
	UpsertEnrolment(ctx context.Context, enrolment Enrolment) error
}

// UpdateFn typed func value is used by campaign store to update
// an existing campaign atomically. UpdateFn should apply updates
// directly to the given campaign pointer.
type UpdateFn func(ctx context.Context, actual *Campaign) error

// Query represents filtering options for listing campaigns.
// Following criteria must be realised as:
// 	`Include + (SearchIn AND OnlyActive AND HavingTags)`
type Query struct {
	// Include campaigns with given ids unconditionally (i.e., other
	// filters do not apply to this).
	Include []string `json:"include,omitempty"`

	// SearchIn limits the search-space to given campaign IDs and
	// returns campaigns that match all other filters within this
	// list.
	SearchIn []string `json:"search_in,omitempty"`

	// OnlyActive signals to return only active campaigns.
	OnlyActive bool `json:"only_active,omitempty"`

	// HavingTags returns only those campaigns that have all the
	// given tags.
	HavingTags []string `json:"having_tags,omitempty"`
}

func (q Query) filterCampaigns(arr []Campaign) []Campaign {
	searchSet := map[string]struct{}{}
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
	isMatch = isMatch && (len(q.HavingTags) == 0 || c.HasTags(q.HavingTags))
	return isMatch
}

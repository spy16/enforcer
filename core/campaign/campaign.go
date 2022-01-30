package campaign

import (
	"context"
	"strings"
	"time"

	"github.com/spy16/enforcer"
)

// Campaign represents a group of rules that an actor needs to complete
// one-by-one.
type Campaign struct {
	ID            int       `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Name          string    `json:"name"`
	Enabled       bool      `json:"enabled"`
	Tags          []string  `json:"tags,omitempty"`
	Priority      int       `json:"priority"`
	StartAt       time.Time `json:"start_at"`
	EndAt         time.Time `json:"end_at"`
	Eligibility   string    `json:"eligibility"`
	MaxEnrolments int       `json:"max_enrolments"`
	RemEnrolments int       `json:"remaining_enrolments"`
	Deadline      int       `json:"deadline"`
	IsUnordered   bool      `json:"is_unordered"`
	Steps         []string  `json:"steps"`
}

// IsActive returns true if the campaign is still active relative
// to the given timestamp.
func (c Campaign) IsActive(at time.Time) bool {
	if !c.Enabled {
		return false
	}
	return c.StartAt.Before(at) && c.EndAt.Before(at)
}

// HasAllTags returns true if the campaign has all given tags.
func (c Campaign) HasAllTags(tags []string) bool {
	set := map[string]struct{}{}
	for _, tag := range c.Tags {
		set[tag] = struct{}{}
	}
	for _, tag := range tags {
		if _, found := set[tag]; !found {
			return false
		}
	}
	return true
}

// Validate sets defaults for the campaign and then performs validations.
func (c *Campaign) Validate() error {
	c.Name = strings.TrimSpace(c.Name)
	c.Tags = cleanTags(c.Tags)
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
		c.UpdatedAt = c.CreatedAt
	}
	c.Eligibility = strings.TrimSpace(c.Eligibility)

	if c.Eligibility == "" && len(c.Steps) == 0 {
		return enforcer.ErrInvalid.WithMsgf("at-least eligibility must be specified")
	}

	for i := range c.Steps {
		c.Steps[i] = strings.TrimSpace(c.Steps[i])
		if c.Steps[i] == "" {
			return enforcer.ErrInvalid.WithMsgf("step rule %d must not be empty", i)
		}
	}

	if c.Deadline < 0 {
		return enforcer.ErrInvalid.WithMsgf("deadline must be 0 or positive")
	}

	if c.Priority < 0 || c.Priority > 100 {
		return enforcer.ErrInvalid.WithMsgf("priority must be in range [0, 100]")
	}
	return nil
}

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

func cleanTags(tags []string) []string {
	var res []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			res = append(res, tag)
		}
	}
	return res
}

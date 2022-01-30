package enforcer

import (
	"strings"
	"time"
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
	CurEnrolments int       `json:"current_enrolments"`
	Deadline      int       `json:"deadline"`
	IsUnordered   bool      `json:"is_unordered"`
	Steps         []string  `json:"steps"`
}

// IsActive returns true if the campaign is still active relative
// to the given timestamp.
func (c Campaign) IsActive(at time.Time) bool {
	return c.Enabled &&
		c.StartAt.Before(at) && c.EndAt.After(at)
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
		return ErrInvalid.WithMsgf("at-least eligibility must be specified")
	}

	for i := range c.Steps {
		c.Steps[i] = strings.TrimSpace(c.Steps[i])
		if c.Steps[i] == "" {
			return ErrInvalid.WithMsgf("step rule %d must not be empty", i)
		}
	}

	if c.Deadline < 0 {
		return ErrInvalid.WithMsgf("deadline must be 0 or positive")
	}

	if c.Priority < 0 || c.Priority > 100 {
		return ErrInvalid.WithMsgf("priority must be in range [0, 100]")
	}
	return nil
}

func (c *Campaign) merge(partial Campaign) error {
	// TODO: merge partial into actual. Return error if non-overridable.

	isUsed := c.IsActive(time.Now()) && c.CurEnrolments > 0
	if isUsed {
		activeEnrErr := ErrInvalid.WithCausef("%d active enrolments", c.CurEnrolments)
		if len(partial.Steps) != 0 {
			return activeEnrErr.WithMsgf("steps cannot be edited")
		}

		if partial.Eligibility != "" {
			return activeEnrErr.WithMsgf("eligibility rule cannot be edited")
		}
	}

	if partial.Name != "" {
		c.Name = partial.Name
	}

	return c.Validate()
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

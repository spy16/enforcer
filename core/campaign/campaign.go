package campaign

import (
	"strings"
	"time"

	"github.com/spy16/enforcer"
)

// Campaign represents a group of rules that an actor needs to complete
// one-by-one.
type Campaign struct {
	Name      string    `json:"id"`
	Tags      []string  `json:"tags,omitempty"`
	Spec      Spec      `json:"spec"`
	Enabled   bool      `json:"enabled"`
	StartAt   time.Time `json:"start_at"`
	EndAt     time.Time `json:"end_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Spec represents a campaign specification.
type Spec struct {
	Steps         []string `json:"steps,omitempty"`
	Deadline      int      `json:"deadline,omitempty"`
	Priority      int      `json:"priority"`
	IsUnordered   bool     `json:"is_unordered"`
	Eligibility   string   `json:"eligibility,omitempty"`
	MaxEnrolments int      `json:"max_enrolments,omitempty"`
	CurEnrolments int      `json:"current_enrolments"`
}

// IsActive returns true if the campaign is active relative to the given
// timestamp.
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

// Validate performs validation of the entire campaign object. If checkSpec
// is true, spec is also validated.
func (c *Campaign) Validate(checkSpec bool) error {
	c.Name = strings.TrimSpace(c.Name)
	c.Tags = cleanTags(c.Tags)
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
		c.UpdatedAt = c.CreatedAt
	}

	if c.Name == "" {
		return enforcer.ErrInvalid.WithMsgf("a unique name must be set")
	}

	if c.StartAt.IsZero() {
		return enforcer.ErrInvalid.WithMsgf("start_at must be set")
	}

	if c.EndAt.IsZero() {
		return enforcer.ErrInvalid.WithMsgf("end_at must be set")
	}

	if checkSpec {
		return c.Spec.validate()
	}
	return nil
}

func (s *Spec) validate() error {
	s.Eligibility = strings.TrimSpace(s.Eligibility)

	if s.Eligibility == "" && len(s.Steps) == 0 {
		return enforcer.ErrInvalid.WithMsgf("at-least eligibility must be specified")
	}

	for i := range s.Steps {
		s.Steps[i] = strings.TrimSpace(s.Steps[i])
		if s.Steps[i] == "" {
			return enforcer.ErrInvalid.WithMsgf("step rule %d must not be empty", i)
		}
	}

	if s.Deadline < 0 {
		return enforcer.ErrInvalid.WithMsgf("deadline must be 0 or positive")
	}

	if s.Priority < 0 || s.Priority > 100 {
		return enforcer.ErrInvalid.WithMsgf("priority must be in range [0, 100]")
	}

	return nil
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
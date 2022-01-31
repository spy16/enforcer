package campaign

import (
	"sort"
	"strings"
	"time"

	"github.com/spy16/enforcer"
)

// Campaign represents a group of rules that an actor needs to complete
// one-by-one.
type Campaign struct {
	Spec          Spec      `json:"spec"`
	Name          string    `json:"name"`
	Scopes        []string  `json:"scope"`
	Enabled       bool      `json:"enabled"`
	StartAt       time.Time `json:"start_at"`
	EndAt         time.Time `json:"end_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ActorType     string    `json:"actor_type"`
	CurEnrolments int       `json:"current_enrolments"`
}

// Spec represents a campaign specification.
type Spec struct {
	Steps         []string `json:"steps,omitempty"`
	Deadline      int      `json:"deadline,omitempty"`
	Priority      int      `json:"priority"`
	IsUnordered   bool     `json:"is_unordered"`
	Eligibility   string   `json:"eligibility,omitempty"`
	MaxEnrolments int      `json:"max_enrolments,omitempty"`
}

// IsActive returns true if the campaign is active relative to the given
// timestamp.
func (c Campaign) IsActive(at time.Time) bool {
	return c.Enabled &&
		c.StartAt.Before(at) && c.EndAt.After(at)
}

// HasScope returns true if the campaign has all given scope-tags.
func (c Campaign) HasScope(scope []string) bool {
	set := map[string]struct{}{}
	for _, scopeTag := range c.Scopes {
		set[scopeTag] = struct{}{}
	}

	for _, scopeTag := range scope {
		if _, found := set[scopeTag]; !found {
			return false
		}
	}
	return true
}

// Validate performs validation of the entire campaign object. If checkSpec
// is true, spec is also validated.
func (c *Campaign) Validate(checkSpec bool) error {
	c.Name = strings.TrimSpace(c.Name)
	c.Scopes = cleanScope(c.Scopes)
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

func (c *Campaign) merge(spec Spec) error {
	// TODO: merge partial into actual. Return error if non-overridable.

	isUsed := c.IsActive(time.Now()) && c.CurEnrolments > 0
	if isUsed {
		activeEnrErr := enforcer.ErrInvalid.WithCausef("%d active enrolments", c.CurEnrolments)
		if len(spec.Steps) != 0 {
			return activeEnrErr.WithMsgf("steps cannot be edited")
		}

		if spec.Eligibility != "" {
			return activeEnrErr.WithMsgf("eligibility rule cannot be edited")
		}
	}

	return c.Validate(true)
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

func cleanScope(tags []string) []string {
	var res []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			res = append(res, tag)
		}
	}
	sort.Slice(tags, func(i, j int) bool {
		return strings.Compare(tags[i], tags[j]) > 0
	})
	return res
}

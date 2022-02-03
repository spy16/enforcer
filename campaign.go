package enforcer

import (
	"sort"
	"strings"
	"time"
)

// Campaign represents a group of rules that an actor needs to complete
// one-by-one.
type Campaign struct {
	Name          string    `json:"name"`
	Scopes        []string  `json:"scope"`
	Enabled       bool      `json:"enabled"`
	StartAt       time.Time `json:"start_at"`
	EndAt         time.Time `json:"end_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	CurEnrolments int       `json:"current_enrolments"`

	// campaign configurations.
	Steps         []string `json:"steps,omitempty"`
	Deadline      int      `json:"deadline,omitempty"`
	Priority      int      `json:"priority"`
	IsUnordered   bool     `json:"is_unordered"`
	Eligibility   string   `json:"eligibility,omitempty"`
	MaxEnrolments int      `json:"max_enrolments,omitempty"`
}

// Updates represents updates that can be applied on a campaign.
type Updates struct {
	StartAt       *time.Time `json:"start_at,omitempty"`
	EndAt         *time.Time `json:"end_at,omitempty"`
	Scopes        []string   `json:"scopes,omitempty"`
	Enabled       *bool      `json:"enabled,omitempty"`
	Steps         []string   `json:"steps,omitempty"`
	Deadline      *int       `json:"deadline,omitempty"`
	Priority      *int       `json:"priority"`
	IsUnordered   *bool      `json:"is_unordered"`
	Eligibility   string     `json:"eligibility,omitempty"`
	MaxEnrolments *int       `json:"max_enrolments,omitempty"`
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
func (c *Campaign) Validate() error {
	c.Name = strings.TrimSpace(c.Name)
	c.Scopes = cleanScope(c.Scopes)
	c.Eligibility = strings.TrimSpace(c.Eligibility)
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
		c.UpdatedAt = c.CreatedAt
	}

	if c.Name == "" {
		return ErrInvalid.WithMsgf("a unique name must be set")
	}

	if c.StartAt.IsZero() {
		return ErrInvalid.WithMsgf("start_at must be set")
	}

	if c.EndAt.IsZero() {
		return ErrInvalid.WithMsgf("end_at must be set")
	}

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

func (c *Campaign) apply(updates Updates) error {
	isUsed := c.IsActive(time.Now()) && c.CurEnrolments > 0
	activeEnrErr := ErrInvalid.WithCausef("%d active enrolments", c.CurEnrolments)

	if updates.Enabled != nil {
		c.Enabled = *updates.Enabled
	}
	if updates.StartAt != nil {
		if isUsed {
			return activeEnrErr.WithMsgf("start-date cannot be modified")
		}
		c.StartAt = *updates.StartAt
	}

	if updates.EndAt != nil {
		if isUsed {
			return activeEnrErr.WithMsgf("start-date cannot be modified")
		}
		c.EndAt = *updates.EndAt
	}

	if len(updates.Scopes) != 0 {
		set := map[string]bool{}
		for _, scope := range updates.Scopes {
			scope = strings.TrimSpace(scope)
			if strings.HasPrefix(scope, "-") {
				scope = scope[1:]
				set[scope] = false
			} else {
				if strings.HasPrefix(scope, "+") {
					scope = scope[1:]
				}
				set[scope] = true
			}
		}

		var scopes []string
		for _, scope := range c.Scopes {
			isRetain, found := set[scope]
			if !found || isRetain {
				scopes = append(scopes, scope)
			}
		}

		for scope, add := range set {
			if !add {
				continue
			}
			scopes = append(scopes, scope)
		}
		c.Scopes = scopes
	}

	if updates.IsUnordered != nil {
		c.IsUnordered = *updates.IsUnordered
	}
	if updates.Priority != nil {
		c.Priority = *updates.Priority
	}

	if updates.Deadline != nil {
		if isUsed {
			return activeEnrErr.WithMsgf("deadline cannot be edited")
		}
		c.Deadline = *updates.Deadline
	}

	if c.Eligibility != "" {
		if isUsed {
			return activeEnrErr.WithMsgf("eligibility rule cannot be edited")
		}
		c.Eligibility = updates.Eligibility
	}

	if len(updates.Steps) != 0 {
		if isUsed {
			return activeEnrErr.WithMsgf("steps cannot be edited")
		}
		c.Steps = updates.Steps
	}

	if updates.MaxEnrolments != nil {
		if *updates.MaxEnrolments < c.CurEnrolments {
			return activeEnrErr.WithMsgf("max-enrolments cannot be updated to lesser value")
		}
		c.MaxEnrolments = *updates.MaxEnrolments
	}

	return c.Validate()
}

func cleanScope(tags []string) []string {
	set := map[string]struct{}{}
	var res []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			if _, found := set[tag]; !found {
				res = append(res, tag)
				set[tag] = struct{}{}
			}
		}
	}
	sort.Slice(tags, func(i, j int) bool {
		return strings.Compare(tags[i], tags[j]) > 0
	})
	return res
}

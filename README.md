# Enforcer

## Concepts

```golang
package campaign

import "time"

type Campaign struct {
	ID        int       `json:"id"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name        string    `json:"name"`
	Tags        []string  `json:"tags,omitempty"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Deadline    int       `json:"deadline"`
	Priority    int       `json:"priority"`
	Eligibility string    `json:"eligibility"`
	Ordered     bool      `json:"ordered"`
	Rules       []string  `json:"rules"`
}

type Enrolment struct {
	ActorID    string       `json:"actor_id"`
	CampaignID int          `json:"campaign_id"`
	StartedAt  time.Time    `json:"started_at"`
	EndsAt     time.Time    `json:"ends_at"`
	Passed     []RuleResult `json:"passed"`
}

type RuleResult struct {
	RuleID   int       `json:"rule_id"`
	ActionID string    `json:"action_id"`
	PassedAt time.Time `json:"passed_at"`
}

```
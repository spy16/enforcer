package enforcer

import "time"

const (
	StatusActive    = "ACTIVE"
	StatusExpired   = "EXPIRED"
	StatusEligible  = "ELIGIBLE"
	StatusCompleted = "COMPLETED"
)

// Enrolment represents a binding between an actor and a campaign
// and contains the progress through the campaign.
type Enrolment struct {
	Status         string       `json:"status"`
	ActorID        string       `json:"actor_id"`
	CampaignID     int          `json:"campaign_id"`
	StartedAt      time.Time    `json:"started_at"`
	EndsAt         time.Time    `json:"ends_at"`
	CompletedSteps []StepResult `json:"completed_steps"`
	RemainingSteps int          `json:"remaining_steps"`
	Campaign       *Campaign    `json:"-"`
}

// StepResult represents a rule that was completed by an actor.
type StepResult struct {
	StepID   int       `json:"step_id"`
	ActionID string    `json:"action_id"`
	PassedAt time.Time `json:"passed_at"`
}

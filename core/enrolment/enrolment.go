package enrolment

import (
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

const (
	StatusActive    = "ACTIVE"
	StatusExpired   = "EXPIRED"
	StatusEligible  = "ELIGIBLE"
	StatusCompleted = "COMPLETED"
)

var val = validator.New()

// Enrolment represents a binding between an actor & a campaign, and
// also contains the progress of the actor in the campaign.
type Enrolment struct {
	ActorID        string       `json:"actor_id" validate:"required"`
	CampaignID     string       `json:"campaign_id" validate:"required,alphanum"`
	Status         string       `json:"status" validate:"alpha,uppercase"`
	StartedAt      time.Time    `json:"started_at,omitempty"`
	EndsAt         time.Time    `json:"ends_at,omitempty"`
	RemainingSteps int          `json:"remaining_steps,omitempty"`
	CompletedSteps []StepResult `json:"completed_steps,omitempty"`
}

// StepResult represents a campaign step that was completed by an
// actor.
type StepResult struct {
	StepID   int       `json:"step_id" validate:"lte=0"`
	DoneAt   time.Time `json:"done_at" validate:"required"`
	ActionID string    `json:"action_id" validate:"required"`
}

func (enr *Enrolment) setStatus() {
	if enr.StartedAt.IsZero() {
		enr.Status = StatusEligible
	} else if enr.RemainingSteps == 0 {
		enr.Status = StatusCompleted
	} else if enr.EndsAt.Before(time.Now()) {
		enr.Status = StatusExpired
	} else {
		enr.Status = StatusActive
	}
}

func (enr *Enrolment) validate() error {
	enr.ActorID = strings.TrimSpace(enr.ActorID)
	enr.CampaignID = strings.TrimSpace(enr.CampaignID)
	enr.setStatus()

	for _, step := range enr.CompletedSteps {
		step.ActionID = strings.TrimSpace(step.ActionID)
		if err := val.Struct(step); err != nil {
			return err
		}
	}
	return val.Struct(enr)
}

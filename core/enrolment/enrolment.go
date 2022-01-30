package enrolment

import (
	"context"
	"time"

	"github.com/spy16/enforcer/core/campaign"
)

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
}

// StepResult represents a rule that was completed by an actor.
type StepResult struct {
	StepID   int       `json:"step_id"`
	ActionID string    `json:"action_id"`
	PassedAt time.Time `json:"passed_at"`
}

// Store implementation provides storage for store.
// Storage layer must ensure efficient query based on actor_id,
// campaign_id and ends_at.
type Store interface {
	GetEnrolment(ctx context.Context, actorID string, campaignID int) (*Enrolment, error)
	ListEnrolments(ctx context.Context, actorID string, status []string) ([]Enrolment, error)
	CreateEnrolment(ctx context.Context, enrolment Enrolment) error
	UpdateEnrolment(ctx context.Context, actorID string, campaignID int, updateFn UpdateFn) (*Enrolment, error)
}

type Query struct {
	Status    []string       `json:"status"`
	Campaigns campaign.Query `json:"campaigns"`
}

// UpdateFn is used by enrolment-store to perform updates atomically.
type UpdateFn func(ctx context.Context, enr *Enrolment) error

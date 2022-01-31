package enrolment

import "context"

// Store implementation provides storage layer for enrolments.
type Store interface {
	GetEnrolment(ctx context.Context, actorID, campaignID string) (*Enrolment, error)
	ListEnrolments(ctx context.Context, actorID string) ([]Enrolment, error)
	CreateEnrolment(ctx context.Context, enrolment Enrolment) error
	UpdateEnrolment(ctx context.Context, actorID string, campaignID int, updateFn UpdateFn) (*Enrolment, error)
}

// UpdateFn is used by enrolment-store to perform updates on enrolments
// atomically.
type UpdateFn func(ctx context.Context, enr *Enrolment) error

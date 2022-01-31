package enrolment

import "context"

// Store implementation provides storage layer for enrolments.
type Store interface {
	GetEnrolment(ctx context.Context, actorID, campaignID string) (*Enrolment, error)
	ListEnrolments(ctx context.Context, actorID string) ([]Enrolment, error)
	UpsertEnrolment(ctx context.Context, enrolment Enrolment) error
}

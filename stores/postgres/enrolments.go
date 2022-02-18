package postgres

import (
	"context"

	"github.com/spy16/enforcer"
)

func (st *Store) GetEnrolment(ctx context.Context, actorID, campaignID string) (*enforcer.Enrolment, error) {
	// TODO implement me
	panic("implement me")
}

func (st *Store) ListEnrolments(ctx context.Context, actorID string) ([]enforcer.Enrolment, error) {
	// TODO implement me
	panic("implement me")
}

func (st *Store) UpsertEnrolment(ctx context.Context, enrolment enforcer.Enrolment) error {
	// TODO implement me
	panic("implement me")
}

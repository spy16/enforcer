package postgres

import (
	"context"

	"github.com/spy16/enforcer"
)

func (st *Store) GetCampaign(ctx context.Context, id string) (*enforcer.Campaign, error) {
	// TODO implement me
	panic("implement me")
}

func (st *Store) ListCampaigns(ctx context.Context, q enforcer.Query) ([]enforcer.Campaign, error) {
	// TODO implement me
	panic("implement me")
}

func (st *Store) CreateCampaign(ctx context.Context, camp enforcer.Campaign) error {
	// TODO implement me
	panic("implement me")
}

func (st *Store) UpdateCampaign(ctx context.Context, id string, updateFn enforcer.UpdateFn) (*enforcer.Campaign, error) {
	// TODO implement me
	panic("implement me")
}

func (st *Store) DeleteCampaign(ctx context.Context, id string) error {
	// TODO implement me
	panic("implement me")
}

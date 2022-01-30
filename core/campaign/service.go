package campaign

import (
	"context"

	"github.com/spy16/enforcer"
)

// Service provides APIs for managing campaigns.
type Service struct{ Store Store }

// GetCampaign returns campaign with given ID. Returns ErrNotFound if not found.
func (en *Service) GetCampaign(ctx context.Context, campaignID int) (*Campaign, error) {
	if campaignID <= 0 {
		return nil, enforcer.ErrInvalid.WithMsgf("invalid id: %d", campaignID)
	}

	res, err := en.Store.GetCampaign(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ListCampaigns returns a list of campaigns matching the given search query.
func (en *Service) ListCampaigns(ctx context.Context, q Query) ([]Campaign, error) {
	res, err := en.Store.ListCampaigns(ctx, q, nil)
	if err != nil {
		return nil, err
	}
	return q.filterCampaigns(res), nil
}

// CreateCampaign validates and inserts a new campaign into the storage.
// Campaign ID is assigned automatically and the stored version of the
// campaign is returned.
func (en *Service) CreateCampaign(ctx context.Context, camp Campaign) (*Campaign, error) {
	if err := camp.Validate(); err != nil {
		return nil, err
	}

	id, err := en.Store.CreateCampaign(ctx, camp)
	if err != nil {
		return nil, err
	}
	camp.ID = id

	return &camp, nil
}

// UpdateCampaign merges the given partial campaign object with the existing
// campaign and stores. The updated version is returned. Some fields may not
// undergo update based on current usage status.
func (en *Service) UpdateCampaign(ctx context.Context, partial Campaign) (*Campaign, error) {
	if partial.ID <= 0 {
		return nil, enforcer.ErrInvalid.WithMsgf("invalid id: %d", partial.ID)
	}

	updateFn := func(ctx context.Context, actual *Campaign) error {
		return actual.merge(partial)
	}

	return en.Store.UpdateCampaign(ctx, partial.ID, updateFn)
}

// DeleteCampaign deletes a campaign by the identifier.
func (en *Service) DeleteCampaign(ctx context.Context, campaignID int) error {
	if campaignID <= 0 {
		return enforcer.ErrInvalid.WithMsgf("invalid id: %d", campaignID)
	}
	return en.Store.DeleteCampaign(ctx, campaignID)
}

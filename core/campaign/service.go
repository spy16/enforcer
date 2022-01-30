package campaign

import (
	"context"

	"github.com/spy16/enforcer"
)

type Service struct {
	Store Store
}

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

func (en *Service) ListCampaigns(ctx context.Context, q Query) ([]Campaign, error) {
	res, err := en.Store.ListCampaigns(ctx, q, nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}

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

func (en *Service) UpdateCampaign(ctx context.Context, partial Campaign) (*Campaign, error) {
	if partial.ID <= 0 {
		return nil, enforcer.ErrInvalid.WithMsgf("invalid id: %d", partial.ID)
	}

	return en.Store.UpdateCampaign(ctx, partial.ID, func(ctx context.Context, actual *Campaign) error {
		return applyUpdate(actual, partial)
	})
}

func (en *Service) DeleteCampaign(ctx context.Context, campaignID int) error {
	if campaignID <= 0 {
		return enforcer.ErrInvalid.WithMsgf("invalid id: %d", campaignID)
	}

	return en.Store.DeleteCampaign(ctx, campaignID)
}

func applyUpdate(actual *Campaign, partial Campaign) error {
	// TODO: merge partial into actual. Return error if non-overridable.
	return actual.Validate()
}

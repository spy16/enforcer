package campaign

import (
	"context"
	"time"

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

	updateFn := func(ctx context.Context, actual *Campaign) error {
		return en.mergePartial(actual, partial)
	}

	return en.Store.UpdateCampaign(ctx, partial.ID, updateFn)
}

func (en *Service) DeleteCampaign(ctx context.Context, campaignID int) error {
	if campaignID <= 0 {
		return enforcer.ErrInvalid.WithMsgf("invalid id: %d", campaignID)
	}
	return en.Store.DeleteCampaign(ctx, campaignID)
}

func (en *Service) mergePartial(actual *Campaign, partial Campaign) error {
	// TODO: merge partial into actual. Return error if non-overridable.

	isUsed := actual.IsActive(time.Now()) && actual.CurEnrolments > 0
	if isUsed {
		activeEnrErr := enforcer.ErrInvalid.WithCausef("%d active enrolments", actual.CurEnrolments)
		if len(partial.Steps) != 0 {
			return activeEnrErr.WithMsgf("steps cannot be edited")
		}

		if partial.Eligibility != "" {
			return activeEnrErr.WithMsgf("eligibility rule cannot be edited")
		}
	}

	if partial.Name != "" {
		actual.Name = partial.Name
	}

	return actual.Validate()
}

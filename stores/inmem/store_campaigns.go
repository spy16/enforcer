package inmem

import (
	"context"

	"github.com/spy16/enforcer"
)

func (mem *Store) GetCampaign(ctx context.Context, campaignID int) (*enforcer.Campaign, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	c, found := mem.campaigns[campaignID]
	if !found {
		return nil, enforcer.ErrNotFound.WithMsgf("campaign %d not found", campaignID)
	}
	return &c, nil
}

func (mem *Store) ListCampaigns(ctx context.Context, q enforcer.Query) ([]enforcer.Campaign, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	var res []enforcer.Campaign
	for _, c := range mem.campaigns {
		res = append(res, c)
	}

	return res, nil
}

func (mem *Store) CreateCampaign(ctx context.Context, c enforcer.Campaign) (int, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	if mem.campaigns == nil {
		mem.campaigns = map[int]enforcer.Campaign{}
		mem.nextID = 1
	}
	// create a new one.
	c.ID = mem.nextID
	mem.nextID++
	mem.campaigns[c.ID] = c

	return c.ID, nil
}

func (mem *Store) UpdateCampaign(ctx context.Context, id int, updateFn enforcer.UpdateCampaignFn) (*enforcer.Campaign, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()
	c, found := mem.campaigns[id]
	if !found {
		return nil, enforcer.ErrNotFound.WithMsgf("campaign %d not found", id)
	}

	if err := updateFn(ctx, &c); err != nil {
		return nil, err
	}
	c.ID = id
	mem.campaigns[id] = c

	return &c, nil
}

func (mem *Store) DeleteCampaign(ctx context.Context, campaignID int) error {
	if len(mem.campaigns) == 0 {
		return nil
	}

	mem.mu.Lock()
	defer mem.mu.Unlock()

	delete(mem.campaigns, campaignID)
	return nil
}

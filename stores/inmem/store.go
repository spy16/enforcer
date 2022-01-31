package inmem

import (
	"context"
	"sync"

	"github.com/spy16/enforcer"
	"github.com/spy16/enforcer/core/campaign"
	"github.com/spy16/enforcer/core/enrolment"
)

var _ campaign.Store = (*Store)(nil)
var _ enrolment.Store = (*Store)(nil)

type Store struct {
	mu         sync.RWMutex
	campaigns  map[string]campaign.Campaign
	enrolments map[string]map[string]enrolment.Enrolment
}

func (mem *Store) GetCampaign(ctx context.Context, name string) (*campaign.Campaign, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	c, found := mem.campaigns[name]
	if !found {
		return nil, enforcer.ErrNotFound
	}
	return &c, nil
}

func (mem *Store) ListCampaigns(ctx context.Context, q campaign.Query) ([]campaign.Campaign, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	var res []campaign.Campaign
	for _, c := range mem.campaigns {
		res = append(res, c)
	}

	return res, nil
}

func (mem *Store) CreateCampaign(ctx context.Context, c campaign.Campaign) error {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	if mem.campaigns == nil {
		mem.campaigns = map[string]campaign.Campaign{}
	}
	mem.campaigns[c.Name] = c

	return nil
}

func (mem *Store) UpdateCampaign(ctx context.Context, name string, updateFn campaign.UpdateFn) (*campaign.Campaign, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	c, found := mem.campaigns[name]
	if !found {
		return nil, enforcer.ErrNotFound
	}

	if err := updateFn(ctx, &c); err != nil {
		return nil, err
	}
	mem.campaigns[name] = c

	return &c, nil
}

func (mem *Store) DeleteCampaign(ctx context.Context, campaignID string) error {
	if len(mem.campaigns) == 0 {
		return nil
	}

	mem.mu.Lock()
	defer mem.mu.Unlock()

	delete(mem.campaigns, campaignID)
	return nil
}

func (mem *Store) GetEnrolment(ctx context.Context, actorID, campaignID string) (*enrolment.Enrolment, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	e, found := mem.enrolments[actorID][campaignID]
	if !found {
		return nil, enforcer.ErrNotFound.
			WithMsgf("enrolment for actor '%s' and campaign '%s'", actorID, campaignID)
	}
	return &e, nil
}

func (mem *Store) ListEnrolments(ctx context.Context, actorID string) ([]enrolment.Enrolment, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	var res []enrolment.Enrolment
	for _, enr := range mem.enrolments[actorID] {
		res = append(res, enr)
	}
	return res, nil
}

func (mem *Store) UpsertEnrolment(ctx context.Context, enr enrolment.Enrolment) error {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	_, found := mem.enrolments[enr.ActorID][enr.CampaignID]
	if !found {
		if mem.enrolments == nil {
			mem.enrolments = map[string]map[string]enrolment.Enrolment{}
		}
		if _, found := mem.enrolments[enr.ActorID]; !found {
			mem.enrolments[enr.ActorID] = map[string]enrolment.Enrolment{}
		}

		camp := mem.campaigns[enr.CampaignID]
		camp.CurEnrolments++
		mem.campaigns[enr.CampaignID] = camp
	}

	mem.enrolments[enr.ActorID][enr.CampaignID] = enr
	return nil
}

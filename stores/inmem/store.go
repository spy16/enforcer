package inmem

import (
	"context"
	"sync"

	"github.com/spy16/enforcer"
)

var _ enforcer.Store = (*Store)(nil)

type Store struct {
	mu         sync.RWMutex
	nextID     int
	campaigns  map[int]enforcer.Campaign
	enrolments map[string]map[int]enforcer.Enrolment
}

func (mem *Store) GetCampaign(ctx context.Context, id int) (*enforcer.Campaign, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	c, found := mem.campaigns[id]
	if !found {
		return nil, enforcer.ErrNotFound
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
	c.ID = mem.nextID
	mem.campaigns[c.ID] = c
	mem.nextID++

	return c.ID, nil
}

func (mem *Store) UpdateCampaign(ctx context.Context, id int, updateFn enforcer.UpdateFn) (*enforcer.Campaign, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	c, found := mem.campaigns[id]
	if !found {
		return nil, enforcer.ErrNotFound
	}

	if err := updateFn(ctx, &c); err != nil {
		return nil, err
	}
	mem.campaigns[id] = c

	return &c, nil
}

func (mem *Store) DeleteCampaign(ctx context.Context, id int) error {
	if len(mem.campaigns) == 0 {
		return nil
	}

	mem.mu.Lock()
	defer mem.mu.Unlock()

	delete(mem.campaigns, id)
	return nil
}

func (mem *Store) GetEnrolment(ctx context.Context, actorID string, campaignID int) (*enforcer.Enrolment, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	e, found := mem.enrolments[actorID][campaignID]
	if !found {
		return nil, enforcer.ErrNotFound.
			WithMsgf("enrolment for actor '%s' and campaign '%d'", actorID, campaignID)
	}
	return &e, nil
}

func (mem *Store) ListEnrolments(ctx context.Context, actorID string) ([]enforcer.Enrolment, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	var res []enforcer.Enrolment
	for _, enr := range mem.enrolments[actorID] {
		res = append(res, enr)
	}
	return res, nil
}

func (mem *Store) UpsertEnrolment(ctx context.Context, enr enforcer.Enrolment) error {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	_, found := mem.enrolments[enr.ActorID][enr.CampaignID]
	if !found {
		if mem.enrolments == nil {
			mem.enrolments = map[string]map[int]enforcer.Enrolment{}
		}
		if _, found := mem.enrolments[enr.ActorID]; !found {
			mem.enrolments[enr.ActorID] = map[int]enforcer.Enrolment{}
		}

		camp := mem.campaigns[enr.CampaignID]
		camp.CurEnrolments++
		mem.campaigns[enr.CampaignID] = camp
	}

	mem.enrolments[enr.ActorID][enr.CampaignID] = enr
	return nil
}

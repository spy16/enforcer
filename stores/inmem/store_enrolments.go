package inmem

import (
	"context"

	"github.com/spy16/enforcer"
	"github.com/spy16/enforcer/core/enrolment"
)

func (mem *Store) GetEnrolment(ctx context.Context, actorID string, campaignID int) (*enrolment.Enrolment, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	e, found := mem.enrolments[actorID][campaignID]
	if !found {
		return nil, enforcer.ErrNotFound.
			WithMsgf("enrolment for actor '%s' and campaign %d", actorID, campaignID)
	}
	return &e, nil
}

func (mem *Store) ListEnrolments(ctx context.Context, actorID string, status []string) ([]enrolment.Enrolment, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	var res []enrolment.Enrolment
	for _, enr := range mem.enrolments[actorID] {
		if len(status) == 0 || contains(status, enr.Status) {
			res = append(res, enr)
		}
	}
	return res, nil
}

func (mem *Store) CreateEnrolment(ctx context.Context, enr enrolment.Enrolment) error {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	if mem.enrolments == nil {
		mem.enrolments = map[string]map[int]enrolment.Enrolment{}
	}
	if _, found := mem.enrolments[enr.ActorID]; !found {
		mem.enrolments[enr.ActorID] = map[int]enrolment.Enrolment{}
	}

	mem.enrolments[enr.ActorID][enr.CampaignID] = enr
	return nil
}

func (mem *Store) UpdateEnrolment(ctx context.Context, actorID string, campaignID int, updateFn enrolment.UpdateFn) (*enrolment.Enrolment, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	enr, found := mem.enrolments[actorID][campaignID]
	if !found {
		return nil, enforcer.ErrNotFound.
			WithMsgf("enrolment for actor '%s' and campaign %d", actorID, campaignID)
	}

	if err := updateFn(ctx, &enr); err != nil {
		return nil, err
	}
	mem.enrolments[enr.ActorID][enr.CampaignID] = enr

	return &enr, nil
}

func contains(arr []string, item string) bool {
	for _, s := range arr {
		if s == item {
			return true
		}
	}
	return false
}
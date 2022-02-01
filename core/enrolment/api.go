package enrolment

import (
	"context"
	"errors"
	"time"

	"github.com/spy16/enforcer"
	"github.com/spy16/enforcer/core/actor"
	"github.com/spy16/enforcer/core/campaign"
)

// API provides functions for managing enrolments.
type API struct {
	Store        Store
	RuleEngine   ruleEngine
	CampaignsAPI campaignsAPI
}

type campaignsAPI interface {
	Get(ctx context.Context, name string) (*campaign.Campaign, error)
	List(ctx context.Context, q campaign.Query) ([]campaign.Campaign, error)
}

type ruleEngine interface {
	Exec(_ context.Context, rule string, data interface{}) (bool, error)
}

// Get returns an enrolment for campaign and an actor. If actor is not already
// enrolled into the campaign and is eligible, a virtual enrolment with status
// StatusEligible is returned.
func (api *API) Get(ctx context.Context, campaignName string, ac actor.Actor) (*Enrolment, error) {
	enr, err := api.Store.GetEnrolment(ctx, ac.ID, campaignName)
	if !errors.Is(err, enforcer.ErrNotFound) {
		enr.setStatus()
		return enr, err
	}

	camp, err := api.CampaignsAPI.Get(ctx, campaignName)
	if err != nil {
		return nil, err
	}

	return api.prepEnrolment(ctx, *camp, ac)
}

// ListExisting returns a list of existing enrolments in one of given statuses.
// The returned list will not include eligible enrolments.
func (api *API) ListExisting(ctx context.Context, actorID string, status []string) ([]Enrolment, error) {
	existing, err := api.Store.ListEnrolments(ctx, actorID)
	if err != nil {
		return nil, err
	}
	return filterByStatus(existing, status), nil
}

// ListAll returns all enrolments including existing and eligible. Eligible
// are computed based on the campaign query and actor data provided.
func (api *API) ListAll(ctx context.Context, ac actor.Actor, campQ campaign.Query) ([]Enrolment, error) {
	existing, err := api.ListExisting(ctx, ac.ID, nil)
	if err != nil {
		return nil, err
	}

	campQ.OnlyActive = true
	campQ.Include = collectCampaignIDs(existing)
	camps, err := api.CampaignsAPI.List(ctx, campQ)
	if err != nil {
		return nil, err
	}

	var res []Enrolment
	alreadyEnrolled := map[string]struct{}{}
	for _, enrolment := range existing {
		enrolment.setStatus()
		alreadyEnrolled[enrolment.CampaignID] = struct{}{}
		res = append(res, enrolment)
	}

	for _, camp := range camps {
		if _, exists := alreadyEnrolled[camp.Name]; exists {
			continue
		}

		enr, err := api.prepEnrolment(ctx, camp, ac)
		if err != nil {
			if errors.Is(err, enforcer.ErrIneligible) {
				continue
			}
			return nil, err
		}
		res = append(res, *enr)
	}
	return res, nil
}

// Enrol binds the given actor to the campaign. Boolean flag will be set only if
// a new enrolment is created.
func (api *API) Enrol(ctx context.Context, campaignName string, ac actor.Actor) (*Enrolment, bool, error) {
	enr, err := api.Store.GetEnrolment(ctx, ac.ID, campaignName)
	if !errors.Is(err, enforcer.ErrNotFound) {
		enr.setStatus()
		return enr, false, err
	}

	camp, err := api.CampaignsAPI.Get(ctx, campaignName)
	if err != nil {
		return nil, false, err
	}

	newEnr, err := api.prepEnrolment(ctx, *camp, ac)
	if err != nil {
		return nil, false, err
	}

	newEnr.StartedAt = time.Now()
	newEnr.EndsAt = camp.EndAt
	if camp.Spec.Deadline > 0 {
		// relative end_date due to deadline (in days)
		newEnr.EndsAt = newEnr.StartedAt.AddDate(0, 0, camp.Spec.Deadline)
	}
	newEnr.setStatus()

	return newEnr, true, api.Store.UpsertEnrolment(ctx, *newEnr)
}

// Ingest processes the action within current enrolments and returns the list of
// enrolments that progressed. If completeMulti is false, only one enrolment will
// be progressed.
func (api *API) Ingest(ctx context.Context, completeMulti bool, act actor.Action) ([]Enrolment, error) {
	if err := act.Validate(); err != nil {
		return nil, err
	}

	applicable, err := api.ListExisting(ctx, act.Actor.ID, []string{StatusActive})
	if err != nil {
		return nil, err
	}
	api.sortApplicable(applicable)

	var res []Enrolment
	var isAffected bool
	var completionErr error
	for _, enr := range applicable {
		isAffected, completionErr = api.applyCompletion(ctx, act, &enr)
		if completionErr != nil {
			break
		} else if isAffected {
			if completionErr = api.Store.UpsertEnrolment(ctx, enr); completionErr != nil {
				break
			}
			res = append(res, enr)
			if !completeMulti {
				break
			}
		}
	}
	return res, completionErr
}

func (api *API) sortApplicable(applicable []Enrolment) {
	// TODO: sort based on priority, end_date etc.
}

func (api *API) prepEnrolment(ctx context.Context, camp campaign.Campaign, ac actor.Actor) (*Enrolment, error) {
	if err := api.checkEligibility(ctx, camp, ac); err != nil {
		return nil, err
	}

	return &Enrolment{
		Status:         StatusEligible,
		ActorID:        ac.ID,
		CampaignID:     camp.Name,
		RemainingSteps: len(camp.Spec.Steps),
	}, nil
}

func (api *API) checkEligibility(ctx context.Context, camp campaign.Campaign, ac actor.Actor) error {
	if camp.Spec.Eligibility == "" {
		return nil
	}

	isPass, err := api.RuleEngine.Exec(ctx, camp.Spec.Eligibility, ruleExecEnv(ac, nil))
	if err != nil {
		return err
	} else if !isPass {
		return enforcer.ErrIneligible
	}
	return nil
}

func (api *API) applyCompletion(ctx context.Context, act actor.Action, enr *Enrolment) (bool, error) {
	camp, err := api.CampaignsAPI.Get(ctx, enr.CampaignID)
	if err != nil {
		return false, err
	}
	env := ruleExecEnv(act.Actor, &act)

	if camp.Spec.IsUnordered {
		done := map[int]struct{}{}
		for _, step := range enr.CompletedSteps {
			done[step.StepID] = struct{}{}
		}
		for i, step := range camp.Spec.Steps {
			if _, alreadyDone := done[i]; alreadyDone {
				continue
			}

			pass, err := api.RuleEngine.Exec(ctx, step, env)
			if err != nil {
				return false, err
			} else if pass {
				enr.CompletedSteps = append(enr.CompletedSteps, StepResult{
					StepID:   i,
					DoneAt:   act.Time,
					ActionID: act.ID,
				})
				enr.RemainingSteps = len(camp.Spec.Steps) - len(enr.CompletedSteps)
				return true, nil
			}
		}

		return false, nil
	}

	nextStepID := len(enr.CompletedSteps)
	if nextStepID >= len(camp.Spec.Steps) {
		return false, enforcer.ErrInternal.WithMsgf("campaign has lesser steps than enrolment")
	}

	pass, err := api.RuleEngine.Exec(ctx, camp.Spec.Steps[nextStepID], env)
	if err != nil || !pass {
		return false, err
	}

	enr.CompletedSteps = append(enr.CompletedSteps, StepResult{
		StepID:   nextStepID,
		DoneAt:   act.Time,
		ActionID: act.ID,
	})
	enr.RemainingSteps = len(camp.Spec.Steps) - len(enr.CompletedSteps)
	return true, nil
}

func ruleExecEnv(ac actor.Actor, act *actor.Action) map[string]interface{} {
	d := map[string]interface{}{}
	if act != nil {
		d["event"] = mergeMap(act.Data, map[string]interface{}{"id": act.ID, "time": act.Time})
		ac = act.Actor
	}
	d["actor"] = mergeMap(ac.Attribs, map[string]interface{}{"id": ac.ID})
	return d
}

func mergeMap(m1, m2 map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range m1 {
		res[k] = v
	}
	for k, v := range m2 {
		res[k] = v
	}
	return res
}

func collectCampaignIDs(existing []Enrolment) []string {
	var res []string
	for _, enrolment := range existing {
		res = append(res, enrolment.CampaignID)
	}
	return res
}

package inmem

import (
	"sync"

	"github.com/spy16/enforcer/core/campaign"
	"github.com/spy16/enforcer/core/enrolment"
)

var _ campaign.Store = (*Store)(nil)
var _ enrolment.Store = (*Store)(nil)

type Store struct {
	mu         sync.RWMutex
	nextID     int
	campaigns  map[int]campaign.Campaign
	enrolments map[string]map[int]enrolment.Enrolment
}

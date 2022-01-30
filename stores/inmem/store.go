package inmem

import (
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

package enrolment

import (
	"strings"

	"github.com/spy16/enforcer"
)

// Actor represents some entity performing an action.
type Actor struct {
	ID      string                 `json:"id"`
	Attribs map[string]interface{} `json:"attribs"`
}

func (a *Actor) Validate() error {
	a.ID = strings.TrimSpace(a.ID)
	if a.ID == "" {
		return enforcer.ErrInvalid.WithMsgf("empty actor_id")
	}
	return nil
}

package actor

import (
	"fmt"
	"strings"

	"github.com/spy16/enforcer"
)

// Actor represents some entity performing an action.
type Actor struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Attribs map[string]interface{} `json:"attribs"`
}

// Validate validates the actor object. ID and Type are mandatory.
func (a *Actor) Validate() error {
	a.ID = strings.TrimSpace(a.ID)
	a.Type = strings.TrimSpace(a.Type)

	if a.ID == "" {
		return enforcer.ErrInvalid.WithMsgf("empty actor_id")
	}

	if a.Type == "" {
		return enforcer.ErrInvalid.WithMsgf("empty type")
	}
	return nil
}

func (a Actor) String() string {
	return fmt.Sprintf("Actor{%s, type='%s'}", a.ID, a.Type)
}

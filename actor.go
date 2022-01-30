package enforcer

import "strings"

// Actor represents some entity performing an action.
type Actor struct {
	ID      string                 `json:"id"`
	Attribs map[string]interface{} `json:"attribs"`
}

func (a *Actor) Validate() error {
	a.ID = strings.TrimSpace(a.ID)
	if a.ID == "" {
		return ErrInvalid.WithMsgf("empty actor_id")
	}
	return nil
}

type Event map[string]interface{}

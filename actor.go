package enforcer

import (
	"fmt"
	"strings"
	"time"
)

// Actor represents some entity performing an action.
type Actor struct {
	ID      string                 `json:"id"`
	Attribs map[string]interface{} `json:"attribs"`
}

// Validate validates the actor object. ID and Type are mandatory.
func (a *Actor) Validate() error {
	a.ID = strings.TrimSpace(a.ID)
	if a.ID == "" {
		return ErrInvalid.WithMsgf("empty actor_id")
	}
	return nil
}

func (a Actor) String() string { return fmt.Sprintf("Actor{id='%s'}", a.ID) }

// Action represents an activity/action executed by an actor.
type Action struct {
	ID      string                 `json:"id"`
	Time    time.Time              `json:"time"`
	Data    map[string]interface{} `json:"data"`
	ActorID string                 `json:"actor_id"`
}

// Validate performs validation of given action.
func (act *Action) Validate() error {
	act.ID = strings.TrimSpace(act.ID)
	act.ActorID = strings.TrimSpace(act.ActorID)
	if act.Time.IsZero() {
		act.Time = time.Now()
	}

	if act.ID == "" {
		return ErrInvalid.WithMsgf("id cannot be empty")
	}

	if act.ActorID == "" {
		return ErrInvalid.WithMsgf("actor_id cannot be empty")
	}

	return nil
}

func (act Action) String() string {
	return fmt.Sprintf("Action{id='%s'}", act.ID)
}

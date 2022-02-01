package actor

import (
	"fmt"
	"strings"
	"time"

	"github.com/spy16/enforcer"
)

// Action represents an activity/action executed by an actor.
type Action struct {
	ID    string                 `json:"id"`
	Time  time.Time              `json:"time"`
	Data  map[string]interface{} `json:"data"`
	Actor Actor                  `json:"actor"`
}

// Validate performs validation of given action.
func (act *Action) Validate() error {
	act.ID = strings.TrimSpace(act.ID)
	if act.Time.IsZero() {
		act.Time = time.Now()
	}

	if act.ID == "" {
		return enforcer.ErrInvalid.WithMsgf("empty action_id")
	}

	return act.Actor.Validate()
}

func (act Action) String() string {
	return fmt.Sprintf("Action{id='%s'}", act.ID)
}

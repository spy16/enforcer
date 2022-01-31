package actor

import (
	"fmt"
	"strings"
	"time"
)

// Action represents an activity/action executed by an actor.
type Action struct {
	ID    string                 `json:"id"`
	Time  time.Time              `json:"time"`
	Actor Actor                  `json:"actor"`
	Data  map[string]interface{} `json:"data"`
}

// Validate performs validation of given action.
func (act *Action) Validate() error {
	act.ID = strings.TrimSpace(act.ID)
	if act.Time.IsZero() {
		act.Time = time.Now()
	}
	return act.Actor.Validate()
}

func (act Action) String() string {
	return fmt.Sprintf("Action{id='%s'}", act.ID)
}

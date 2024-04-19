package states

import (
	"context"
)

type AutomataState interface {
	Run(ctx context.Context) (AutomataState, error)
	String() string
}

// StateFactory provides ability to generate any AutomataState
type StateFactory interface {
	CreateAttempterState() (AutomataState, error)
	CreateFailoverState(error) (AutomataState, error)
	CreateInitState() (AutomataState, error)
	CreateLeaderState() (AutomataState, error)
	CreateStoppingState() (AutomataState, error)
}

package stopping

import (
	"context"
	"log/slog"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
)

func New(logger *slog.Logger, stateFactory states.StateFactory, connManager states.ConnManager) *State {
	logger = logger.With("subsystem", "StoppingState")
	return &State{
		logger:       logger,
		stateFactory: stateFactory,
		connManager:  connManager,
	}
}

type State struct {
	logger       *slog.Logger
	stateFactory states.StateFactory
	connManager  states.ConnManager
}

func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Entered StoppingState, releasing resources")

	// if there's an active connection, close it.
	conn, err := s.connManager.Get()
	if err == nil {
		conn.Close()
	}

	s.logger.LogAttrs(ctx, slog.LevelInfo, "Resources released")

	return nil, nil
}

func (s *State) String() string {
	return "StoppingState"
}

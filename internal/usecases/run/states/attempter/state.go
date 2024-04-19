package attempter

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/go-zookeeper/zk"
)

const NodePath = "/election"

func New(logger *slog.Logger, stateFactory states.StateFactory, connManager states.ConnManager, args cmdargs.RunArgs) *State {
	logger = logger.With("subsystem", "InitState")
	return &State{
		logger:       logger,
		stateFactory: stateFactory,
		args:         args,
		connManager:  connManager,
	}
}

type State struct {
	logger       *slog.Logger
	stateFactory states.StateFactory
	connManager  states.ConnManager
	args         cmdargs.RunArgs
}

func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Entered AttempterState")
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Trying to become a leader...")

	select {
	case <-ctx.Done():
		s.logger.LogAttrs(ctx, slog.LevelInfo, "Ctx cancelled, going to stopping state...")
		return s.stateFactory.CreateStoppingState()
	default:
	}

	conn, err := s.connManager.Get()
	if err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "Failed to get connection")
		return s.stateFactory.CreateFailoverState(err)
	}

	// trying to crate ephemeral node to become leader
	_, err = conn.Create(NodePath, []byte{}, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))

	if errors.Is(err, zk.ErrNodeExists) { // check if leadership is already taken
		s.logger.LogAttrs(ctx, slog.LevelInfo, "Unable to become leader, already exists")
		time.Sleep(s.args.AttempterTimeout)
		return s, nil
	} else if err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "Failed to create ephemeral node")
		return s.stateFactory.CreateFailoverState(err)
	}

	s.logger.LogAttrs(ctx, slog.LevelInfo, "Became leader successfully!")

	return s.stateFactory.CreateLeaderState()
}

func (s *State) String() string {
	return "AttempterState"
}

package init_

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/go-zookeeper/zk"
)

const Timeout = 10

func New(logger *slog.Logger, states states.StateFactory, connManager states.ConnManager, args cmdargs.RunArgs) *State {
	logger = logger.With("subsystem", "InitState")
	return &State{
		logger:       logger,
		stateFactory: states,
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
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Entered InitState")
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Checking availability of resources...")

	select {
	case <-ctx.Done():
		s.logger.LogAttrs(ctx, slog.LevelInfo, "Ctx cancelled, going to stopping state...")
		return s.stateFactory.CreateStoppingState()
	default:
	}

	// validate connection
	conn, _, err := zk.Connect(s.args.ZookeeperServers, Timeout*time.Second)
	if err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "Failed to connect to ZooKeeper server(s)")
		return s.stateFactory.CreateFailoverState(err)
	}
	s.connManager.Set(conn)

	// validate directory to store files
	_, err = os.Stat(s.args.FileDir)
	if err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "Failed to access file")
		return s.stateFactory.CreateFailoverState(err)
	}

	s.logger.LogAttrs(ctx, slog.LevelInfo, "Resources are available, finishing Init state.")

	return s.stateFactory.CreateAttempterState()
}

func (s *State) String() string {
	return "InitState"
}

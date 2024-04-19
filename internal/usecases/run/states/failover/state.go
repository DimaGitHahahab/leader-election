package failover

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/go-zookeeper/zk"
)

const retriesAmount = 5

func New(logger *slog.Logger, stateFactory states.StateFactory, connManager states.ConnManager, err error, args *cmdargs.RunArgs) *State {
	logger = logger.With("subsystem", "InitState")
	return &State{
		logger:       logger,
		err:          err,
		stateFactory: stateFactory,
		connManager:  connManager,
		args:         args,
	}
}

type State struct {
	logger       *slog.Logger
	stateFactory states.StateFactory
	connManager  states.ConnManager
	err          error
	args         *cmdargs.RunArgs
}

func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Attempting to recover from error...")

	timeout := 1
	for attempt := 1; attempt <= retriesAmount; attempt++ {
		select {
		case <-ctx.Done():
			s.logger.LogAttrs(ctx, slog.LevelInfo, "Ctx cancelled, going to stopping state...")
			return s.stateFactory.CreateStoppingState()
		default:
		}

		s.logger.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf("Attempt #%d", attempt))
		newConn, _, err := zk.Connect(s.args.ZookeeperServers, time.Duration(timeout)*time.Second)
		if err != nil {
			s.logger.LogAttrs(ctx, slog.LevelError, "Failed to reconnect to ZooKeeper server(s)")
			timeout *= 2
			continue // connection is not established, try again
		}

		// connection established
		s.connManager.Set(newConn)

		// validate directory
		_, err = os.Stat(s.args.FileDir)
		if err != nil {
			s.logger.LogAttrs(ctx, slog.LevelError, "Failed to access file")
			return s.stateFactory.CreateStoppingState()
		}

		// repaired successfully
		return s.stateFactory.CreateAttempterState()
	}

	s.logger.LogAttrs(ctx, slog.LevelError, "Failed to reconnect to ZooKeeper server(s) after all attempts")
	return s.stateFactory.CreateStoppingState()
}

func (s *State) String() string {
	return "FailoverState"
}

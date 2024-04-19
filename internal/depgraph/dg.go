package depgraph

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/attempter"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/failover"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/init_"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/leader"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/stopping"
	"github.com/go-zookeeper/zk"
)

// DepGraph manages dependencies, implements states.ConnManager, states.StateFactory
type DepGraph struct {
	logger *dgEntity[*slog.Logger]

	conn *dgEntity[*zk.Conn] // connection with zk, used to implement states.ConnManager

	stateRunner *dgEntity[*run.LoopRunner]

	args *cmdargs.RunArgs // config

	// states that are used to implement states.StateFactory
	attempterState *dgEntity[*attempter.State]
	failoverState  *dgEntity[*failover.State]
	InitState      *dgEntity[*init_.State]
	LeaderState    *dgEntity[*leader.State]
	StoppingState  *dgEntity[*stopping.State]
}

func New(args *cmdargs.RunArgs) *DepGraph {
	return &DepGraph{
		logger:      &dgEntity[*slog.Logger]{},
		conn:        &dgEntity[*zk.Conn]{},
		stateRunner: &dgEntity[*run.LoopRunner]{},

		args: args,

		attempterState: &dgEntity[*attempter.State]{},
		failoverState:  &dgEntity[*failover.State]{},
		InitState:      &dgEntity[*init_.State]{},
		LeaderState:    &dgEntity[*leader.State]{},
		StoppingState:  &dgEntity[*stopping.State]{},
	}
}

func (dg *DepGraph) GetLogger() (*slog.Logger, error) {
	return dg.logger.get(func() (*slog.Logger, error) {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})), nil
	})
}

func (dg *DepGraph) GetRunner() (run.Runner, error) {
	return dg.stateRunner.get(func() (*run.LoopRunner, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("get logger: %w", err)
		}
		return run.NewLoopRunner(logger), nil
	})
}

func (dg *DepGraph) CreateAttempterState() (states.AutomataState, error) {
	return dg.attempterState.get(func() (*attempter.State, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("get logger: %w", err)
		}

		return attempter.New(logger, dg, dg, *dg.args), nil
	})
}

func (dg *DepGraph) CreateFailoverState(err error) (states.AutomataState, error) {
	return dg.failoverState.get(func() (*failover.State, error) {
		logger, loggerErr := dg.GetLogger()
		if loggerErr != nil {
			return nil, fmt.Errorf("get logger: %w", loggerErr)
		}

		return failover.New(logger, dg, dg, err, dg.args), nil
	})
}

func (dg *DepGraph) CreateInitState() (states.AutomataState, error) {
	return dg.InitState.get(func() (*init_.State, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("get logger: %w", err)
		}

		return init_.New(logger, dg, dg, *dg.args), nil
	})
}

func (dg *DepGraph) CreateLeaderState() (states.AutomataState, error) {
	return dg.LeaderState.get(func() (*leader.State, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("get logger: %w", err)
		}

		return leader.New(logger, dg, *dg.args), nil
	})
}

func (dg *DepGraph) CreateStoppingState() (states.AutomataState, error) {
	return dg.StoppingState.get(func() (*stopping.State, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("get logger: %w", err)
		}

		return stopping.New(logger, dg, dg), nil
	})
}

func (dg *DepGraph) Set(conn *zk.Conn) {
	dg.conn.Do(func() {
		dg.conn.value = conn
	})
}

func (dg *DepGraph) Get() (*zk.Conn, error) {
	return dg.conn.get(func() (*zk.Conn, error) {
		if dg.conn.value == nil {
			return nil, fmt.Errorf("connection not set")
		}
		return dg.conn.value, nil
	})
}

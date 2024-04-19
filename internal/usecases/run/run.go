package run

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
)

var _ Runner = &LoopRunner{}

type Runner interface {
	Run(ctx context.Context, state states.AutomataState) error
}

func NewLoopRunner(logger *slog.Logger) *LoopRunner {
	logger = logger.With("subsystem", "StateRunner")
	return &LoopRunner{
		logger: logger,
	}
}

type LoopRunner struct {
	logger *slog.Logger
}

func (r *LoopRunner) Run(ctx context.Context, state states.AutomataState) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		stop()
	}()

	for state != nil {
		r.logger.LogAttrs(ctx, slog.LevelInfo, "start running state", slog.String("state", state.String()))
		var err error
		state, err = state.Run(ctx)
		if err != nil {
			r.logger.LogAttrs(ctx, slog.LevelError, fmt.Sprintf("state %s run: %v", state.String(), err))
		}
	}

	r.logger.LogAttrs(ctx, slog.LevelInfo, "no new state, finish")
	return nil
}

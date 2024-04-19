package leader

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
)

const FileNamePrefix = "leader_file_"

func New(logger *slog.Logger, stateFactory states.StateFactory, args cmdargs.RunArgs) *State {
	logger = logger.With("subsystem", "InitState")
	return &State{
		logger:       logger,
		stateFactory: stateFactory,
		args:         args,
	}
}

type State struct {
	logger       *slog.Logger
	stateFactory states.StateFactory
	args         cmdargs.RunArgs
}

func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Entered LeaderState")
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Writing file to disk...")

	select {
	case <-ctx.Done():
		s.logger.LogAttrs(ctx, slog.LevelInfo, "Ctx cancelled, going to stopping state...")
		return s.stateFactory.CreateStoppingState()
	default:
	}

	fileName := getFileName()

	// writing empty file
	err := os.WriteFile(filepath.Join(s.args.FileDir, fileName), []byte{}, 0o644)
	if err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "Failed to write file")
		return s.stateFactory.CreateFailoverState(err)
	}
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Successful writing!")

	files, err := os.ReadDir(s.args.FileDir)
	if err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "Failed to read directory")
		return s.stateFactory.CreateFailoverState(err)
	}

	// if number of files exceeds capacity, remove the oldest ones
	if len(files) > s.args.StorageCapacity {

		// sort by date of creation
		sort.Slice(files, func(i, j int) bool {
			iInf, _ := files[i].Info()
			jInf, _ := files[j].Info()
			return iInf.ModTime().Before(jInf.ModTime())
		})

		// remove
		for i := 0; i < len(files)-s.args.StorageCapacity; i++ {
			err = os.Remove(filepath.Join(s.args.FileDir, files[i].Name()))
			if err != nil {
				s.logger.LogAttrs(ctx, slog.LevelError, "Failed to delete old file")
				return s.stateFactory.CreateFailoverState(err)
			}
		}
	}

	time.Sleep(s.args.LeaderTimeout)

	return s, nil
}

func (s *State) String() string {
	return "LeaderState"
}

func getFileName() string {
	// file name consists of common prefix and unique suffix so there will be no duplicate files
	return FileNamePrefix + time.Now().Format(time.RFC3339)
}

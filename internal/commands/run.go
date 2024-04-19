package commands

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/depgraph"
	"github.com/spf13/cobra"
)

func InitRunCommand() (cobra.Command, error) {
	cmdArgs := cmdargs.RunArgs{}
	cmd := cobra.Command{
		Use:   "run",
		Short: "Starts a leader election node",
		Long: `This command starts the leader election node that connects to zookeeper
		and starts to try to acquire leadership by creation of ephemeral node`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dg := depgraph.New(&cmdArgs)
			logger, err := dg.GetLogger()
			if err != nil {
				return fmt.Errorf("get logger: %w", err)
			}
			logger.Info("args received", slog.String("servers", strings.Join(cmdArgs.ZookeeperServers, ", ")))

			runner, err := dg.GetRunner()
			if err != nil {
				return fmt.Errorf("get runner: %w", err)
			}
			firstState, err := dg.CreateInitState()
			if err != nil {
				return fmt.Errorf("create init state: %w", err)
			}
			err = runner.Run(cmd.Context(), firstState)
			if err != nil {
				return fmt.Errorf("run states: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&(cmdArgs.ZookeeperServers), "zk-servers", "s", []string{}, "Set the zookeeper servers.")
	cmd.Flags().DurationVarP(&(cmdArgs.LeaderTimeout), "leader-timeout", "l", 0, "Periodicity of the leader's file writing to disk.")
	cmd.Flags().DurationVarP(&(cmdArgs.AttempterTimeout), "attempter-timeout", "a", 0, "Periodicity with which the attempter tries to become a leader.")
	cmd.Flags().StringVarP(&(cmdArgs.FileDir), "file-dir", "f", "", "Directory where the leader should write files.")
	cmd.Flags().IntVarP(&(cmdArgs.StorageCapacity), "storage-capacity", "c", 0, "Maximum number of files in the file-dir directory.")

	return cmd, nil
}

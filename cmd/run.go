package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/grafana/k8s-manifest-tail/internal/logging"
	"github.com/grafana/k8s-manifest-tail/internal/telemetry"
	"github.com/grafana/k8s-manifest-tail/pkg"
	"github.com/spf13/cobra"
	"time"
)

var runAndWatchCmd = &cobra.Command{
	Use:     "run",
	Short:   "Fetch manifests and keep watching for changes",
	PreRunE: LoadConfiguration,
	RunE:    runRun,
}

func init() {
	rootCmd.AddCommand(runAndWatchCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	clients, err := GetKubeProvider().Provide(Configuration)
	if err != nil {
		return fmt.Errorf("create kubernetes clients: %w", err)
	}

	logger, shutdownTelemetry, err := telemetry.SetupLogging(ctx, Configuration.Logging)
	if err != nil {
		return fmt.Errorf("configure telemetry logging: %w", err)
	}
	defer func() { _ = shutdownTelemetry(context.Background()) }()
	diffLogger := logging.NewDiffLogger(Configuration.Logging, logger)

	tail := pkg.Tail{
		Clients:    clients,
		Config:     Configuration,
		DiffLogger: diffLogger,
		Processor:  GetManifestProcessor(),
	}
	total, err := tail.RunFullManifestCheck(ctx)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Fetched %d manifest(s)\n", total)

	refreshErrCh := make(chan error, 1)
	refreshInterval, _ := Configuration.GetRefreshInterval() // Error checked during config validation
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				total, runErr := tail.RunFullManifestCheck(ctx)
				if runErr != nil {
					if !errors.Is(runErr, context.Canceled) {
						select {
						case refreshErrCh <- runErr:
						default:
						}
					}
					cancel()
					return
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Fetched %d manifest(s)\n", total)
			}
		}
	}()

	err = tail.WatchResources(ctx)
	cancel()
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	select {
	case runErr := <-refreshErrCh:
		if runErr != nil && !errors.Is(runErr, context.Canceled) {
			return runErr
		}
	default:
	}
	return nil
}

package cmd

import (
	"context"
	"fmt"
	"github.com/grafana/k8s-manifest-tail/pkg"
	"time"

	"github.com/spf13/cobra"

	"github.com/grafana/k8s-manifest-tail/internal/logging"
	"github.com/grafana/k8s-manifest-tail/internal/telemetry"
)

var runCmd = &cobra.Command{
	Use:     "run",
	Short:   "Fetch manifests once and exit",
	PreRunE: LoadConfiguration,
	RunE:    runRun,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
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
	return nil
}

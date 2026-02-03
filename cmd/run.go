package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/grafana/k8s-manifest-tail/internal/discovery"
	"github.com/grafana/k8s-manifest-tail/internal/logging"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
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

	fetcher := discovery.NewFetcher(clients, Configuration)
	if manifestProcessor == nil {
		SetManifestProcessor(manifest.NewProcessor(Configuration))
	}

	logger, shutdownTelemetry, err := telemetry.SetupLogging(ctx, Configuration.Logging)
	if err != nil {
		return fmt.Errorf("configure telemetry logging: %w", err)
	}
	defer func() { _ = shutdownTelemetry(context.Background()) }()
	diffLogger := logging.NewDiffLogger(Configuration.Logging, logger)

	var total int
	for _, rule := range Configuration.Objects {
		objects, err := fetcher.FetchResources(ctx, rule)
		if err != nil {
			return err
		}
		for i := range objects {
			obj := objects[i].DeepCopy()
			total++
			diff, err := manifestProcessor.Process(rule, obj, Configuration)
			if err != nil {
				return fmt.Errorf("process %s %s/%s: %w", rule.Kind, obj.GetNamespace(), obj.GetName(), err)
			}
			diffLogger.Log(diff)
		}
	}

	//telemetry.Info(logger, fmt.Sprintf("Fetched %d manifest(s)", total))
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Fetched %d manifest(s)\n", total)
	return nil
}

var manifestProcessor manifest.Processor

// SetManifestProcessor overrides the manifest processor used by the run command (primarily for tests).
func SetManifestProcessor(p manifest.Processor) {
	manifestProcessor = p
}

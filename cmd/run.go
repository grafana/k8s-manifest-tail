package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/discovery"
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
	var total int
	for _, rule := range Configuration.Objects {
		objects, err := fetcher.FetchResources(ctx, rule)
		if err != nil {
			return err
		}
		for i := range objects {
			obj := objects[i].DeepCopy()
			total++
			if err := manifestProcessor(rule, obj, Configuration); err != nil {
				return fmt.Errorf("process %s %s/%s: %w", rule.Kind, obj.GetNamespace(), obj.GetName(), err)
			}
		}
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Fetched %d manifest(s)\n", total)
	return nil
}

type manifestProcessorFunc func(rule config.ObjectRule, obj *unstructured.Unstructured, cfg *config.Config) error

var manifestProcessor manifestProcessorFunc = func(config.ObjectRule, *unstructured.Unstructured, *config.Config) error {
	return nil
}

// SetManifestProcessor overrides the manifest processor used by the run command (primarily for tests).
// Passing nil restores the default no-op processor.
func SetManifestProcessor(fn manifestProcessorFunc) {
	if fn == nil {
		manifestProcessor = func(config.ObjectRule, *unstructured.Unstructured, *config.Config) error {
			return nil
		}
		return
	}
	manifestProcessor = fn
}

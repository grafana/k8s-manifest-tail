package cmd

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/discovery"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List the Kubernetes objects matched by the configuration",
	PreRunE: LoadConfiguration,
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()

	clients, err := GetKubeProvider().Provide(Configuration)
	if err != nil {
		return fmt.Errorf("create kubernetes clients: %w", err)
	}

	fetcher := discovery.NewFetcher(clients, Configuration)
	printer := &listPrinter{out: cmd.OutOrStdout()}
	for _, rule := range Configuration.Objects {
		items, err := fetcher.FetchResources(ctx, rule)
		if err != nil {
			return err
		}
		for _, item := range items {
			printer.Print(rule, item.GetNamespace(), item.GetName())
		}
	}
	printer.Flush()
	return nil
}

type listPrinter struct {
	out           io.Writer
	headerPrinted bool
	itemsPrinted  bool
}

func (p *listPrinter) Print(rule config.ObjectRule, namespace, name string) {
	if !p.headerPrinted {
		_, _ = fmt.Fprintf(p.out, "%-15s %-15s %-20s %s\n", "API VERSION", "KIND", "NAMESPACE", "NAME")
		p.headerPrinted = true
	}
	_, _ = fmt.Fprintf(p.out, "%-15s %-15s %-20s %s\n", rule.APIVersion, rule.Kind, namespaceOrDash(namespace), name)
	p.itemsPrinted = true
}

func (p *listPrinter) Flush() {
	if !p.itemsPrinted {
		_, _ = fmt.Fprintln(p.out, "No resources found")
	}
}

func namespaceOrDash(ns string) string {
	if ns == "" {
		return "-"
	}
	return ns
}

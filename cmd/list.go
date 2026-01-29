package cmd

import (
	"context"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/kube"
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

	printer := &listPrinter{out: cmd.OutOrStdout()}
	for _, rule := range Configuration.Objects {
		if err := listResourcesForRule(ctx, clients, rule, Configuration, printer); err != nil {
			return err
		}
	}
	printer.Flush()
	return nil
}

func listResourcesForRule(ctx context.Context, clients *kube.Clients, rule config.ObjectRule, cfg *config.Config, printer *listPrinter) error {
	mapping, err := resolveMapping(clients.Mapper, rule)
	if err != nil {
		return err
	}

	namespaces := effectiveNamespaces(rule, cfg)
	excludeSet := toSet(cfg.ExcludeNamespaces)
	resourceClient := clients.Dynamic.Resource(mapping.Resource)

	switch mapping.Scope.Name() {
	case meta.RESTScopeNameNamespace:
		if len(namespaces) == 0 {
			list, err := resourceClient.Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
			if err != nil {
				return fmt.Errorf("list %s across namespaces: %w", rule.Kind, err)
			}
			for _, item := range list.Items {
				ns := item.GetNamespace()
				if shouldExcludeNamespace(ns, excludeSet) {
					continue
				}
				printer.Print(rule, ns, item.GetName())
			}
			return nil
		}

		for _, ns := range namespaces {
			if shouldExcludeNamespace(ns, excludeSet) {
				continue
			}
			list, err := resourceClient.Namespace(ns).List(ctx, metav1.ListOptions{})
			if err != nil {
				return fmt.Errorf("list %s in namespace %s: %w", rule.Kind, ns, err)
			}
			for _, item := range list.Items {
				printer.Print(rule, ns, item.GetName())
			}
		}
	default:
		list, err := resourceClient.Namespace("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("list %s (cluster-scoped): %w", rule.Kind, err)
		}
		for _, item := range list.Items {
			printer.Print(rule, item.GetNamespace(), item.GetName())
		}
	}

	return nil
}

func resolveMapping(mapper meta.RESTMapper, rule config.ObjectRule) (*meta.RESTMapping, error) {
	gv, err := schema.ParseGroupVersion(rule.APIVersion)
	if err != nil {
		return nil, fmt.Errorf("parse apiVersion %q: %w", rule.APIVersion, err)
	}

	gk := schema.GroupKind{
		Group: gv.Group,
		Kind:  rule.Kind,
	}

	mapping, err := mapper.RESTMapping(gk, gv.Version)
	if err != nil {
		return nil, fmt.Errorf("resolve resource for %s (%s): %w", rule.Kind, rule.APIVersion, err)
	}
	return mapping, nil
}

func effectiveNamespaces(rule config.ObjectRule, cfg *config.Config) []string {
	if len(rule.Namespaces) > 0 {
		return cloneAndDedupe(rule.Namespaces)
	}
	return cloneAndDedupe(cfg.Namespaces)
}

func cloneAndDedupe(input []string) []string {
	if len(input) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(input))
	var result []string
	for _, ns := range input {
		if ns == "" {
			continue
		}
		if _, ok := seen[ns]; ok {
			continue
		}
		seen[ns] = struct{}{}
		result = append(result, ns)
	}
	slices.Sort(result)
	return result
}

func toSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		set[v] = struct{}{}
	}
	return set
}

func shouldExcludeNamespace(ns string, exclude map[string]struct{}) bool {
	if ns == "" {
		return false
	}
	_, ok := exclude[ns]
	return ok
}

type listPrinter struct {
	out           io.Writer
	headerPrinted bool
	itemsPrinted  bool
}

func (p *listPrinter) Print(rule config.ObjectRule, namespace, name string) {
	if !p.headerPrinted {
		fmt.Fprintf(p.out, "%-15s %-15s %-20s %s\n", "KIND", "API VERSION", "NAMESPACE", "NAME")
		p.headerPrinted = true
	}
	fmt.Fprintf(p.out, "%-15s %-15s %-20s %s\n", rule.Kind, rule.APIVersion, namespaceOrDash(namespace), name)
	p.itemsPrinted = true
}

func (p *listPrinter) Flush() {
	if !p.itemsPrinted {
		fmt.Fprintln(p.out, "No resources found")
	}
}

func namespaceOrDash(ns string) string {
	if ns == "" {
		return "-"
	}
	return ns
}

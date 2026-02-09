package kube

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	memdiscovery "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/grafana/k8s-manifest-tail/internal/config"
)

// Clients bundles the interfaces needed to query Kubernetes APIs.
type Clients struct {
	Dynamic dynamic.Interface
	Mapper  meta.RESTMapper
}

// Provider creates Kubernetes API clients.
type Provider interface {
	Provide(cfg *config.Config) (*Clients, error)
}

type defaultProvider struct{}

// NewProvider returns the default Kubernetes client provider.
func NewProvider() Provider {
	return &defaultProvider{}
}

func (p *defaultProvider) Provide(cfg *config.Config) (*Clients, error) {
	restCfg, err := buildRestConfig(cfg.KubeconfigPath)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("create dynamic client: %w", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("create discovery client: %w", err)
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memdiscovery.NewMemCacheClient(discoveryClient))

	return &Clients{
		Dynamic: dynamicClient,
		Mapper:  mapper,
	}, nil
}

func buildRestConfig(explicitPath string) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if explicitPath != "" {
		loadingRules.ExplicitPath = explicitPath
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("build rest config: %w", err)
	}
	return restConfig, nil
}

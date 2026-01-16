package cmd

import "github.com/grafana/k8s-manifest-tail/internal/kube"

var clientProvider kube.Provider = kube.NewProvider()

// GetKubeProvider returns the current Kubernetes provider (used internally).
func GetKubeProvider() kube.Provider {
	return clientProvider
}

// SetKubeProvider overrides the Kubernetes client provider (useful for tests).
// Passing nil restores the default provider.
func SetKubeProvider(p kube.Provider) {
	if p == nil {
		clientProvider = kube.NewProvider()
		return
	}
	clientProvider = p
}

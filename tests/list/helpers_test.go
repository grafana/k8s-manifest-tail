package list_test

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/grafana/k8s-manifest-tail/cmd"
	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/kube"

	. "github.com/onsi/ginkgo/v2"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var testScheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(testScheme)
}

type resourceMapping struct {
	GVR   schema.GroupVersionResource
	GVK   schema.GroupVersionKind
	Scope meta.RESTScope
}

type staticProvider struct {
	clients *kube.Clients
}

func (s *staticProvider) Provide(_ *config.Config) (*kube.Clients, error) {
	return s.clients, nil
}

func newFakeProvider(objects []runtime.Object, mappings []resourceMapping) kube.Provider {
	dynamicClient := fake.NewSimpleDynamicClient(testScheme, objects...)
	mapper := newRESTMapper(mappings)
	return &staticProvider{
		clients: &kube.Clients{
			Dynamic: dynamicClient,
			Mapper:  mapper,
		},
	}
}

func newRESTMapper(mappings []resourceMapping) meta.RESTMapper {
	groupVersions := make(map[schema.GroupVersion]struct{})
	for _, m := range mappings {
		gv := schema.GroupVersion{Group: m.GVR.Group, Version: m.GVR.Version}
		groupVersions[gv] = struct{}{}
	}
	var gvList []schema.GroupVersion
	for gv := range groupVersions {
		gvList = append(gvList, gv)
	}
	mapper := meta.NewDefaultRESTMapper(gvList)
	for _, m := range mappings {
		mapper.AddSpecific(m.GVK, m.GVR, m.GVR, m.Scope)
	}
	return mapper
}

func runListCommand(configPath string, args ...string) (string, string, error) {
	allArgs := append([]string{"list", "--config", configPath}, args...)
	var stdout, stderr bytes.Buffer
	err := cmd.ExecuteWithArgs(allArgs, &stdout, &stderr)
	cmd.ResetConfiguration()
	return stdout.String(), stderr.String(), err
}

func writeConfigFile(t GinkgoTInterface, contents string) string {
	t.Helper()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

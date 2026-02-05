package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	g := gomega.NewWithT(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
kubeconfig: "/tmp/kubeconfig"
output:
  directory: out
  format: json
logging:
  logDiffs: compact
objects:
  - apiVersion: v1
    kind: Pod
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(cfg.Kubeconfig).To(gomega.Equal("/tmp/kubeconfig"))
	g.Expect(cfg.Output.Directory).To(gomega.Equal("out"))
	g.Expect(cfg.Output.Format).To(gomega.Equal(OutputFormatJSON))
	g.Expect(cfg.Logging.Mode()).To(gomega.Equal(LogDiffsCompact))
	g.Expect(cfg.Objects).To(gomega.HaveLen(1))
	g.Expect(cfg.Objects[0].Kind).To(gomega.Equal("Pod"))
}

func TestLoadMissingFile(t *testing.T) {
	t.Parallel()

	g := gomega.NewWithT(t)

	_, err := Load("missing.yaml")
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.Equal("open config: open missing.yaml: no such file or directory"))
}

func TestObjectRuleValidateDuplicateNamespaces(t *testing.T) {
	t.Parallel()

	g := gomega.NewWithT(t)

	rule := ObjectRule{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Namespaces: []string{"default", "prod", "default"},
	}

	err := rule.Validate()
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("duplicate namespace"))
}

func TestConfigValidateInvokesRuleValidation(t *testing.T) {
	t.Parallel()

	g := gomega.NewWithT(t)

	cfg := &Config{
		Objects: []ObjectRule{
			{APIVersion: "v1", Kind: "Pod", Namespaces: []string{"default"}},
			{APIVersion: "apps/v1", Kind: "Deployment", Namespaces: []string{"prod", "stage"}},
		},
	}

	err := cfg.Validate()
	g.Expect(err).NotTo(gomega.HaveOccurred())
}

func TestApplyEnvOverrides(t *testing.T) {
	g := gomega.NewWithT(t)
	cfg := &Config{}
	t.Setenv("K8S_MANIFEST_TAIL_OUTPUT_DIRECTORY", "from-env")
	t.Setenv("K8S_MANIFEST_TAIL_OUTPUT_FORMAT", "JSON")
	t.Setenv("K8S_MANIFEST_TAIL_LOGGING_LOG_DIFFS", "compact")
	t.Setenv("K8S_MANIFEST_TAIL_REFRESH_INTERVAL", "2h")
	t.Setenv("K8S_MANIFEST_TAIL_NAMESPACES", "default,prod")
	t.Setenv("K8S_MANIFEST_TAIL_EXCLUDE_NAMESPACES", "kube-system")

	ApplyEnvOverrides(cfg)

	g.Expect(cfg.Output.Directory).To(gomega.Equal("from-env"))
	g.Expect(cfg.Output.Format).To(gomega.Equal(OutputFormatJSON))
	g.Expect(cfg.Logging.LogDiffs).To(gomega.Equal(LogDiffsCompact))
	g.Expect(cfg.RefreshInterval).To(gomega.Equal("2h"))
	g.Expect(cfg.Namespaces).To(gomega.Equal([]string{"default", "prod"}))
	g.Expect(cfg.ExcludeNamespaces).To(gomega.Equal([]string{"kube-system"}))
}

func TestLoggingConfigUnmarshalBool(t *testing.T) {
	t.Parallel()

	g := gomega.NewWithT(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := `
logging:
  logDiffs: false
objects:
  - apiVersion: v1
    kind: Pod
`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(cfg.Logging.Mode()).To(gomega.Equal(LogDiffsDisabled))
}

func TestLoggingConfigInvalidMode(t *testing.T) {
	t.Parallel()

	g := gomega.NewWithT(t)

	cfg := &Config{
		Logging: LoggingConfig{LogDiffs: LogDiffMode("unknown")},
		Objects: []ObjectRule{
			{APIVersion: "v1", Kind: "Pod"},
		},
	}

	err := cfg.Validate()
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("validate logging config"))
}

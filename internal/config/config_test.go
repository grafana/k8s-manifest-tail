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

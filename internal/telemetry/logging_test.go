package telemetry

import (
	"context"
	"testing"

	"github.com/onsi/gomega"

	"github.com/grafana/k8s-manifest-tail/internal/config"
)

func TestSetupLoggingCreatesLogger(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	logger, shutdown, err := SetupLogging(context.Background(), config.LoggingConfig{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(logger).NotTo(gomega.BeNil())
	g.Expect(shutdown).NotTo(gomega.BeNil())
}

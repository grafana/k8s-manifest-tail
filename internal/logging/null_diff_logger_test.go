package logging_test

import (
	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/logging"
	"testing"

	"github.com/onsi/gomega"
)

func TestNullDiffLoggerDoesNothing(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	stub := &StubLogger{}
	logger := logging.NewDiffLogger(config.LoggingConfig{LogDiffs: config.LogDiffsDisabled}, stub)
	logger.Log(config.ObjectRule{}, nil, nil)
	g.Expect(stub.records).To(gomega.HaveLen(0))
}

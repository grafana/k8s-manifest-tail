package logging_test

import (
	"context"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/embedded"
)

type StubLogger struct {
	embedded.Logger
	records []log.Record
}

func (s *StubLogger) Emit(_ context.Context, rec log.Record) {
	s.records = append(s.records, rec)
}

func (s *StubLogger) Enabled(context.Context, log.EnabledParameters) bool {
	return true
}

package describe_test

import (
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	projectRoot = func() string {
		_, file, _, ok := runtime.Caller(0)
		if !ok {
			panic("unable to determine caller information")
		}
		return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	}()
	fixturesDir = filepath.Join(projectRoot, "tests", "describe", "fixtures")
)

func TestDescribe(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Describe Suite")
}

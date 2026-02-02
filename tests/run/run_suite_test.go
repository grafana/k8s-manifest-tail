package run_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRunCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Run Command Suite")
}

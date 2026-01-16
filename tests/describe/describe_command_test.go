package describe_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Describe command", func() {
	It("exposes help text", func() {
		result := runCLI("--help")
		Expect(result.Error).NotTo(HaveOccurred(), result.Stdout+"\n"+result.Stderr)
		Expect(result.Stdout).To(ContainSubstring("k8s-manifest-tail"))
		Expect(result.Stdout).To(ContainSubstring("describe"))
	})

	DescribeTable("prints the configured rules",
		func(fixture string, expectedSnippets ...string) {
			result := runCLI("describe", "--config", fixturePath(fixture))
			Expect(result.Error).NotTo(HaveOccurred(), fmt.Sprintf("stderr: %s", result.Stderr))
			Expect(result.Stdout).To(ContainSubstring("This configuration will get manifests for:"))
			for _, snippet := range expectedSnippets {
				Expect(result.Stdout).To(ContainSubstring(snippet))
			}
		},
		Entry("all namespaces default", "basic.yaml",
			"Pods in all namespaces",
		),
		Entry("global namespaces apply to all rules", "global_namespaces.yaml",
			`Pods in the "default", "prod", or "staging" namespaces`,
			`Deployments in the "default", "prod", or "staging" namespaces`,
		),
		Entry("object overrides limit scope", "object_overrides.yaml",
			`Deployments in the "default" namespace`,
			`Services in the "prod" or "staging" namespaces`,
		),
		Entry("excluded namespaces documented", "exclude_namespaces.yaml",
			`Pods in all namespaces except "kube-system" or "observability"`,
		),
	)
})

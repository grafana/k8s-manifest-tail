package manifest

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestRedactEnvValuesFilter(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	pod := newUnstructured("v1", "Pod", "default", "api")
	pod.Object["spec"] = map[string]interface{}{
		"containers": []interface{}{
			map[string]interface{}{
				"name": "app",
				"env": []interface{}{
					map[string]interface{}{"name": "FOO", "value": "bar"},
					map[string]interface{}{"name": "SECRET", "valueFrom": map[string]interface{}{"secretKeyRef": map[string]interface{}{"name": "sec"}}},
				},
			},
		},
	}

	filter := RedactEnvValuesFilter{}
	err := filter.Apply(pod)

	g.Expect(err).NotTo(gomega.HaveOccurred())
	env := pod.Object["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["env"].([]interface{})
	first := env[0].(map[string]interface{})
	second := env[1].(map[string]interface{})
	g.Expect(first["value"]).To(gomega.Equal(redactedValue))
	g.Expect(second).To(gomega.HaveKey("valueFrom"))
}

func TestRedactEnvValuesFilterHandlesCronJob(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cronJob := newUnstructured("batch/v1", "CronJob", "", "nightly")
	cronJob.Object["spec"] = map[string]interface{}{
		"jobTemplate": map[string]interface{}{
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name": "worker",
								"env": []interface{}{
									map[string]interface{}{"name": "TOKEN", "value": "shhh"},
								},
							},
						},
					},
				},
			},
		},
	}

	filter := RedactEnvValuesFilter{}
	err := filter.Apply(cronJob)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	env := cronJob.Object["spec"].(map[string]interface{})["jobTemplate"].(map[string]interface{})["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["env"].([]interface{})
	g.Expect(env[0].(map[string]interface{})["value"]).To(gomega.Equal(redactedValue))
}

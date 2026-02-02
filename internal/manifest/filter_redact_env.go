package manifest

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// RedactEnvValuesFilter masks literal environment variable values inside Pod specs.
type RedactEnvValuesFilter struct{}

var redactedValue = "<Redacted>"

// Apply redacts env[].value fields in Pod specs and well-known workload templates.
func (RedactEnvValuesFilter) Apply(obj *unstructured.Unstructured) error {
	if obj == nil {
		return nil
	}

	var specPath []string
	switch obj.GetKind() {
	case "Pod":
		specPath = []string{"spec"}
	case "Deployment", "ReplicaSet", "StatefulSet", "DaemonSet":
		specPath = []string{"spec", "template", "spec"}
	case "Job":
		specPath = []string{"spec", "template", "spec"}
	case "CronJob":
		specPath = []string{"spec", "jobTemplate", "spec", "template", "spec"}
	default:
		return nil
	}

	return redactEnvAt(obj.Object, specPath...)
}

func redactEnvAt(obj map[string]interface{}, path ...string) error {
	spec, found, err := unstructured.NestedMap(obj, path...)
	if err != nil || !found {
		return err
	}

	anyChanged := false
	for _, field := range []string{"containers", "initContainers"} {
		raw, ok := spec[field]
		if !ok || raw == nil {
			continue
		}
		containerSlice, ok := raw.([]interface{})
		if !ok {
			continue
		}
		changed := false
		for i, c := range containerSlice {
			containerMap, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			envRaw, exists := containerMap["env"]
			if !exists || envRaw == nil {
				continue
			}
			envSlice, ok := envRaw.([]interface{})
			if !ok {
				continue
			}
			envChanged := false
			for j, env := range envSlice {
				envMap, ok := env.(map[string]interface{})
				if !ok {
					continue
				}
				if _, hasValueFrom := envMap["valueFrom"]; hasValueFrom {
					continue
				}
				if _, ok := envMap["value"]; ok {
					envMap["value"] = redactedValue
					envSlice[j] = envMap
					envChanged = true
				}
			}
			if envChanged {
				containerMap["env"] = envSlice
				containerSlice[i] = containerMap
				changed = true
			}
		}
		if changed {
			spec[field] = containerSlice
			anyChanged = true
		}
	}

	if anyChanged {
		return unstructured.SetNestedMap(obj, spec, path...)
	}
	return nil
}

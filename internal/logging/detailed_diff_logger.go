package logging

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/grafana/k8s-manifest-tail/internal/telemetry"

	"github.com/grafana/k8s-manifest-tail/internal/manifest"
	"go.opentelemetry.io/otel/log"
	"k8s.io/apimachinery/pkg/runtime"
)

// DetailedDiffLogger prints a JSON diff statement.
type DetailedDiffLogger struct {
	logger log.Logger
}

type DetailedDiffReport struct {
	Kind      string      `json:"kind"`
	Name      string      `json:"name"`
	Namespace string      `json:"namespace,omitempty"`
	Action    string      `json:"action"`
	Previous  interface{} `json:"previous,omitempty"`
	Current   interface{} `json:"current,omitempty"`
}

// Log implements DiffLogger.
func (l *DetailedDiffLogger) Log(diff *manifest.Diff) {
	if l.logger == nil || diff == nil || (diff.Previous == nil && diff.Current == nil) {
		return
	}

	payload := &DetailedDiffReport{}
	if diff.Previous == nil {
		payload.Kind = diff.Current.GetKind()
		payload.Name = diff.Current.GetName()
		payload.Namespace = diff.Current.GetNamespace()
		payload.Action = "created"
	} else if diff.Current == nil {
		payload.Kind = diff.Previous.GetKind()
		payload.Name = diff.Previous.GetName()
		payload.Namespace = diff.Previous.GetNamespace()
		payload.Action = "deleted"
	} else {
		payload.Kind = diff.Current.GetKind()
		payload.Name = diff.Current.GetName()
		payload.Namespace = diff.Current.GetNamespace()
		payload.Action = "modified"
		previousDiff, currentDiff := GetMinimalDifference(diff)
		payload.Previous = previousDiff
		payload.Current = currentDiff
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	if payload.Namespace == "" {
		telemetry.Info(
			l.logger,
			string(jsonBytes),
			log.String(fmt.Sprintf("k8s.%s.name", strings.ToLower(payload.Kind)), payload.Name),
		)
	} else {
		telemetry.Info(
			l.logger,
			string(jsonBytes),
			log.String(fmt.Sprintf("k8s.%s.name", strings.ToLower(payload.Kind)), payload.Name),
			log.String("k8s.namespace.name", payload.Namespace),
		)
	}
}

func GetMinimalDifference(diff *manifest.Diff) (interface{}, interface{}) {
	if diff == nil {
		return nil, nil
	}
	return computeDiff(runtime.DeepCopyJSON(diff.Previous.Object), runtime.DeepCopyJSON(diff.Current.Object))
}

func computeDiff(prev, curr map[string]interface{}) (map[string]interface{}, map[string]interface{}) {
	if prev == nil && curr == nil {
		return nil, nil
	}
	if reflect.DeepEqual(prev, curr) {
		return nil, nil
	}
	if prev == nil {
		return nil, runtime.DeepCopyJSON(curr)
	}
	if curr == nil {
		return runtime.DeepCopyJSON(prev), nil
	}

	prevResult := make(map[string]interface{})
	currResult := make(map[string]interface{})

	keys := make(map[string]struct{})
	for k := range prev {
		keys[k] = struct{}{}
	}
	for k := range curr {
		keys[k] = struct{}{}
	}

	for key := range keys {
		pVal, pOk := prev[key]
		cVal, cOk := curr[key]

		switch {
		case !pOk:
			currResult[key] = runtime.DeepCopyJSONValue(cVal)
		case !cOk:
			prevResult[key] = runtime.DeepCopyJSONValue(pVal)
		default:
			pMap, pIsMap := pVal.(map[string]interface{})
			cMap, cIsMap := cVal.(map[string]interface{})
			if pIsMap && cIsMap {
				subPrev, subCurr := computeDiff(pMap, cMap)
				if len(subPrev) > 0 {
					prevResult[key] = subPrev
				}
				if len(subCurr) > 0 {
					currResult[key] = subCurr
				}
				continue
			}
			if reflect.DeepEqual(pVal, cVal) {
				continue
			}
			prevResult[key] = runtime.DeepCopyJSONValue(pVal)
			currResult[key] = runtime.DeepCopyJSONValue(cVal)
		}
	}

	if len(prevResult) == 0 {
		prevResult = nil
	}
	if len(currResult) == 0 {
		currResult = nil
	}
	return prevResult, currResult
}

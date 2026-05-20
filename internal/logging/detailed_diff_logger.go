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
			log.String("action", payload.Action),
			log.String(fmt.Sprintf("k8s.%s.name", strings.ToLower(payload.Kind)), payload.Name),
		)
	} else {
		telemetry.Info(
			l.logger,
			string(jsonBytes),
			log.String("action", payload.Action),
			log.String(fmt.Sprintf("k8s.%s.name", strings.ToLower(payload.Kind)), payload.Name),
			log.String("k8s.namespace.name", payload.Namespace),
		)
	}
}

func GetMinimalDifference(diff *manifest.Diff) (interface{}, interface{}) {
	if diff == nil {
		return nil, nil
	}
	var prevMap map[string]interface{}
	if diff.Previous != nil {
		prevMap = runtime.DeepCopyJSON(diff.Previous.Object)
	}
	var currMap map[string]interface{}
	if diff.Current != nil {
		currMap = runtime.DeepCopyJSON(diff.Current.Object)
	}
	return computeDiff(prevMap, currMap)
}

// computeDiff recursively produces minimal prev/curr maps containing only changed fields.
// Call structure:
//   - Nested maps: computeDiff recurses directly.
//   - Slices with a merge key (e.g. "name"): delegates to diffSlices → diffSlicesByKey,
//     which matches elements by key and calls back into computeDiff per pair.
//   - Slices without a merge key: treated atomically (full slice returned, no recursion).
//   - Scalar values: compared with DeepEqual; included only when different.
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
			pSlice, pIsSlice := pVal.([]interface{})
			cSlice, cIsSlice := cVal.([]interface{})
			if pIsSlice && cIsSlice {
				subPrev, subCurr := diffSlices(pSlice, cSlice)
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

// mergeKeyCandidates lists field names, in priority order, that are treated
// as natural merge keys for slices of objects. `name` covers the vast
// majority of Kubernetes keyed lists (containers, env, ports, volumes,
// volumeMounts, initContainers, ...).
var mergeKeyCandidates = []string{"name"}

func diffSlices(prev, curr []interface{}) ([]interface{}, []interface{}) {
	if reflect.DeepEqual(prev, curr) {
		return nil, nil
	}
	if key := findMergeKey(prev, curr); key != "" {
		return diffSlicesByKey(prev, curr, key)
	}
	prevCopy, _ := runtime.DeepCopyJSONValue(prev).([]interface{})
	currCopy, _ := runtime.DeepCopyJSONValue(curr).([]interface{})
	return prevCopy, currCopy
}

func findMergeKey(prev, curr []interface{}) string {
	if len(prev) == 0 || len(curr) == 0 {
		return ""
	}
	for _, candidate := range mergeKeyCandidates {
		if allHaveKey(prev, candidate) && allHaveKey(curr, candidate) {
			return candidate
		}
	}
	return ""
}

func allHaveKey(slice []interface{}, key string) bool {
	for _, v := range slice {
		m, ok := v.(map[string]interface{})
		if !ok {
			return false
		}
		if _, exists := m[key]; !exists {
			return false
		}
	}
	return true
}

func diffSlicesByKey(prev, curr []interface{}, key string) ([]interface{}, []interface{}) {
	prevIndex := make(map[string]map[string]interface{}, len(prev))
	prevOrder := make([]string, 0, len(prev))
	for _, v := range prev {
		m := v.(map[string]interface{})
		k := fmt.Sprint(m[key])
		if _, seen := prevIndex[k]; !seen {
			prevOrder = append(prevOrder, k)
		}
		prevIndex[k] = m
	}
	currIndex := make(map[string]map[string]interface{}, len(curr))
	currOrder := make([]string, 0, len(curr))
	for _, v := range curr {
		m := v.(map[string]interface{})
		k := fmt.Sprint(m[key])
		if _, seen := currIndex[k]; !seen {
			currOrder = append(currOrder, k)
		}
		currIndex[k] = m
	}

	var prevResult, currResult []interface{}
	for _, k := range prevOrder {
		pElem := prevIndex[k]
		cElem, ok := currIndex[k]
		if !ok {
			prevResult = append(prevResult, runtime.DeepCopyJSONValue(pElem))
			continue
		}
		subPrev, subCurr := computeDiff(pElem, cElem)
		if len(subPrev) == 0 && len(subCurr) == 0 {
			continue
		}
		if subPrev == nil {
			subPrev = map[string]interface{}{}
		}
		if subCurr == nil {
			subCurr = map[string]interface{}{}
		}
		subPrev[key] = runtime.DeepCopyJSONValue(pElem[key])
		subCurr[key] = runtime.DeepCopyJSONValue(cElem[key])
		prevResult = append(prevResult, subPrev)
		currResult = append(currResult, subCurr)
	}
	for _, k := range currOrder {
		if _, ok := prevIndex[k]; ok {
			continue
		}
		currResult = append(currResult, runtime.DeepCopyJSONValue(currIndex[k]))
	}
	return prevResult, currResult
}

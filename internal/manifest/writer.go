package manifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	"github.com/grafana/k8s-manifest-tail/internal/config"
)

// Writer persists manifests to disk in the desired format.
type Writer struct {
	baseDir string
	format  config.OutputFormat
}

// NewWriter builds a manifest writer for the supplied configuration.
func NewWriter(cfg config.OutputConfig) *Writer {
	return &Writer{
		baseDir: cfg.Directory,
		format:  cfg.Format,
	}
}

// Process saves the manifest for the supplied object to disk and reports differences.
func (w *Writer) Process(rule config.ObjectRule, obj *unstructured.Unstructured, _ *config.Config) (*Diff, error) {
	if err := w.ensureBaseDir(); err != nil {
		return nil, err
	}

	kindDir := filepath.Join(w.baseDir, sanitizePathSegment(rule.Kind))
	nsDir := filepath.Join(kindDir, sanitizePathSegment(namespaceSegment(obj.GetNamespace())))
	fileName := fmt.Sprintf("%s.%s", sanitizePathSegment(obj.GetName()), w.extension())
	path := filepath.Join(nsDir, fileName)

	prevObj, prevJSON, err := w.loadExisting(path)
	if err != nil {
		return nil, err
	}

	rawJSON, err := obj.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal object: %w", err)
	}
	newJSON, err := canonicalizeJSON(rawJSON)
	if err != nil {
		return nil, fmt.Errorf("canonicalize object json: %w", err)
	}

	if prevJSON != nil && bytes.Equal(prevJSON, newJSON) {
		return nil, nil
	}

	data, err := w.serialize(obj)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(nsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create directory %s: %w", nsDir, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return nil, fmt.Errorf("write manifest %s: %w", path, err)
	}

	return &Diff{
		Previous: prevObj,
		Current:  obj.DeepCopy(),
	}, nil
}

// Delete removes the manifest for the supplied object from disk.
func (w *Writer) Delete(rule config.ObjectRule, obj *unstructured.Unstructured, _ *config.Config) error {
	if err := w.ensureBaseDir(); err != nil {
		return err
	}
	kindDir := filepath.Join(w.baseDir, sanitizePathSegment(rule.Kind))
	nsDir := filepath.Join(kindDir, sanitizePathSegment(namespaceSegment(obj.GetNamespace())))
	fileName := fmt.Sprintf("%s.%s", sanitizePathSegment(obj.GetName()), w.extension())
	path := filepath.Join(nsDir, fileName)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove manifest %s: %w", path, err)
	}
	return nil
}

func (w *Writer) ensureBaseDir() error {
	if w.baseDir == "" {
		return fmt.Errorf("output directory is required")
	}
	return nil
}

func (w *Writer) serialize(obj *unstructured.Unstructured) ([]byte, error) {
	jsonBytes, err := obj.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal object: %w", err)
	}

	switch w.format {
	case config.OutputFormatJSON:
		var buf bytes.Buffer
		if err := json.Indent(&buf, jsonBytes, "", "  "); err != nil {
			return nil, fmt.Errorf("format json: %w", err)
		}
		buf.WriteByte('\n')
		return buf.Bytes(), nil
	case config.OutputFormatYAML:
		yamlBytes, err := yaml.JSONToYAML(jsonBytes)
		if err != nil {
			return nil, fmt.Errorf("convert to yaml: %w", err)
		}
		if len(yamlBytes) == 0 || yamlBytes[len(yamlBytes)-1] != '\n' {
			yamlBytes = append(yamlBytes, '\n')
		}
		return yamlBytes, nil
	default:
		return nil, fmt.Errorf("unsupported output format %q", w.format)
	}
}

func (w *Writer) extension() string {
	if w.format == config.OutputFormatJSON {
		return "json"
	}
	return "yaml"
}

func namespaceSegment(ns string) string {
	if ns == "" {
		return "cluster"
	}
	return ns
}

func sanitizePathSegment(value string) string {
	if value == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer("/", "_", "\\", "_", " ", "_")
	return replacer.Replace(value)
}

func (w *Writer) loadExisting(path string) (*unstructured.Unstructured, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("read existing manifest %s: %w", path, err)
	}
	jsonBytes, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, nil, fmt.Errorf("convert existing manifest %s to json: %w", path, err)
	}
	canonical, err := canonicalizeJSON(jsonBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("canonicalize existing manifest %s: %w", path, err)
	}
	obj := &unstructured.Unstructured{}
	if err := obj.UnmarshalJSON(canonical); err != nil {
		return nil, nil, fmt.Errorf("decode existing manifest %s: %w", path, err)
	}
	return obj, canonical, nil
}

func canonicalizeJSON(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, input); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

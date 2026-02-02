package manifest

import (
	"bytes"
	"encoding/json"
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

// Process saves the manifest for the supplied object to disk.
func (w *Writer) Process(rule config.ObjectRule, obj *unstructured.Unstructured) error {
	if err := w.ensureBaseDir(); err != nil {
		return err
	}

	kindDir := filepath.Join(w.baseDir, sanitizePathSegment(rule.Kind))
	nsDir := filepath.Join(kindDir, sanitizePathSegment(namespaceSegment(obj.GetNamespace())))
	if err := os.MkdirAll(nsDir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", nsDir, err)
	}

	fileName := fmt.Sprintf("%s.%s", sanitizePathSegment(obj.GetName()), w.extension())
	path := filepath.Join(nsDir, fileName)
	data, err := w.serialize(obj)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write manifest %s: %w", path, err)
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
	switch w.format {
	case config.OutputFormatJSON:
		return "json"
	default:
		return "yaml"
	}
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

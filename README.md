# K8s Manifest Tail

A utility for fetching Kubernetes Manifest documents from a running cluster. This utility can be run inside or outside 
a Kubernetes cluster, and utilizes a config file to determine what kind of objects to detect. Manifests files are stored
in an output directory in the format: `<outputDir>/<kind>/<namespace>/<name>.yaml`

## Running

This utility supports a few methods for running:

* `describe` - Reads the config file and prints a human description of what resources will be fetched (e.g., "Deployments in the `default` namespace"). Useful for validating your configuration before contacting the cluster.
* `list` - Simply list the objects that would be detected by this utility. Runs and exits.
* `run-once` - Runs once, gathering the manifest files and exiting.
* `run` - Runs once, gathering the manifest files, and then sets up watchers to monitor for additions, 
  deletions, or modifications and updates or deletes those manifest files.

## Configuration

This utility requires a config.yaml file that at minimum defines the objects list. This list defines what kinds of
Kubernetes objects to detect and fetch manifests for. Other options may be configurable from the command line.

### Configuration file

```yaml
# Configuration for manifest output
output:
  # The directory path, relative to the working directory, where to store manifest files.
  directory: output

  # The format to use when saving manifest files. Valid options: yaml, json
  format: yaml

# When to fetch another full set of all objects
refreshInterval: 1d

# Namespaces to look for any objects
namespaces: []

# Namespaces to skip for any objects
excludeNamespaces: []

# Rules per kind
objects:
  - apiVersion: v1
    kind: Pod
  - apiVersion: apps/v1
    kind: Deployment
    namespaces: [default]  # This list of namespaces takes precedent over the global namespace list.
    namePattern: "alloy-.*" # Optional regular expression to match object names.
```

This config file will get manifests for all Pods, and Deployments within the `default` namespace whose names match the
regular expression `alloy-.*` (for example `alloy-logs` or `alloy-metrics`). It will store them as YAML files inside the
directory named `output`.

### OTLP logging

Set the `logging.otlp` block in `config.yaml` (or the CLI/env overrides) to emit OpenTelemetry logs. Any of the
standard `OTEL_EXPORTER_OTLP_*` or `OTEL_EXPORTER_OTLP_LOGS_*` environment variables (for example
`OTEL_EXPORTER_OTLP_LOGS_ENDPOINT`) automatically enable the OTLP exporter and are passed straight through to the
underlying OpenTelemetry Go exporter. Toggle `logging.logManifests` (or the env var
`K8S_MANIFEST_TAIL_LOGGING_LOG_MANIFESTS=true`) to include the manifest payloads themselves in the emitted log
records, which pairs well with diff logging or OTLP shipping.

### Configuration Flags

The following configuration flags may be passed via the command-line:

* -f|--output-format <json|yaml> (default: "yaml")
* -o|--output-directory <string> (default: "output")
* --refresh-interval <duration> (default: "1d")
* -n|--namespaces <string list> (default: []) - The list of namespaces to look for *any* objects. Empty means look in all namespaces.
* --exclude-namespaces <string list> (default: []) - The list of namespaces to skip when looking for *any* objects.

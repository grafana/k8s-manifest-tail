# Development Plan

## Design Principles

- Favor small, testable functions with clear responsibilities.
- Every feature should be covered by unit tests and, when applicable, higher-level/system tests.
- Validate and propagate all errors with informative messages; never ignore returned errors.
- Keep observability (traces, metrics, logs) in mind for every component to aid debugging.
- Prefer incremental, reversible changes to keep the plan adaptable.

1. **Config Schema Alignment** *(completed: config loader + updated sample)*
   - Confirm the shipped `config.yaml` keys (`outputDir`, `outputFormat`, etc.) match the README schema (`output.directory`, `output.format`).
   - Decide on a canonical structure, update docs or sample accordingly, and add a minimal unit test that loads the config fixture to guard against regressions.

2. **Configuration Loader** *(completed: YAML loader + CLI overrides/tests)*
   - Implement a small package that reads the YAML config, applies command-line overrides, and validates required fields (e.g., at least one object definition, supported output formats).
   - Cover parsing/validation with table-driven tests using in-memory YAML strings.

3. **CLI Skeleton** *(completed: Cobra root with describe/list/run/run-and-watch commands and shared flags)*
   - Introduce Cobra with three commands: `list`, `run`, and `run-and-watch`.
   - Wire up shared persistent flags (kubeconfig, output directory/format, refresh interval, namespace filters) and ensure `k8s-manifest-tail list` prints configured targets using the loaded config.

4. **Behavior-Driven Test Harness**
   - Establish a BDD-style testing framework (e.g., Ginkgo/Gomega or similar helpers) for CLI behavior, starting with integration-ish tests that assert `--help` output and `describe` command behavior.
   - Create reusable fixtures/helpers so new commands can easily add behavioral tests.

5. **OpenTelemetry Baseline**
   - Define the initial OpenTelemetry tracer/meter setup so subsequent packages can emit spans/metrics.
   - Provide configuration flags/env vars for enabling/disabling telemetry and document default exporters.

6. **Kubernetes Client Factory**
   - Add a helper that builds a rest.Config via in-cluster detection or kubeconfig path and surfaces meaningful errors.
   - Unit test the helper by mocking environment variables and kubeconfig paths where possible.

7. **Object Discovery & Fetch**
   - Implement logic that, for each configured object rule, lists matching resources (respecting include/exclude namespaces) and fetches their manifests.
   - Start with support for a single simple kind (e.g., Pods) and add tests using the fake Kubernetes client.

8. **Manifest Serialization**
   - Add writer utilities that render the fetched manifests as YAML/JSON and save them into `<outputDir>/<kind>/<namespace>/<name>.yaml|json`, creating directories as needed.
   - Include unit tests that write to a temp directory and assert the expected files and formats.

9. **`run` Command Execution Path**
   - Wire the config, discovery, and writer pieces together so `run` performs a full fetch cycle once, returning errors via exit codes.
   - Add an integration-style test (using the fake client) that ensures desired files are produced for a sample config.

10. **`run-and-watch` Continuous Mode**
   - Introduce informers/watchers for each requested kind so the tool updates or removes manifests on add/update/delete events.
   - Provide a mechanism to respect `refreshInterval` for periodic full syncs and ensure graceful shutdown (context cancellation, signal handling).

11. **Polish & Observability**
   - Add structured logging, metrics hooks (if needed), and clear error messages.
   - Document CLI usage examples in README and ensure `make build` produces the binary into `build/`.

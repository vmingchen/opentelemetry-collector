# Prometheus Exporter

Exports metric data to a [Prometheus](https://prometheus.io/) back-end.

The following settings are required:

- `endpoint` (no default): Where to send metric data

The following settings can be optionally configured:

- `constlabels` (no default): key/values that are applied for every exported metric.
- `namespace` (no default): if set, exports metrics under the provided value.

Example:

```yaml
exporters:
  prometheus:
    endpoint: "1.2.3.4:1234"
    namespace: test-space
    const_labels:
      label1: value1
      "another label": spaced value
```

The full list of settings exposed for this exporter are documented [here](./config.go)
with detailed sample configurations [here](./testdata/config.yaml).

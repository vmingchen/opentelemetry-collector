// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fileexporter

import (
	"context"
	"os"

	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
)

const (
	// The value of "type" key in configuration.
	typeStr = "file"
)

// Factory is the factory for logging exporter.
type Factory struct {
}

// Type gets the type of the Exporter config created by this factory.
func (f *Factory) Type() configmodels.Type {
	return typeStr
}

// CreateDefaultConfig creates the default configuration for exporter.
func (f *Factory) CreateDefaultConfig() configmodels.Exporter {
	return &Config{
		ExporterSettings: configmodels.ExporterSettings{
			TypeVal: typeStr,
			NameVal: typeStr,
		},
	}
}

// CreateTraceExporter creates a trace exporter based on this config.
func (f *Factory) CreateTraceExporter(logger *zap.Logger, config configmodels.Exporter) (component.TraceExporterOld, error) {
	return f.createExporter(config)
}

// CreateMetricsExporter creates a metrics exporter based on this config.
func (f *Factory) CreateMetricsExporter(logger *zap.Logger, config configmodels.Exporter) (component.MetricsExporterOld, error) {
	return f.createExporter(config)
}

// CreateLogExporter creates a log exporter based on this config.
func (f *Factory) CreateLogExporter(
	ctx context.Context,
	params component.ExporterCreateParams,
	cfg configmodels.Exporter,
) (component.LogExporter, error) {
	return f.createExporter(cfg)
}

var _ component.LogExporterFactory = (*Factory)(nil)

func (f *Factory) createExporter(config configmodels.Exporter) (*Exporter, error) {
	cfg := config.(*Config)

	// There must be one exporter for metrics, traces, and logs. We maintain a
	// map of exporters per config.

	// Check to see if there is already a exporter for this config.
	exporter, ok := exporters[cfg]

	if !ok {
		file, err := os.OpenFile(cfg.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return nil, err
		}
		exporter = &Exporter{file: file}

		// Remember the receiver in the map
		exporters[cfg] = exporter
	}
	return exporter, nil
}

// This is the map of already created File exporters for particular configurations.
// We maintain this map because the Factory is asked trace and metric receivers separately
// when it gets CreateTraceReceiver() and CreateMetricsReceiver() but they must not
// create separate objects, they must use one Receiver object per configuration.
var exporters = map[*Config]*Exporter{}

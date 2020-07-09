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

package loggingexporter

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
)

const (
	// The value of "type" key in configuration.
	typeStr                   = "logging"
	defaultSamplingInitial    = 2
	defaultSamplingThereafter = 500
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
		LogLevel:           "info",
		SamplingInitial:    defaultSamplingInitial,
		SamplingThereafter: defaultSamplingThereafter,
	}
}

// CreateTraceExporter creates a trace exporter based on this config.
func (f *Factory) CreateTraceExporter(_ context.Context, _ component.ExporterCreateParams, config configmodels.Exporter) (component.TraceExporter, error) {
	cfg := config.(*Config)

	exporterLogger, err := f.createLogger(cfg)
	if err != nil {
		return nil, err
	}

	lexp, err := NewTraceExporter(config, cfg.LogLevel, exporterLogger)
	if err != nil {
		return nil, err
	}
	return lexp, nil
}

func (f *Factory) createLogger(cfg *Config) (*zap.Logger, error) {
	var level zapcore.Level
	err := (&level).UnmarshalText([]byte(cfg.LogLevel))
	if err != nil {
		return nil, err
	}

	// We take development config as the base since it matches the purpose
	// of logging exporter being used for debugging reasons (so e.g. console encoder)
	conf := zap.NewDevelopmentConfig()
	conf.Level = zap.NewAtomicLevelAt(level)
	conf.Sampling = &zap.SamplingConfig{
		Initial:    cfg.SamplingInitial,
		Thereafter: cfg.SamplingThereafter,
	}

	logginglogger, err := conf.Build()
	if err != nil {
		return nil, err
	}
	return logginglogger, nil
}

// CreateMetricsExporter creates a metrics exporter based on this config.
func (f *Factory) CreateMetricsExporter(_ context.Context, _ component.ExporterCreateParams, config configmodels.Exporter) (component.MetricsExporter, error) {
	cfg := config.(*Config)

	exporterLogger, err := f.createLogger(cfg)
	if err != nil {
		return nil, err
	}

	lexp, err := NewMetricsExporter(config, cfg.LogLevel, exporterLogger)
	if err != nil {
		return nil, err
	}
	return lexp, nil
}

// CreateLogExporter creates a log exporter based on this config.
func (f *Factory) CreateLogExporter(_ context.Context, _ component.ExporterCreateParams, config configmodels.Exporter) (component.LogExporter, error) {
	cfg := config.(*Config)

	exporterLogger, err := f.createLogger(cfg)
	if err != nil {
		return nil, err
	}

	lexp, err := NewLogExporter(config, cfg.LogLevel, exporterLogger)
	if err != nil {
		return nil, err
	}
	return lexp, nil
}

var _ component.LogExporterFactory = (*Factory)(nil)

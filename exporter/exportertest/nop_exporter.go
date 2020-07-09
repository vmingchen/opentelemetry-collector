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

package exportertest

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumerdata"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/internal/data"
)

type nopExporterOld struct {
	name     string
	retError error
}

func (ne *nopExporterOld) Start(ctx context.Context, host component.Host) error {
	return nil
}

func (ne *nopExporterOld) ConsumeTraceData(_ context.Context, _ consumerdata.TraceData) error {
	return ne.retError
}

func (ne *nopExporterOld) ConsumeMetricsData(_ context.Context, _ consumerdata.MetricsData) error {
	return ne.retError
}

// Shutdown stops the exporter and is invoked during shutdown.
func (ne *nopExporterOld) Shutdown(context.Context) error {
	return nil
}

const (
	nopTraceExporterName   = "nop_trace"
	nopMetricsExporterName = "nop_metrics"
	nopLogExporterName     = "nop_log"
)

// NewNopTraceExporterOld creates an TraceExporter that just drops the received data.
func NewNopTraceExporterOld() component.TraceExporterOld {
	ne := &nopExporterOld{
		name: nopTraceExporterName,
	}
	return ne
}

// NewNopMetricsExporterOld creates an MetricsExporter that just drops the received data.
func NewNopMetricsExporterOld() component.MetricsExporterOld {
	ne := &nopExporterOld{
		name: nopMetricsExporterName,
	}
	return ne
}

type nopExporter struct {
	name     string
	retError error
}

func (ne *nopExporter) Start(ctx context.Context, host component.Host) error {
	return nil
}

func (ne *nopExporter) ConsumeTraces(_ context.Context, _ pdata.Traces) error {
	return ne.retError
}

func (ne *nopExporter) ConsumeMetrics(ctx context.Context, md pdata.Metrics) error {
	return ne.retError
}

func (ne *nopExporter) ConsumeLogs(ctx context.Context, ld data.Logs) error {
	return ne.retError
}

// Shutdown stops the exporter and is invoked during shutdown.
func (ne *nopExporter) Shutdown(context.Context) error {
	return nil
}

// NewNopTraceExporterOld creates an TraceExporter that just drops the received data.
func NewNopTraceExporter() component.TraceExporter {
	ne := &nopExporter{
		name: nopTraceExporterName,
	}
	return ne
}

// NewNopMetricsExporterOld creates an MetricsExporter that just drops the received data.
func NewNopMetricsExporter() component.MetricsExporter {
	ne := &nopExporter{
		name: nopMetricsExporterName,
	}
	return ne
}

// NewNopLogExporterOld creates an LogExporter that just drops the received data.
func NewNopLogExporter() component.LogExporter {
	ne := &nopExporter{
		name: nopLogExporterName,
	}
	return ne
}

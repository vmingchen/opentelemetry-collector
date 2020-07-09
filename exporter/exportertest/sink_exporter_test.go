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
	"testing"

	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	tracepb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/consumer/consumerdata"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/consumer/pdatautil"
	"go.opentelemetry.io/collector/internal/data"
	"go.opentelemetry.io/collector/internal/data/testdata"
)

func TestSinkTraceExporterOld(t *testing.T) {
	sink := new(SinkTraceExporterOld)
	td := consumerdata.TraceData{
		Spans: make([]*tracepb.Span, 7),
	}
	want := make([]consumerdata.TraceData, 0, 7)
	for i := 0; i < 7; i++ {
		err := sink.ConsumeTraceData(context.Background(), td)
		require.Nil(t, err)
		want = append(want, td)
	}
	got := sink.AllTraces()
	assert.Equal(t, want, got)
}

func TestSinkTraceExporter(t *testing.T) {
	sink := new(SinkTraceExporter)
	td := testdata.GenerateTraceDataOneSpan()
	want := make([]pdata.Traces, 0, 7)
	for i := 0; i < 7; i++ {
		err := sink.ConsumeTraces(context.Background(), td)
		require.Nil(t, err)
		want = append(want, td)
	}
	got := sink.AllTraces()
	assert.Equal(t, want, got)
}

func TestSinkMetricsExporterOld(t *testing.T) {
	sink := new(SinkMetricsExporterOld)
	md := consumerdata.MetricsData{
		Metrics: make([]*metricspb.Metric, 7),
	}
	want := make([]consumerdata.MetricsData, 0, 7)
	for i := 0; i < 7; i++ {
		err := sink.ConsumeMetricsData(context.Background(), md)
		require.Nil(t, err)
		want = append(want, md)
	}
	got := sink.AllMetrics()
	assert.Equal(t, want, got)
}

func TestSinkMetricsExporter(t *testing.T) {
	sink := new(SinkMetricsExporter)
	md := testdata.GenerateMetricDataOneMetric()
	want := make([]pdata.Metrics, 0, 7)
	for i := 0; i < 7; i++ {
		err := sink.ConsumeMetrics(context.Background(), pdatautil.MetricsFromInternalMetrics(md))
		require.Nil(t, err)
		want = append(want, pdatautil.MetricsFromInternalMetrics(md))
	}
	got := sink.AllMetrics()
	assert.Equal(t, want, got)
}

func TestSinkLogExporter(t *testing.T) {
	sink := new(SinkLogExporter)
	md := testdata.GenerateLogDataOneLogNoResource()
	want := make([]data.Logs, 0, 7)
	for i := 0; i < 7; i++ {
		err := sink.ConsumeLogs(context.Background(), md)
		require.Nil(t, err)
		want = append(want, md)
	}
	got := sink.AllLogs()
	assert.Equal(t, want, got)
}

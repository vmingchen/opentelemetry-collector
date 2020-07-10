// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// obsreport_test instead of just obsreport to avoid dependency cycle between
// obsreport_test and obsreporttest
package obsreport_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"

	"go.opentelemetry.io/collector/obsreport"
	"go.opentelemetry.io/collector/obsreport/obsreporttest"
)

const (
	exporter   = "fakeExporter"
	processor  = "fakeProcessor"
	receiver   = "fakeReicever"
	transport  = "fakeTransport"
	format     = "fakeFormat"
	legacyName = "fakeLegacyName"
)

var (
	errFake = errors.New("errFake")
)

type receiveTestParams struct {
	transport string
	err       error
}

func TestConfigure(t *testing.T) {
	type args struct {
		generateLegacy bool
		generateNew    bool
	}
	tests := []struct {
		name      string
		args      args
		wantViews []*view.View
	}{
		{
			name: "none",
		},
		{
			name: "legacy_only",
			args: args{
				generateLegacy: true,
			},
			wantViews: obsreport.LegacyAllViews,
		},
		{
			name: "new_only",
			args: args{
				generateNew: true,
			},
			wantViews: obsreport.AllViews(),
		},
		{
			name: "new_only",
			args: args{
				generateNew:    true,
				generateLegacy: true,
			},
			wantViews: append(
				obsreport.LegacyAllViews,
				obsreport.AllViews()...),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotViews := obsreport.Configure(tt.args.generateLegacy, tt.args.generateNew)
			assert.Equal(t, tt.wantViews, gotViews)
		})
	}
}

func TestReceiveTraceDataOp(t *testing.T) {
	doneFn, err := obsreporttest.SetupRecordedMetricsTest()
	require.NoError(t, err)
	defer doneFn()

	ss := &spanStore{}
	trace.RegisterExporter(ss)
	defer trace.UnregisterExporter(ss)

	parentCtx, parentSpan := trace.StartSpan(context.Background(),
		t.Name(), trace.WithSampler(trace.AlwaysSample()))
	defer parentSpan.End()

	receiverCtx := obsreport.ReceiverContext(parentCtx, receiver, transport, "")
	params := []receiveTestParams{
		{transport, errFake},
		{"", nil},
	}
	rcvdSpans := []int{13, 42}
	for i, param := range params {
		ctx := obsreport.StartTraceDataReceiveOp(receiverCtx, receiver, param.transport)
		assert.NotNil(t, ctx)

		obsreport.EndTraceDataReceiveOp(
			ctx,
			format,
			rcvdSpans[i],
			param.err)
	}

	spans := ss.PullAllSpans()
	require.Equal(t, len(params), len(spans))

	var acceptedSpans, refusedSpans int
	for i, span := range spans {
		assert.Equal(t, "receiver/"+receiver+"/TraceDataReceived", span.Name)
		switch params[i].err {
		case nil:
			acceptedSpans += rcvdSpans[i]
			assert.Equal(t, int64(rcvdSpans[i]), span.Attributes[obsreport.AcceptedSpansKey])
			assert.Equal(t, int64(0), span.Attributes[obsreport.RefusedSpansKey])
			assert.Equal(t, trace.Status{Code: trace.StatusCodeOK}, span.Status)
		case errFake:
			refusedSpans += rcvdSpans[i]
			assert.Equal(t, int64(0), span.Attributes[obsreport.AcceptedSpansKey])
			assert.Equal(t, int64(rcvdSpans[i]), span.Attributes[obsreport.RefusedSpansKey])
			assert.Equal(t, params[i].err.Error(), span.Status.Message)
		default:
			t.Fatalf("unexpected param: %v", params[i])
		}
		switch params[i].transport {
		case "":
			assert.NotContains(t, span.Attributes, obsreport.TransportKey)
		default:
			assert.Equal(t, params[i].transport, span.Attributes[obsreport.TransportKey])
		}
	}
	// Check legacy metrics.
	legacyReceiverTags := []tag.Tag{{Key: obsreport.LegacyTagKeyReceiver, Value: receiver}}
	obsreporttest.CheckValueForView(t, legacyReceiverTags, int64(acceptedSpans), obsreport.LegacyViewReceiverReceivedSpans.Name)
	obsreporttest.CheckValueForView(t, legacyReceiverTags, int64(refusedSpans), obsreport.LegacyViewReceiverDroppedSpans.Name)

	// Check new metrics.
	obsreporttest.CheckReceiverTracesViews(t, receiver, transport, int64(acceptedSpans), int64(refusedSpans))
}

func TestReceiveMetricsOp(t *testing.T) {
	doneFn, err := obsreporttest.SetupRecordedMetricsTest()
	require.NoError(t, err)
	defer doneFn()

	ss := &spanStore{}
	trace.RegisterExporter(ss)
	defer trace.UnregisterExporter(ss)

	parentCtx, parentSpan := trace.StartSpan(context.Background(),
		t.Name(), trace.WithSampler(trace.AlwaysSample()))
	defer parentSpan.End()

	receiverCtx := obsreport.ReceiverContext(parentCtx, receiver, transport, "")
	params := []receiveTestParams{
		{transport, errFake},
		{"", nil},
	}
	rcvdMetricPts := []int{23, 29}
	rcvdTimeSeries := []int{2, 3}
	for i, param := range params {
		ctx := obsreport.StartMetricsReceiveOp(receiverCtx, receiver, param.transport)
		assert.NotNil(t, ctx)

		obsreport.EndMetricsReceiveOp(
			ctx,
			format,
			rcvdMetricPts[i],
			rcvdTimeSeries[i],
			param.err)
	}

	spans := ss.PullAllSpans()
	require.Equal(t, len(params), len(spans))

	var receivedTimeSeries, droppedTimeSeries int
	var acceptedMetricPoints, refusedMetricPoints int
	for i, span := range spans {
		assert.Equal(t, "receiver/"+receiver+"/MetricsReceived", span.Name)
		switch params[i].err {
		case nil:
			receivedTimeSeries += rcvdTimeSeries[i]
			acceptedMetricPoints += rcvdMetricPts[i]
			assert.Equal(t, int64(rcvdMetricPts[i]), span.Attributes[obsreport.AcceptedMetricPointsKey])
			assert.Equal(t, int64(0), span.Attributes[obsreport.RefusedMetricPointsKey])
			assert.Equal(t, trace.Status{Code: trace.StatusCodeOK}, span.Status)
		case errFake:
			droppedTimeSeries += rcvdTimeSeries[i]
			refusedMetricPoints += rcvdMetricPts[i]
			assert.Equal(t, int64(0), span.Attributes[obsreport.AcceptedMetricPointsKey])
			assert.Equal(t, int64(rcvdMetricPts[i]), span.Attributes[obsreport.RefusedMetricPointsKey])
			assert.Equal(t, params[i].err.Error(), span.Status.Message)
		default:
			t.Fatalf("unexpected param: %v", params[i])
		}
		switch params[i].transport {
		case "":
			assert.NotContains(t, span.Attributes, obsreport.TransportKey)
		default:
			assert.Equal(t, params[i].transport, span.Attributes[obsreport.TransportKey])
		}
	}

	// Check legacy metrics.
	legacyReceiverTags := []tag.Tag{{Key: obsreport.LegacyTagKeyReceiver, Value: receiver}}
	obsreporttest.CheckValueForView(t, legacyReceiverTags, int64(receivedTimeSeries), obsreport.LegacyViewReceiverReceivedTimeSeries.Name)
	obsreporttest.CheckValueForView(t, legacyReceiverTags, int64(droppedTimeSeries), obsreport.LegacyViewReceiverDroppedTimeSeries.Name)

	// Check new metrics.
	obsreporttest.CheckReceiverMetricsViews(t, receiver, transport, int64(acceptedMetricPoints), int64(refusedMetricPoints))
}

func TestExportTraceDataOp(t *testing.T) {
	doneFn, err := obsreporttest.SetupRecordedMetricsTest()
	require.NoError(t, err)
	defer doneFn()

	ss := &spanStore{}
	trace.RegisterExporter(ss)
	defer trace.UnregisterExporter(ss)

	parentCtx, parentSpan := trace.StartSpan(context.Background(),
		t.Name(), trace.WithSampler(trace.AlwaysSample()))
	defer parentSpan.End()

	exporterCtx := obsreport.ExporterContext(parentCtx, exporter)
	errs := []error{nil, errFake}
	numExportedSpans := []int{22, 14}
	for i, err := range errs {
		ctx := obsreport.StartTraceDataExportOp(exporterCtx, exporter)
		assert.NotNil(t, ctx)

		var numDroppedSpans int
		if err != nil {
			numDroppedSpans = numExportedSpans[i]
		}

		obsreport.EndTraceDataExportOp(ctx, numExportedSpans[i], numDroppedSpans, err)
	}

	spans := ss.PullAllSpans()
	require.Equal(t, len(errs), len(spans))

	var sentSpans, failedToSendSpans int
	for i, span := range spans {
		assert.Equal(t, "exporter/"+exporter+"/TraceDataExported", span.Name)
		switch errs[i] {
		case nil:
			sentSpans += numExportedSpans[i]
			assert.Equal(t, int64(numExportedSpans[i]), span.Attributes[obsreport.SentSpansKey])
			assert.Equal(t, int64(0), span.Attributes[obsreport.FailedToSendSpansKey])
			assert.Equal(t, trace.Status{Code: trace.StatusCodeOK}, span.Status)
		case errFake:
			failedToSendSpans += numExportedSpans[i]
			assert.Equal(t, int64(0), span.Attributes[obsreport.SentSpansKey])
			assert.Equal(t, int64(numExportedSpans[i]), span.Attributes[obsreport.FailedToSendSpansKey])
			assert.Equal(t, errs[i].Error(), span.Status.Message)
		default:
			t.Fatalf("unexpected error: %v", errs[i])
		}
	}

	// Check legacy metrics.
	legacyExporterTags := []tag.Tag{{Key: obsreport.LegacyTagKeyExporter, Value: exporter}}
	obsreporttest.CheckValueForView(t, legacyExporterTags, int64(sentSpans)+int64(failedToSendSpans), obsreport.LegacyViewExporterReceivedSpans.Name)
	obsreporttest.CheckValueForView(t, legacyExporterTags, int64(failedToSendSpans), obsreport.LegacyViewExporterDroppedSpans.Name)

	// Check new metrics.
	obsreporttest.CheckExporterTracesViews(t, exporter, int64(sentSpans), int64(failedToSendSpans))
}

func TestExportMetricsOp(t *testing.T) {
	doneFn, err := obsreporttest.SetupRecordedMetricsTest()
	require.NoError(t, err)
	defer doneFn()

	ss := &spanStore{}
	trace.RegisterExporter(ss)
	defer trace.UnregisterExporter(ss)

	parentCtx, parentSpan := trace.StartSpan(context.Background(),
		t.Name(), trace.WithSampler(trace.AlwaysSample()))
	defer parentSpan.End()

	exporterCtx := obsreport.ExporterContext(parentCtx, exporter)
	errs := []error{nil, errFake}
	toSendMetricPts := []int{17, 23}
	toSendTimeSeries := []int{3, 5}
	for i, err := range errs {
		ctx := obsreport.StartMetricsExportOp(exporterCtx, exporter)
		assert.NotNil(t, ctx)

		var numDroppedTimeSeires int
		if err != nil {
			numDroppedTimeSeires = toSendTimeSeries[i]
		}

		obsreport.EndMetricsExportOp(
			ctx,
			toSendMetricPts[i],
			toSendTimeSeries[i],
			numDroppedTimeSeires,
			err)
	}

	spans := ss.PullAllSpans()
	require.Equal(t, len(errs), len(spans))

	var receivedTimeSeries, droppedTimeSeries int
	var sentPoints, failedToSendPoints int
	for i, span := range spans {
		assert.Equal(t, "exporter/"+exporter+"/MetricsExported", span.Name)
		receivedTimeSeries += toSendTimeSeries[i]
		switch errs[i] {
		case nil:
			sentPoints += toSendMetricPts[i]
			assert.Equal(t, int64(toSendMetricPts[i]), span.Attributes[obsreport.SentMetricPointsKey])
			assert.Equal(t, int64(0), span.Attributes[obsreport.FailedToSendMetricPointsKey])
			assert.Equal(t, trace.Status{Code: trace.StatusCodeOK}, span.Status)
		case errFake:
			failedToSendPoints += toSendMetricPts[i]
			droppedTimeSeries += toSendTimeSeries[i]
			assert.Equal(t, int64(0), span.Attributes[obsreport.SentMetricPointsKey])
			assert.Equal(t, int64(toSendMetricPts[i]), span.Attributes[obsreport.FailedToSendMetricPointsKey])
			assert.Equal(t, errs[i].Error(), span.Status.Message)
		default:
			t.Fatalf("unexpected error: %v", errs[i])
		}
	}

	// Check legacy metrics.
	legacyExporterTags := []tag.Tag{{Key: obsreport.LegacyTagKeyExporter, Value: exporter}}
	obsreporttest.CheckValueForView(t, legacyExporterTags, int64(receivedTimeSeries), obsreport.LegacyViewExporterReceivedTimeSeries.Name)
	obsreporttest.CheckValueForView(t, legacyExporterTags, int64(droppedTimeSeries), obsreport.LegacyViewExporterDroppedTimeSeries.Name)

	// Check new metrics.
	obsreporttest.CheckExporterMetricsViews(t, exporter, int64(sentPoints), int64(failedToSendPoints))
}

func TestReceiveWithLongLivedCtx(t *testing.T) {
	ss := &spanStore{}
	trace.RegisterExporter(ss)
	defer trace.UnregisterExporter(ss)

	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})
	defer func() {
		trace.ApplyConfig(trace.Config{
			DefaultSampler: trace.ProbabilitySampler(1e-4),
		})
	}()

	parentCtx, parentSpan := trace.StartSpan(context.Background(), t.Name())
	defer parentSpan.End()

	longLivedCtx := obsreport.ReceiverContext(parentCtx, receiver, transport, legacyName)
	ops := []struct {
		numSpans int
		err      error
	}{
		{numSpans: 13},
		{numSpans: 42, err: errFake},
	}
	for _, op := range ops {
		// Use a new context on each operation to simulate distinct operations
		// under the same long lived context.
		ctx := obsreport.StartTraceDataReceiveOp(
			longLivedCtx,
			receiver,
			transport,
			obsreport.WithLongLivedCtx())
		assert.NotNil(t, ctx)

		obsreport.EndTraceDataReceiveOp(
			ctx,
			format,
			op.numSpans,
			op.err)
	}

	spans := ss.PullAllSpans()
	require.Equal(t, len(ops), len(spans))

	for i, span := range spans {
		assert.Equal(t, trace.SpanID{}, span.ParentSpanID)
		require.Equal(t, 1, len(span.Links))
		link := span.Links[0]
		assert.Equal(t, trace.LinkTypeParent, link.Type)
		assert.Equal(t, parentSpan.SpanContext().TraceID, link.TraceID)
		assert.Equal(t, parentSpan.SpanContext().SpanID, link.SpanID)
		assert.Equal(t, "receiver/"+receiver+"/TraceDataReceived", span.Name)
		assert.Equal(t, transport, span.Attributes[obsreport.TransportKey])
		switch ops[i].err {
		case nil:
			assert.Equal(t, int64(ops[i].numSpans), span.Attributes[obsreport.AcceptedSpansKey])
			assert.Equal(t, int64(0), span.Attributes[obsreport.RefusedSpansKey])
			assert.Equal(t, trace.Status{Code: trace.StatusCodeOK}, span.Status)
		case errFake:
			assert.Equal(t, int64(0), span.Attributes[obsreport.AcceptedSpansKey])
			assert.Equal(t, int64(ops[i].numSpans), span.Attributes[obsreport.RefusedSpansKey])
			assert.Equal(t, ops[i].err.Error(), span.Status.Message)
		default:
			t.Fatalf("unexpected error: %v", ops[i].err)
		}
	}
}

func TestProcessorTraceData(t *testing.T) {
	doneFn, err := obsreporttest.SetupRecordedMetricsTest()
	require.NoError(t, err)
	defer doneFn()

	const acceptedSpans = 27
	const refusedSpans = 19
	const droppedSpans = 13

	processorCtx := obsreport.ProcessorContext(context.Background(), processor)

	obsreport.ProcessorTraceDataAccepted(processorCtx, acceptedSpans)
	obsreport.ProcessorTraceDataRefused(processorCtx, refusedSpans)
	obsreport.ProcessorTraceDataDropped(processorCtx, droppedSpans)

	obsreporttest.CheckProcessorTracesViews(t, processor, acceptedSpans, refusedSpans, droppedSpans)
}

func TestProcessorMetricsData(t *testing.T) {
	doneFn, err := obsreporttest.SetupRecordedMetricsTest()
	require.NoError(t, err)
	defer doneFn()

	const acceptedPoints = 29
	const refusedPoints = 11
	const droppedPoints = 17

	processorCtx := obsreport.ProcessorContext(context.Background(), processor)
	obsreport.ProcessorMetricsDataAccepted(processorCtx, acceptedPoints)
	obsreport.ProcessorMetricsDataRefused(processorCtx, refusedPoints)
	obsreport.ProcessorMetricsDataDropped(processorCtx, droppedPoints)

	obsreporttest.CheckProcessorMetricsViews(t, processor, acceptedPoints, refusedPoints, droppedPoints)
}

func TestProcessorMetricViews(t *testing.T) {
	measures := []stats.Measure{
		stats.Int64("firstMeasure", "test firstMeasure", stats.UnitDimensionless),
		stats.Int64("secondMeasure", "test secondMeasure", stats.UnitBytes),
	}
	legacyViews := []*view.View{
		{
			Name:        measures[0].Name(),
			Description: measures[0].Description(),
			Measure:     measures[0],
			Aggregation: view.Sum(),
		},
		{
			Measure:     measures[1],
			Aggregation: view.Count(),
		},
	}

	tests := []struct {
		name       string
		withLegacy bool
		withNew    bool
		want       []*view.View
	}{
		{
			name: "none",
		},
		{
			name:       "legacy_only",
			withLegacy: true,
			want:       legacyViews,
		},
		{
			name:    "new_only",
			withNew: true,
			want: []*view.View{
				{
					Name:        "processor/test_type/" + measures[0].Name(),
					Description: measures[0].Description(),
					Measure:     measures[0],
					Aggregation: view.Sum(),
				},
				{
					Name:        "processor/test_type/" + measures[1].Name(),
					Measure:     measures[1],
					Aggregation: view.Count(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obsreport.Configure(tt.withLegacy, tt.withNew)
			got := obsreport.ProcessorMetricViews("test_type", legacyViews)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProcessorLogRecords(t *testing.T) {
	doneFn, err := obsreporttest.SetupRecordedMetricsTest()
	require.NoError(t, err)
	defer doneFn()

	const acceptedRecords = 29
	const refusedRecords = 11
	const droppedRecords = 17

	processorCtx := obsreport.ProcessorContext(context.Background(), processor)
	obsreport.ProcessorLogRecordsAccepted(processorCtx, acceptedRecords)
	obsreport.ProcessorLogRecordsRefused(processorCtx, refusedRecords)
	obsreport.ProcessorLogRecordsDropped(processorCtx, droppedRecords)

	obsreporttest.CheckProcessorLogsViews(t, processor, acceptedRecords, refusedRecords, droppedRecords)
}

type spanStore struct {
	sync.Mutex
	spans []*trace.SpanData
}

func (ss *spanStore) ExportSpan(sd *trace.SpanData) {
	ss.Lock()
	ss.spans = append(ss.spans, sd)
	ss.Unlock()
}

func (ss *spanStore) PullAllSpans() []*trace.SpanData {
	ss.Lock()
	capturedSpans := ss.spans
	ss.spans = nil
	ss.Unlock()
	return capturedSpans
}

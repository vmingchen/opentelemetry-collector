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

package testbed

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"go.uber.org/atomic"

	"go.opentelemetry.io/collector/consumer/consumerdata"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/consumer/pdatautil"
)

// MockBackend is a backend that allows receiving the data locally.
type MockBackend struct {
	// Metric and trace consumers
	tc *MockTraceConsumer
	mc *MockMetricConsumer

	receiver DataReceiver

	// Log file
	logFilePath string
	logFile     *os.File

	// Start/stop flags
	isStarted bool
	stopOnce  sync.Once
	startedAt time.Time

	// Recording fields.
	isRecording        bool
	recordMutex        sync.Mutex
	ReceivedTraces     []pdata.Traces
	ReceivedMetrics    []pdata.Metrics
	ReceivedTracesOld  []consumerdata.TraceData
	ReceivedMetricsOld []consumerdata.MetricsData
}

// NewMockBackend creates a new mock backend that receives data using specified receiver.
func NewMockBackend(logFilePath string, receiver DataReceiver) *MockBackend {
	mb := &MockBackend{
		logFilePath: logFilePath,
		receiver:    receiver,
		tc:          &MockTraceConsumer{},
		mc:          &MockMetricConsumer{},
	}
	mb.tc.backend = mb
	mb.mc.backend = mb
	return mb
}

func (mb *MockBackend) ReportFatalError(err error) {
	log.Printf("Fatal error reported: %v", err)
}

// Start a backend.
func (mb *MockBackend) Start() error {
	log.Printf("Starting mock backend...")

	var err error

	// Open log file
	mb.logFile, err = os.Create(mb.logFilePath)
	if err != nil {
		return err
	}

	err = mb.receiver.Start(mb.tc, mb.mc)
	if err != nil {
		return err
	}

	mb.isStarted = true
	mb.startedAt = time.Now()
	return nil
}

// Stop the backend
func (mb *MockBackend) Stop() {
	mb.stopOnce.Do(func() {
		if !mb.isStarted {
			return
		}

		log.Printf("Stopping mock backend...")

		mb.logFile.Close()
		mb.receiver.Stop()

		// Print stats.
		log.Printf("Stopped backend. %s", mb.GetStats())
	})
}

// EnableRecording enables recording of all data received by MockBackend.
func (mb *MockBackend) EnableRecording() {
	mb.recordMutex.Lock()
	defer mb.recordMutex.Unlock()
	mb.isRecording = true
}

func (mb *MockBackend) GetStats() string {
	received := mb.DataItemsReceived()
	return printer.Sprintf("Received:%10d items (%d/sec)", received, int(float64(received)/time.Since(mb.startedAt).Seconds()))
}

// DataItemsReceived returns total number of received spans and metrics.
func (mb *MockBackend) DataItemsReceived() uint64 {
	return mb.tc.spansReceived.Load() + mb.mc.metricsReceived.Load()
}

// ClearReceivedItems clears the list of received traces and metrics. Note: counters
// return by DataItemsReceived() are not cleared, they are cumulative.
func (mb *MockBackend) ClearReceivedItems() {
	mb.recordMutex.Lock()
	defer mb.recordMutex.Unlock()
	mb.ReceivedTraces = nil
	mb.ReceivedMetrics = nil
	mb.ReceivedTracesOld = nil
	mb.ReceivedMetricsOld = nil
}

func (mb *MockBackend) ConsumeTrace(td pdata.Traces) {
	mb.recordMutex.Lock()
	defer mb.recordMutex.Unlock()
	if mb.isRecording {
		mb.ReceivedTraces = append(mb.ReceivedTraces, td)
	}
}

// ConsumeTraceOld consumes trace data in old representation
func (mb *MockBackend) ConsumeTraceOld(td consumerdata.TraceData) {
	mb.recordMutex.Lock()
	defer mb.recordMutex.Unlock()
	if mb.isRecording {
		mb.ReceivedTracesOld = append(mb.ReceivedTracesOld, td)
	}
}

func (mb *MockBackend) ConsumeMetric(md pdata.Metrics) {
	mb.recordMutex.Lock()
	defer mb.recordMutex.Unlock()
	if mb.isRecording {
		mb.ReceivedMetrics = append(mb.ReceivedMetrics, md)
	}
}

// ConsumeMetricOld consumes metric data in old representation
func (mb *MockBackend) ConsumeMetricOld(md consumerdata.MetricsData) {
	mb.recordMutex.Lock()
	defer mb.recordMutex.Unlock()
	if mb.isRecording {
		mb.ReceivedMetricsOld = append(mb.ReceivedMetricsOld, md)
	}
}

type MockTraceConsumer struct {
	spansReceived atomic.Uint64
	backend       *MockBackend
}

func (tc *MockTraceConsumer) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
	tc.spansReceived.Add(uint64(td.SpanCount()))

	rs := td.ResourceSpans()
	for i := 0; i < rs.Len(); i++ {
		ils := rs.At(i).InstrumentationLibrarySpans()
		for j := 0; j < ils.Len(); j++ {
			spans := ils.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				var spanSeqnum int64
				var traceSeqnum int64

				seqnumAttr, ok := span.Attributes().Get("load_generator.span_seq_num")
				if ok {
					spanSeqnum = seqnumAttr.IntVal()
				}

				seqnumAttr, ok = span.Attributes().Get("load_generator.trace_seq_num")
				if ok {
					traceSeqnum = seqnumAttr.IntVal()
				}

				// Ignore the seqnums for now. We will use them later.
				_ = spanSeqnum
				_ = traceSeqnum

			}
		}
	}

	tc.backend.ConsumeTrace(td)

	return nil
}

func (tc *MockTraceConsumer) ConsumeTraceData(ctx context.Context, td consumerdata.TraceData) error {
	tc.spansReceived.Add(uint64(len(td.Spans)))

	for _, span := range td.Spans {
		var spanSeqnum int64
		var traceSeqnum int64

		if span.Attributes != nil {
			seqnumAttr, ok := span.Attributes.AttributeMap["load_generator.span_seq_num"]
			if ok {
				spanSeqnum = seqnumAttr.GetIntValue()
			}

			seqnumAttr, ok = span.Attributes.AttributeMap["load_generator.trace_seq_num"]
			if ok {
				traceSeqnum = seqnumAttr.GetIntValue()
			}

			// Ignore the seqnums for now. We will use them later.
			_ = spanSeqnum
			_ = traceSeqnum
		}

	}

	tc.backend.ConsumeTraceOld(td)

	return nil
}

type MockMetricConsumer struct {
	metricsReceived atomic.Uint64
	backend         *MockBackend
}

func (mc *MockMetricConsumer) ConsumeMetrics(_ context.Context, md pdata.Metrics) error {
	_, dataPoints := pdatautil.MetricAndDataPointCount(md)
	mc.metricsReceived.Add(uint64(dataPoints))
	mc.backend.ConsumeMetric(md)
	return nil
}

func (mc *MockMetricConsumer) ConsumeMetricsData(ctx context.Context, md consumerdata.MetricsData) error {
	dataPoints := 0
	for _, metric := range md.Metrics {
		for _, ts := range metric.Timeseries {
			dataPoints += len(ts.Points)
		}
	}

	mc.metricsReceived.Add(uint64(dataPoints))

	mc.backend.ConsumeMetricOld(md)

	return nil
}

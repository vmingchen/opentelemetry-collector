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
	"fmt"
	"log"
	"sync"
	"time"

	resourcepb "github.com/census-instrumentation/opencensus-proto/gen-go/resource/v1"
	"go.uber.org/atomic"
	"golang.org/x/text/message"

	"go.opentelemetry.io/collector/consumer/consumerdata"
)

var printer = message.NewPrinter(message.MatchLanguage("en"))

// LoadGenerator is a simple load generator.
type LoadGenerator struct {
	sender DataSender

	dataProvider DataProvider

	// Number of batches of data items sent.
	batchesSent atomic.Uint64

	// Number of data items (spans or metric data points) sent.
	dataItemsSent atomic.Uint64

	stopOnce   sync.Once
	stopWait   sync.WaitGroup
	stopSignal chan struct{}

	options LoadOptions

	// Record information about previous errors to avoid flood of error messages.
	prevErr error
}

// LoadOptions defines the options to use for generating the load.
type LoadOptions struct {
	// DataItemsPerSecond specifies how many spans or metric data points to generate each second.
	DataItemsPerSecond int

	// ItemsPerBatch specifies how many spans or metric data points per batch to generate.
	// Should be greater than zero. The number of batches generated per second will be
	// DataItemsPerSecond/ItemsPerBatch.
	ItemsPerBatch int

	// Attributes to add to each generated data item. Can be empty.
	Attributes map[string]string
}

// NewLoadGenerator creates a load generator that sends data using specified sender.
func NewLoadGenerator(dataProvider DataProvider, sender DataSender) (*LoadGenerator, error) {
	if sender == nil {
		return nil, fmt.Errorf("cannot create load generator without DataSender")
	}

	lg := &LoadGenerator{
		stopSignal:   make(chan struct{}),
		sender:       sender,
		dataProvider: dataProvider,
	}

	return lg, nil
}

// Start the load.
func (lg *LoadGenerator) Start(options LoadOptions) {
	lg.options = options

	if lg.options.ItemsPerBatch == 0 {
		// 10 items per batch by default.
		lg.options.ItemsPerBatch = 10
	}

	log.Printf("Starting load generator at %d items/sec.", lg.options.DataItemsPerSecond)

	// Indicate that generation is in progress.
	lg.stopWait.Add(1)

	// Begin generation
	go lg.generate()
}

// Stop the load.
func (lg *LoadGenerator) Stop() {
	lg.stopOnce.Do(func() {
		// Signal generate() to stop.
		close(lg.stopSignal)

		// Wait for it to stop.
		lg.stopWait.Wait()

		// Print stats.
		log.Printf("Stopped generator. %s", lg.GetStats())
	})
}

// GetStats returns the stats as a printable string.
func (lg *LoadGenerator) GetStats() string {
	return printer.Sprintf("Sent:%10d items", lg.DataItemsSent())
}

func (lg *LoadGenerator) DataItemsSent() uint64 {
	return lg.dataItemsSent.Load()
}

// IncDataItemsSent is used when a test bypasses the LoadGenerator and sends data
// directly via TestCases's Sender. This is necessary so that the total number of sent
// items in the end is correct, because the reports are printed from LoadGenerator's
// fields. This is not the best way, a better approach would be to refactor the
// reports to use their own counter and load generator and other sending sources
// to contribute to this counter. This could be done as a future improvement.
func (lg *LoadGenerator) IncDataItemsSent() {
	lg.dataItemsSent.Inc()
}

func (lg *LoadGenerator) generate() {
	// Indicate that generation is done at the end
	defer lg.stopWait.Done()

	if lg.options.DataItemsPerSecond == 0 {
		return
	}

	lg.dataProvider.SetLoadGeneratorCounters(&lg.batchesSent, &lg.dataItemsSent)

	err := lg.sender.Start()
	if err != nil {
		log.Printf("Cannot start sender: %v", err)
		return
	}

	t := time.NewTicker(time.Second / time.Duration(lg.options.DataItemsPerSecond/lg.options.ItemsPerBatch))
	defer t.Stop()
	done := false
	for !done {
		select {
		case <-t.C:
			switch lg.sender.(type) {
			case TraceDataSender:
				lg.generateTrace()
			case TraceDataSenderOld:
				lg.generateTraceOld()
			case MetricDataSender:
				lg.generateMetrics()
			case MetricDataSenderOld:
				lg.generateMetricsOld()
			default:
				log.Printf("Invalid type of LoadGenerator sender")
			}

		case <-lg.stopSignal:
			done = true
		}
	}
	// Send all pending generated data.
	lg.sender.Flush()
}

func (lg *LoadGenerator) generateTrace() {
	traceSender := lg.sender.(TraceDataSender)

	traceData, done := lg.dataProvider.GenerateTraces()
	if done {
		return
	}

	err := traceSender.SendSpans(traceData)
	if err == nil {
		lg.prevErr = nil
	} else if lg.prevErr == nil || lg.prevErr.Error() != err.Error() {
		lg.prevErr = err
		log.Printf("Cannot send traces: %v", err)
	}
}

func (lg *LoadGenerator) generateTraceOld() {
	traceSender := lg.sender.(TraceDataSenderOld)

	spans, done := lg.dataProvider.GenerateTracesOld()
	if done {
		return
	}
	traceData := consumerdata.TraceData{
		Spans: spans,
	}

	err := traceSender.SendSpans(traceData)
	if err == nil {
		lg.prevErr = nil
	} else if lg.prevErr == nil || lg.prevErr.Error() != err.Error() {
		lg.prevErr = err
		log.Printf("Cannot send traces: %v", err)
	}
}

func (lg *LoadGenerator) generateMetrics() {
	metricSender := lg.sender.(MetricDataSender)

	metricData, done := lg.dataProvider.GenerateMetrics()
	if done {
		return
	}

	err := metricSender.SendMetrics(metricData)
	if err == nil {
		lg.prevErr = nil
	} else if lg.prevErr == nil || lg.prevErr.Error() != err.Error() {
		lg.prevErr = err
		log.Printf("Cannot send metrics: %v", err)
	}
}

func (lg *LoadGenerator) generateMetricsOld() {
	metricSender := lg.sender.(MetricDataSenderOld)

	resource := &resourcepb.Resource{
		Labels: lg.options.Attributes,
	}
	metrics, done := lg.dataProvider.GenerateMetricsOld()
	if done {
		return
	}
	metricData := consumerdata.MetricsData{
		Resource: resource,
		Metrics:  metrics,
	}

	err := metricSender.SendMetrics(metricData)
	if err == nil {
		lg.prevErr = nil
	} else if lg.prevErr == nil || lg.prevErr.Error() != err.Error() {
		lg.prevErr = err
		log.Printf("Cannot send metrics: %v", err)
	}
}

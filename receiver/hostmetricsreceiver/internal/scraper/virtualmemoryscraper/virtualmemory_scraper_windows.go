// Copyright 2020, OpenTelemetry Authors
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

// +build windows

package virtualmemoryscraper

import (
	"context"
	"math"
	"time"

	"go.opencensus.io/trace"

	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/windows/pdh"
)

const (
	pageReadsPerSecPath  = `\Memory\Page Reads/sec`
	pageWritesperSecPath = `\Memory\Page Writes/sec`
)

// scraper for VirtualMemory Metrics
type scraper struct {
	config *Config

	pageReadsPerSecCounter  pdh.PerfCounterScraper
	pageWritesPerSecCounter pdh.PerfCounterScraper

	startTime            pdata.TimestampUnixNano
	prevPagingScrapeTime time.Time
	cumulativePageReads  float64
	cumulativePageWrites float64
}

// newVirtualMemoryScraper creates a set of VirtualMemory related metrics
func newVirtualMemoryScraper(_ context.Context, cfg *Config) *scraper {
	return &scraper{config: cfg}
}

// Initialize
func (s *scraper) Initialize(_ context.Context) error {
	s.startTime = pdata.TimestampUnixNano(uint64(time.Now().UnixNano()))
	s.prevPagingScrapeTime = time.Now()

	var err error

	s.pageReadsPerSecCounter, err = pdh.NewPerfCounter(pageReadsPerSecPath, true)
	if err != nil {
		return err
	}

	s.pageWritesPerSecCounter, err = pdh.NewPerfCounter(pageWritesperSecPath, true)
	if err != nil {
		return err
	}

	return nil
}

// Close
func (s *scraper) Close(_ context.Context) error {
	var errors []error

	err := s.pageReadsPerSecCounter.Close()
	if err != nil {
		errors = append(errors, err)
	}

	err = s.pageWritesPerSecCounter.Close()
	if err != nil {
		errors = append(errors, err)
	}

	return componenterror.CombineErrors(errors)
}

// ScrapeMetrics
func (s *scraper) ScrapeMetrics(ctx context.Context) (pdata.MetricSlice, error) {
	_, span := trace.StartSpan(ctx, "virtualmemoryscraper.ScrapeMetrics")
	defer span.End()

	metrics := pdata.NewMetricSlice()

	var errors []error

	err := s.scrapeAndAppendSwapUsageMetric(metrics)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.scrapeAndAppendPagingMetric(metrics)
	if err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return metrics, componenterror.CombineErrors(errors)
	}

	return metrics, nil
}

var getPageFileStats = getPageFileStatsInternal

func (s *scraper) scrapeAndAppendSwapUsageMetric(metrics pdata.MetricSlice) error {
	pageFiles, err := getPageFileStats()
	if err != nil {
		return err
	}

	idx := metrics.Len()
	metrics.Resize(idx + 1)
	initializeSwapUsageMetric(metrics.At(idx), pageFiles)
	return nil
}

func initializeSwapUsageMetric(metric pdata.Metric, pageFiles []*pageFileData) {
	swapUsageDescriptor.CopyTo(metric.MetricDescriptor())

	idps := metric.Int64DataPoints()
	idps.Resize(2 * len(pageFiles))

	idx := 0
	for _, pageFile := range pageFiles {
		initializeSwapUsageDataPoint(idps.At(idx+0), pageFile.name, usedLabelValue, int64(pageFile.used))
		initializeSwapUsageDataPoint(idps.At(idx+1), pageFile.name, freeLabelValue, int64(pageFile.total-pageFile.used))
		idx += 2
	}
}

func initializeSwapUsageDataPoint(dataPoint pdata.Int64DataPoint, deviceLabel string, stateLabel string, value int64) {
	labelsMap := dataPoint.LabelsMap()
	labelsMap.Insert(deviceLabelName, deviceLabel)
	labelsMap.Insert(stateLabelName, stateLabel)
	dataPoint.SetTimestamp(pdata.TimestampUnixNano(uint64(time.Now().UnixNano())))
	dataPoint.SetValue(value)
}

func (s *scraper) scrapeAndAppendPagingMetric(metrics pdata.MetricSlice) error {
	now := time.Now()
	durationSinceLastScraped := now.Sub(s.prevPagingScrapeTime).Seconds()
	s.prevPagingScrapeTime = now

	pageReadsPerSecValues, err := s.pageReadsPerSecCounter.ScrapeData()
	if err != nil {
		return err
	}

	pageWritesPerSecValues, err := s.pageWritesPerSecCounter.ScrapeData()
	if err != nil {
		return err
	}

	s.cumulativePageReads += (pageReadsPerSecValues[0].Value * durationSinceLastScraped)
	s.cumulativePageWrites += (pageWritesPerSecValues[0].Value * durationSinceLastScraped)

	idx := metrics.Len()
	metrics.Resize(idx + 1)
	initializePagingMetric(metrics.At(idx), s.startTime, s.cumulativePageReads, s.cumulativePageWrites)
	return nil
}

func initializePagingMetric(metric pdata.Metric, startTime pdata.TimestampUnixNano, reads float64, writes float64) {
	swapPagingDescriptor.CopyTo(metric.MetricDescriptor())

	idps := metric.Int64DataPoints()
	idps.Resize(2)
	initializePagingDataPoint(idps.At(0), startTime, inDirectionLabelValue, reads)
	initializePagingDataPoint(idps.At(1), startTime, outDirectionLabelValue, writes)
}

func initializePagingDataPoint(dataPoint pdata.Int64DataPoint, startTime pdata.TimestampUnixNano, directionLabel string, value float64) {
	labelsMap := dataPoint.LabelsMap()
	labelsMap.Insert(typeLabelName, majorTypeLabelValue)
	labelsMap.Insert(directionLabelName, directionLabel)
	dataPoint.SetStartTime(startTime)
	dataPoint.SetTimestamp(pdata.TimestampUnixNano(uint64(time.Now().UnixNano())))
	dataPoint.SetValue(int64(math.Round(value)))
}

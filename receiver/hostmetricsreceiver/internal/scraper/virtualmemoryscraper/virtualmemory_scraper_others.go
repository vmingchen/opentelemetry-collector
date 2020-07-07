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

// +build !windows

package virtualmemoryscraper

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"go.opencensus.io/trace"

	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer/pdata"
)

// scraper for VirtualMemory Metrics
type scraper struct {
	config    *Config
	startTime pdata.TimestampUnixNano
}

// newVirtualMemoryScraper creates a VirtualMemory Scraper
func newVirtualMemoryScraper(_ context.Context, cfg *Config) *scraper {
	return &scraper{config: cfg}
}

// Initialize
func (s *scraper) Initialize(_ context.Context) error {
	bootTime, err := host.BootTime()
	if err != nil {
		return err
	}

	s.startTime = pdata.TimestampUnixNano(bootTime)
	return nil
}

// Close
func (s *scraper) Close(_ context.Context) error {
	return nil
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

	err = s.scrapeAndAppendPagingMetrics(metrics)
	if err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return metrics, componenterror.CombineErrors(errors)
	}

	return metrics, nil
}

var getVirtualMemory = mem.VirtualMemory

func (s *scraper) scrapeAndAppendSwapUsageMetric(metrics pdata.MetricSlice) error {
	vmem, err := getVirtualMemory()
	if err != nil {
		return err
	}

	idx := metrics.Len()
	metrics.Resize(idx + 1)
	initializeSwapUsageMetric(metrics.At(idx), vmem)
	return nil
}

func initializeSwapUsageMetric(metric pdata.Metric, vmem *mem.VirtualMemoryStat) {
	swapUsageDescriptor.CopyTo(metric.MetricDescriptor())

	idps := metric.Int64DataPoints()
	idps.Resize(3)
	initializeSwapUsageDataPoint(idps.At(0), usedLabelValue, int64(vmem.SwapTotal-vmem.SwapFree-vmem.SwapCached))
	initializeSwapUsageDataPoint(idps.At(1), freeLabelValue, int64(vmem.SwapFree))
	initializeSwapUsageDataPoint(idps.At(2), cachedLabelValue, int64(vmem.SwapCached))
}

func initializeSwapUsageDataPoint(dataPoint pdata.Int64DataPoint, stateLabel string, value int64) {
	labelsMap := dataPoint.LabelsMap()
	labelsMap.Insert(stateLabelName, stateLabel)
	dataPoint.SetTimestamp(pdata.TimestampUnixNano(uint64(time.Now().UnixNano())))
	dataPoint.SetValue(value)
}

var getSwapMemory = mem.SwapMemory

func (s *scraper) scrapeAndAppendPagingMetrics(metrics pdata.MetricSlice) error {
	swap, err := getSwapMemory()
	if err != nil {
		return err
	}

	idx := metrics.Len()
	metrics.Resize(idx + 2)
	initializePagingMetric(metrics.At(idx+0), s.startTime, swap)
	initializePageFaultsMetric(metrics.At(idx+1), s.startTime, swap)
	return nil
}

func initializePagingMetric(metric pdata.Metric, startTime pdata.TimestampUnixNano, swap *mem.SwapMemoryStat) {
	swapPagingDescriptor.CopyTo(metric.MetricDescriptor())

	idps := metric.Int64DataPoints()
	idps.Resize(4)
	initializePagingDataPoint(idps.At(0), startTime, majorTypeLabelValue, inDirectionLabelValue, int64(swap.Sin))
	initializePagingDataPoint(idps.At(1), startTime, majorTypeLabelValue, outDirectionLabelValue, int64(swap.Sout))
	initializePagingDataPoint(idps.At(2), startTime, minorTypeLabelValue, inDirectionLabelValue, int64(swap.PgIn))
	initializePagingDataPoint(idps.At(3), startTime, minorTypeLabelValue, outDirectionLabelValue, int64(swap.PgOut))
}

func initializePagingDataPoint(dataPoint pdata.Int64DataPoint, startTime pdata.TimestampUnixNano, typeLabel string, directionLabel string, value int64) {
	labelsMap := dataPoint.LabelsMap()
	labelsMap.Insert(typeLabelName, typeLabel)
	labelsMap.Insert(directionLabelName, directionLabel)
	dataPoint.SetStartTime(startTime)
	dataPoint.SetTimestamp(pdata.TimestampUnixNano(uint64(time.Now().UnixNano())))
	dataPoint.SetValue(value)
}

func initializePageFaultsMetric(metric pdata.Metric, startTime pdata.TimestampUnixNano, swap *mem.SwapMemoryStat) {
	swapPageFaultsDescriptor.CopyTo(metric.MetricDescriptor())

	idps := metric.Int64DataPoints()
	idps.Resize(1)
	initializePageFaultDataPoint(idps.At(0), startTime, minorTypeLabelValue, int64(swap.PgFault))
	// TODO add swap.PgMajFault once available in gopsutil
}

func initializePageFaultDataPoint(dataPoint pdata.Int64DataPoint, startTime pdata.TimestampUnixNano, typeLabel string, value int64) {
	dataPoint.LabelsMap().Insert(typeLabelName, typeLabel)
	dataPoint.SetStartTime(startTime)
	dataPoint.SetTimestamp(pdata.TimestampUnixNano(uint64(time.Now().UnixNano())))
	dataPoint.SetValue(value)
}

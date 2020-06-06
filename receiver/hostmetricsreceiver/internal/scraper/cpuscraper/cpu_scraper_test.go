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

package cpuscraper

import (
	"context"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver/internal"
)

type validationFn func(*testing.T, pdata.MetricSlice)

func TestScrapeMetrics_MinimalData(t *testing.T) {
	createScraperAndValidateScrapedMetrics(t, &Config{}, func(t *testing.T, metrics pdata.MetricSlice) {
		// expect 1 metric
		assert.Equal(t, 1, metrics.Len())

		// for cpu seconds metric, expect a datapoint for each state label, including at least 4 standard states
		hostCPUTimeMetric := metrics.At(0)
		internal.AssertDescriptorEqual(t, metricCPUSecondsDescriptor, hostCPUTimeMetric.MetricDescriptor())
		assert.GreaterOrEqual(t, hostCPUTimeMetric.DoubleDataPoints().Len(), 4)
		internal.AssertDoubleMetricLabelDoesNotExist(t, hostCPUTimeMetric, 0, cpuLabelName)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 0, stateLabelName, userStateLabelValue)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 1, stateLabelName, systemStateLabelValue)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 2, stateLabelName, idleStateLabelValue)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 3, stateLabelName, interruptStateLabelValue)
	})
}

func TestScrapeMetrics_AllData(t *testing.T) {
	config := &Config{
		ReportPerCPU: true,
	}

	createScraperAndValidateScrapedMetrics(t, config, func(t *testing.T, metrics pdata.MetricSlice) {
		// expect 1 metric
		assert.Equal(t, 1, metrics.Len())

		// for cpu seconds metric, expect a datapoint for each state label & core combination with at least 4 standard states
		hostCPUTimeMetric := metrics.At(0)
		internal.AssertDescriptorEqual(t, metricCPUSecondsDescriptor, hostCPUTimeMetric.MetricDescriptor())
		assert.GreaterOrEqual(t, hostCPUTimeMetric.DoubleDataPoints().Len(), runtime.NumCPU()*4)
		internal.AssertDoubleMetricLabelExists(t, hostCPUTimeMetric, 0, cpuLabelName)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 0, stateLabelName, userStateLabelValue)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 1, stateLabelName, systemStateLabelValue)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 2, stateLabelName, idleStateLabelValue)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 3, stateLabelName, interruptStateLabelValue)
	})
}

func TestScrapeMetrics_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		return
	}

	createScraperAndValidateScrapedMetrics(t, &Config{}, func(t *testing.T, metrics pdata.MetricSlice) {
		// for cpu seconds metric, expect a datapoint for all 8 state labels
		hostCPUTimeMetric := metrics.At(0)
		internal.AssertDescriptorEqual(t, metricCPUSecondsDescriptor, hostCPUTimeMetric.MetricDescriptor())
		assert.Equal(t, 8, hostCPUTimeMetric.DoubleDataPoints().Len())
		internal.AssertDoubleMetricLabelDoesNotExist(t, hostCPUTimeMetric, 0, cpuLabelName)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 0, stateLabelName, userStateLabelValue)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 1, stateLabelName, systemStateLabelValue)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 2, stateLabelName, idleStateLabelValue)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 3, stateLabelName, interruptStateLabelValue)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 4, stateLabelName, niceStateLabelValue)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 5, stateLabelName, softIRQStateLabelValue)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 6, stateLabelName, stealStateLabelValue)
		internal.AssertDoubleMetricLabelHasValue(t, hostCPUTimeMetric, 7, stateLabelName, waitStateLabelValue)
	})
}

func createScraperAndValidateScrapedMetrics(t *testing.T, config *Config, assertFn validationFn) {
	scraper := newCPUScraper(context.Background(), config)
	err := scraper.Initialize(context.Background())
	require.NoError(t, err, "Failed to initialize cpu scraper: %v", err)
	defer func() { assert.NoError(t, scraper.Close(context.Background())) }()

	metrics, err := scraper.ScrapeMetrics(context.Background())
	require.NoError(t, err, "Failed to scrape metrics: %v", err)

	assertFn(t, metrics)
}

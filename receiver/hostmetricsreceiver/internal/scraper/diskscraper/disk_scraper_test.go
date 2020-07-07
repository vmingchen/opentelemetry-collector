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

package diskscraper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver/internal"
)

type validationFn func(*testing.T, pdata.MetricSlice)

func TestScrapeMetrics(t *testing.T) {
	createScraperAndValidateScrapedMetrics(t, &Config{}, func(t *testing.T, metrics pdata.MetricSlice) {
		// expect 3 metrics
		assert.Equal(t, 3, metrics.Len())

		// for each disk metric, expect a read & write datapoint for at least one drive
		assertDiskMetricMatchesDescriptorAndHasReadAndWriteDataPoints(t, metrics.At(0), diskIODescriptor)
		assertDiskMetricMatchesDescriptorAndHasReadAndWriteDataPoints(t, metrics.At(1), diskOpsDescriptor)
		assertDiskMetricMatchesDescriptorAndHasReadAndWriteDataPoints(t, metrics.At(2), diskTimeDescriptor)
	})
}

func assertDiskMetricMatchesDescriptorAndHasReadAndWriteDataPoints(t *testing.T, metric pdata.Metric, expectedDescriptor pdata.MetricDescriptor) {
	internal.AssertDescriptorEqual(t, expectedDescriptor, metric.MetricDescriptor())
	assert.GreaterOrEqual(t, metric.Int64DataPoints().Len(), 2)
	internal.AssertInt64MetricLabelHasValue(t, metric, 0, directionLabelName, readDirectionLabelValue)
	internal.AssertInt64MetricLabelHasValue(t, metric, 1, directionLabelName, writeDirectionLabelValue)
}

func createScraperAndValidateScrapedMetrics(t *testing.T, config *Config, assertFn validationFn) {
	scraper := newDiskScraper(context.Background(), config)
	err := scraper.Initialize(context.Background())
	require.NoError(t, err, "Failed to initialize disk scraper: %v", err)
	defer func() { assert.NoError(t, scraper.Close(context.Background())) }()

	metrics, err := scraper.ScrapeMetrics(context.Background())
	require.NoError(t, err, "Failed to scrape metrics: %v", err)

	assertFn(t, metrics)
}

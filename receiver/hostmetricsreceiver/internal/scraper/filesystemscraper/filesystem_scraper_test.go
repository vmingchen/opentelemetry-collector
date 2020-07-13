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

package filesystemscraper

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/shirou/gopsutil/disk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver/internal"
)

func TestScrapeMetrics(t *testing.T) {
	type testCase struct {
		name           string
		partitionsFunc func(bool) ([]disk.PartitionStat, error)
		usageFunc      func(string) (*disk.UsageStat, error)
		expectedErr    string
	}

	testCases := []testCase{
		{
			name: "Standard",
		},
		{
			name:           "Partitions Error",
			partitionsFunc: func(bool) ([]disk.PartitionStat, error) { return nil, errors.New("err1") },
			expectedErr:    "err1",
		},
		{
			name:        "Usage Error",
			usageFunc:   func(string) (*disk.UsageStat, error) { return nil, errors.New("err2") },
			expectedErr: "err2",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			scraper := newFileSystemScraper(context.Background(), &Config{})
			if test.partitionsFunc != nil {
				scraper.partitions = test.partitionsFunc
			}
			if test.usageFunc != nil {
				scraper.usage = test.usageFunc
			}

			err := scraper.Initialize(context.Background())
			require.NoError(t, err, "Failed to initialize file system scraper: %v", err)
			defer func() { assert.NoError(t, scraper.Close(context.Background())) }()

			metrics, err := scraper.ScrapeMetrics(context.Background())
			if test.expectedErr != "" {
				assert.Contains(t, err.Error(), test.expectedErr)
				return
			}
			require.NoError(t, err, "Failed to scrape metrics: %v", err)

			assert.GreaterOrEqual(t, metrics.Len(), 1)

			assertFileSystemUsageMetricValid(t, metrics.At(0), fileSystemUsageDescriptor)

			if isUnix() {
				assertFileSystemUsageMetricHasUnixSpecificStateLabels(t, metrics.At(0))
				assertFileSystemUsageMetricValid(t, metrics.At(1), fileSystemINodesUsageDescriptor)
			}
		})
	}
}

func assertFileSystemUsageMetricValid(t *testing.T, metric pdata.Metric, descriptor pdata.MetricDescriptor) {
	internal.AssertDescriptorEqual(t, descriptor, metric.MetricDescriptor())
	assert.GreaterOrEqual(t, metric.Int64DataPoints().Len(), 2)
	internal.AssertInt64MetricLabelHasValue(t, metric, 0, stateLabelName, usedLabelValue)
	internal.AssertInt64MetricLabelHasValue(t, metric, 1, stateLabelName, freeLabelValue)
}

func assertFileSystemUsageMetricHasUnixSpecificStateLabels(t *testing.T, metric pdata.Metric) {
	internal.AssertInt64MetricLabelHasValue(t, metric, 2, stateLabelName, reservedLabelValue)
}

func isUnix() bool {
	for _, unixOS := range []string{"linux", "darwin", "freebsd", "openbsd", "solaris"} {
		if runtime.GOOS == unixOS {
			return true
		}
	}

	return false
}

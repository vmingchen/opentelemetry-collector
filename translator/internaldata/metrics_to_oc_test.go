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

package internaldata

import (
	"testing"
	"time"

	occommon "github.com/census-instrumentation/opencensus-proto/gen-go/agent/common/v1"
	ocmetrics "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	ocresource "github.com/census-instrumentation/opencensus-proto/gen-go/resource/v1"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/collector/consumer/consumerdata"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/internal/data"
	"go.opentelemetry.io/collector/internal/data/testdata"
	"go.opentelemetry.io/collector/translator/conventions"
)

func TestMetricsDataToOC(t *testing.T) {

	sampleMetricData := testdata.GenerateMetricDataWithCountersHistogramAndSummary()
	attrs := sampleMetricData.ResourceMetrics().At(0).Resource().Attributes()
	attrs.Upsert(conventions.AttributeHostHostname, pdata.NewAttributeValueString("host1"))
	attrs.Upsert(conventions.OCAttributeProcessID, pdata.NewAttributeValueInt(123))
	attrs.Upsert(conventions.OCAttributeProcessStartTime, pdata.NewAttributeValueString("2020-02-11T20:26:00Z"))
	attrs.Upsert(conventions.AttributeTelemetrySDKLanguage, pdata.NewAttributeValueString("CPP"))
	attrs.Upsert(conventions.AttributeTelemetrySDKVersion, pdata.NewAttributeValueString("v2.0.1"))
	attrs.Upsert(conventions.OCAttributeExporterVersion, pdata.NewAttributeValueString("v1.2.0"))

	tests := []struct {
		name     string
		internal data.MetricData
		oc       []consumerdata.MetricsData
	}{
		{
			name:     "empty",
			internal: testdata.GenerateMetricDataEmpty(),
			oc:       []consumerdata.MetricsData(nil),
		},

		{
			name:     "one-empty-resource-metrics",
			internal: testdata.GenerateMetricDataOneEmptyResourceMetrics(),
			oc: []consumerdata.MetricsData{
				{},
			},
		},

		{
			name:     "one-empty-one-nil-resource-metrics",
			internal: testdata.GenerateMetricDataOneEmptyOneNilResourceMetrics(),
			oc: []consumerdata.MetricsData{
				{},
			},
		},

		{
			name:     "no-libraries",
			internal: testdata.GenerateMetricDataNoLibraries(),
			oc: []consumerdata.MetricsData{
				generateOCTestDataNoMetrics(),
			},
		},

		{
			name:     "one-empty-instrumentation-library",
			internal: testdata.GenerateMetricDataOneEmptyInstrumentationLibrary(),
			oc: []consumerdata.MetricsData{
				generateOCTestDataNoMetrics(),
			},
		},

		{
			name:     "one-empty-one-nil-instrumentation-library",
			internal: testdata.GenerateMetricDataOneEmptyOneNilInstrumentationLibrary(),
			oc: []consumerdata.MetricsData{
				generateOCTestDataNoMetrics(),
			},
		},

		{
			name:     "one-metric-no-resource",
			internal: testdata.GenerateMetricDataOneMetricNoResource(),
			oc: []consumerdata.MetricsData{
				{
					Metrics: []*ocmetrics.Metric{
						generateOCTestMetricInt(),
					},
				},
			},
		},

		{
			name:     "one-metric",
			internal: testdata.GenerateMetricDataOneMetric(),
			oc: []consumerdata.MetricsData{
				generateOCTestDataMetricsOneMetric(),
			},
		},

		{
			name:     "one-metric-one-nil",
			internal: testdata.GenerateMetricDataOneMetricOneNil(),
			oc: []consumerdata.MetricsData{
				generateOCTestDataMetricsOneMetric(),
			},
		},

		{
			name:     "one-metric-one-nil-point",
			internal: testdata.GenerateMetricDataOneMetricOneNilPoint(),
			oc: []consumerdata.MetricsData{
				generateOCTestDataMetricsOneMetric(),
			},
		},

		{
			name:     "one-metric-no-labels",
			internal: testdata.GenerateMetricDataOneMetricNoLabels(),
			oc: []consumerdata.MetricsData{
				generateOCTestDataNoLabels(),
			},
		},

		{
			name:     "all-types-no-data-points",
			internal: testdata.GenerateMetricDataAllTypesNoDataPoints(),
			oc: []consumerdata.MetricsData{
				generateOCTestDataNoPoints(),
			},
		},

		{
			name:     "sample-metric",
			internal: sampleMetricData,
			oc: []consumerdata.MetricsData{
				generateOCTestData(t),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := MetricDataToOC(test.internal)
			assert.EqualValues(t, test.oc, got)
		})
	}
}

func generateOCTestData(t *testing.T) consumerdata.MetricsData {
	ts, err := ptypes.TimestampProto(time.Date(2020, 2, 11, 20, 26, 0, 0, time.UTC))
	assert.NoError(t, err)

	return consumerdata.MetricsData{
		Node: &occommon.Node{
			Identifier: &occommon.ProcessIdentifier{
				HostName:       "host1",
				Pid:            123,
				StartTimestamp: ts,
			},
			LibraryInfo: &occommon.LibraryInfo{
				Language:           occommon.LibraryInfo_CPP,
				ExporterVersion:    "v1.2.0",
				CoreLibraryVersion: "v2.0.1",
			},
		},
		Resource: &ocresource.Resource{
			Labels: map[string]string{
				"resource-attr": "resource-attr-val-1",
			},
		},
		Metrics: []*ocmetrics.Metric{
			generateOCTestMetricInt(),
			generateOCTestMetricDouble(),
			generateOCTestMetricHistogram(),
			generateOCTestMetricSummary(),
		},
	}
}

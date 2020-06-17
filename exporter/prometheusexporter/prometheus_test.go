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

package prometheusexporter

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/consumer/consumerdata"
)

func TestPrometheusExporter(t *testing.T) {
	tests := []struct {
		config  *Config
		wantErr string
	}{
		{
			config: &Config{
				Namespace: "test",
				ConstLabels: map[string]string{
					"foo":  "bar",
					"code": "one",
				},
				Endpoint: ":8999",
			},
		},
		{
			config:  &Config{},
			wantErr: "expecting a non-blank address to run the Prometheus metrics handler",
		},
	}

	factory := Factory{}
	for _, tt := range tests {
		// Run it a few times to ensure that shutdowns exit cleanly.
		for j := 0; j < 3; j++ {
			consumer, err := factory.CreateMetricsExporter(zap.NewNop(), tt.config)

			if tt.wantErr != "" {
				require.Equal(t, tt.wantErr, err.Error())
				continue
			}

			assert.NotNil(t, consumer)

			require.Nil(t, err)
			require.NoError(t, consumer.Shutdown(context.Background()))
		}
	}
}

func TestPrometheusExporter_endToEnd(t *testing.T) {
	config := &Config{
		Namespace: "test",
		ConstLabels: map[string]string{
			"foo":  "bar",
			"code": "one",
		},
		Endpoint: ":7777",
	}

	factory := Factory{}
	consumer, err := factory.CreateMetricsExporter(zap.NewNop(), config)
	assert.NoError(t, err)

	defer consumer.Shutdown(context.Background())

	assert.NotNil(t, consumer)

	for delta := 0; delta <= 20; delta += 10 {
		consumer.ConsumeMetricsData(context.Background(), consumerdata.MetricsData{Metrics: metricBuilder(int64(delta))})

		res, err := http.Get("http://localhost:7777/metrics")
		if err != nil {
			t.Fatalf("Failed to perform a scrape: %v", err)
		}
		if g, w := res.StatusCode, 200; g != w {
			t.Errorf("Mismatched HTTP response status code: Got: %d Want: %d", g, w)
		}
		blob, _ := ioutil.ReadAll(res.Body)
		_ = res.Body.Close()
		want := []string{
			`# HELP test_this_one_there_where_ Extra ones`,
			`# TYPE test_this_one_there_where_ counter`,
			fmt.Sprintf(`test_this_one_there_where_{arch="x86",code="one",foo="bar",os="windows"} %v`, 99+delta),
			fmt.Sprintf(`test_this_one_there_where_{arch="x86",code="one",foo="bar",os="linux"} %v`, 100+delta),
		}

		for _, w := range want {
			if !strings.Contains(string(blob), w) {
				t.Errorf("Missing %v from response:\n%v", w, string(blob))
			}
		}
	}
}

func metricBuilder(delta int64) []*metricspb.Metric {
	return []*metricspb.Metric{
		{
			MetricDescriptor: &metricspb.MetricDescriptor{
				Name:        "this/one/there(where)",
				Description: "Extra ones",
				Unit:        "1",
				Type:        metricspb.MetricDescriptor_CUMULATIVE_INT64,
				LabelKeys: []*metricspb.LabelKey{
					{Key: "os", Description: "Operating system"},
					{Key: "arch", Description: "Architecture"},
				},
			},
			Timeseries: []*metricspb.TimeSeries{
				{
					StartTimestamp: &timestamp.Timestamp{
						Seconds: 1543160298,
						Nanos:   100000090,
					},
					LabelValues: []*metricspb.LabelValue{
						{Value: "windows"},
						{Value: "x86"},
					},
					Points: []*metricspb.Point{
						{
							Timestamp: &timestamp.Timestamp{
								Seconds: 1543160298,
								Nanos:   100000997,
							},
							Value: &metricspb.Point_Int64Value{
								Int64Value: 99 + delta,
							},
						},
					},
				},
			},
		},
		{
			MetricDescriptor: &metricspb.MetricDescriptor{
				Name:        "this/one/there(where)",
				Description: "Extra ones",
				Unit:        "1",
				Type:        metricspb.MetricDescriptor_CUMULATIVE_INT64,
				LabelKeys: []*metricspb.LabelKey{
					{Key: "os", Description: "Operating system"},
					{Key: "arch", Description: "Architecture"},
				},
			},
			Timeseries: []*metricspb.TimeSeries{
				{
					StartTimestamp: &timestamp.Timestamp{
						Seconds: 1543160298,
						Nanos:   100000090,
					},
					LabelValues: []*metricspb.LabelValue{
						{Value: "linux"},
						{Value: "x86"},
					},
					Points: []*metricspb.Point{
						{
							Timestamp: &timestamp.Timestamp{
								Seconds: 1543160298,
								Nanos:   100000997,
							},
							Value: &metricspb.Point_Int64Value{
								Int64Value: 100 + delta,
							},
						},
					},
				},
			},
		},
	}
}

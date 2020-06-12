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

package zipkinreceiver

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	commonpb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/common/v1"
	tracepb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumerdata"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/exporter/zipkinexporter"
	"go.opentelemetry.io/collector/internal"
	"go.opentelemetry.io/collector/testutils"
	tracetranslator "go.opentelemetry.io/collector/translator/trace"
	"go.opentelemetry.io/collector/translator/trace/zipkin"
)

const zipkinReceiver = "zipkin_receiver_test"

func TestShortIDSpanConversion(t *testing.T) {
	shortID, _ := zipkinmodel.TraceIDFromHex("0102030405060708")
	assert.Equal(t, uint64(0), shortID.High, "wanted 64bit traceID, so TraceID.High must be zero")

	zc := zipkinmodel.SpanContext{
		TraceID: shortID,
		ID:      zipkinmodel.ID(shortID.Low),
	}
	zs := zipkinmodel.SpanModel{
		SpanContext: zc,
	}

	ocSpan, _ := zipkinSpanToTraceSpan(&zs)
	require.Len(t, ocSpan.TraceId, 16, "incorrect OC proto trace id length")

	want := []byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8}
	assert.Equal(t, want, ocSpan.TraceId)
}

func TestNew(t *testing.T) {
	type args struct {
		address      string
		nextConsumer consumer.TraceConsumerOld
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name:    "nil nextConsumer",
			args:    args{},
			wantErr: componenterror.ErrNilNextConsumer,
		},
		{
			name: "happy path",
			args: args{
				nextConsumer: exportertest.NewNopTraceExporterOld(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(zipkinReceiver, tt.args.address, tt.args.nextConsumer)
			require.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				require.NotNil(t, got)
			} else {
				require.Nil(t, got)
			}
		})
	}
}

func TestZipkinReceiverPortAlreadyInUse(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:")
	require.NoError(t, err, "failed to open a port: %v", err)
	defer l.Close()
	_, portStr, err := net.SplitHostPort(l.Addr().String())
	require.NoError(t, err, "failed to split listener address: %v", err)
	traceReceiver, err := New(zipkinReceiver, "localhost:"+portStr, exportertest.NewNopTraceExporterOld())
	require.NoError(t, err, "Failed to create receiver: %v", err)
	err = traceReceiver.Start(context.Background(), componenttest.NewNopHost())
	if err == nil {
		traceReceiver.Shutdown(context.Background())
		t.Fatal("conflict on port was expected")
	}
}

func TestCustomHTTPServer(t *testing.T) {
	zr, err := New(zipkinReceiver, "localhost:9411", exportertest.NewNopTraceExporterOld())
	require.NoError(t, err, "Failed to create receiver: %v", err)

	server := &http.Server{}
	zr = zr.WithHTTPServer(server)
	assert.True(t, assert.ObjectsAreEqual(server, zr.server), "custom server passed to New was not used")
}

func TestConvertSpansToTraceSpans_json(t *testing.T) {
	// Using Adrian Cole's sample at https://gist.github.com/adriancole/e8823c19dfed64e2eb71
	blob, err := ioutil.ReadFile("./testdata/sample1.json")
	require.NoError(t, err, "Failed to read sample JSON file: %v", err)
	zi := new(ZipkinReceiver)
	reqs, err := zi.v2ToTraceSpans(blob, nil)
	require.NoError(t, err, "Failed to parse convert Zipkin spans in JSON to Trace spans: %v", err)

	require.Len(t, reqs, 1, "Expecting only one request since all spans share same node/localEndpoint: %v", len(reqs))

	req := reqs[0]
	wantNode := &commonpb.Node{
		ServiceInfo: &commonpb.ServiceInfo{
			Name: "frontend",
		},
	}
	assert.Equal(t, wantNode, req.Node)

	nonNilSpans := 0
	for _, span := range req.Spans {
		if span != nil {
			nonNilSpans++
		}
	}
	// Expecting 9 non-nil spans
	require.Equal(t, 9, nonNilSpans, "Incorrect non-nil spans count")
}

func TestConversionRoundtrip(t *testing.T) {
	// The goal is to convert from:
	// 1. Original Zipkin JSON as that's the format that Zipkin receivers will receive
	// 2. Into TraceProtoSpans
	// 3. Into SpanData
	// 4. Back into Zipkin JSON (in this case the Zipkin exporter has been configured)
	receiverInputJSON := []byte(`
[{
  "traceId": "4d1e00c0db9010db86154a4ba6e91385",
  "parentId": "86154a4ba6e91385",
  "id": "4d1e00c0db9010db",
  "kind": "CLIENT",
  "name": "get",
  "timestamp": 1472470996199000,
  "duration": 207000,
  "localEndpoint": {
    "serviceName": "frontend",
    "ipv6": "7::80:807f"
  },
  "remoteEndpoint": {
    "serviceName": "backend",
    "ipv4": "192.168.99.101",
    "port": 9000
  },
  "annotations": [
    {
      "timestamp": 1472470996238000,
      "value": "foo"
    },
    {
      "timestamp": 1472470996403000,
      "value": "bar"
    }
  ],
  "tags": {
    "http.path": "/api",
    "clnt/finagle.version": "6.45.0"
  }
},
{
  "traceId": "4d1e00c0db9010db86154a4ba6e91385",
  "parentId": "86154a4ba6e91386",
  "id": "4d1e00c0db9010db",
  "kind": "SERVER",
  "name": "put",
  "timestamp": 1472470996199000,
  "duration": 207000,
  "localEndpoint": {
    "serviceName": "frontend",
    "ipv6": "7::80:807f"
  },
  "remoteEndpoint": {
    "serviceName": "frontend",
    "ipv4": "192.168.99.101",
    "port": 9000
  },
  "annotations": [
    {
      "timestamp": 1472470996238000,
      "value": "foo"
    },
    {
      "timestamp": 1472470996403000,
      "value": "bar"
    }
  ],
  "tags": {
    "http.path": "/api",
    "clnt/finagle.version": "6.45.0"
  }
}]`)

	zi := &ZipkinReceiver{nextConsumer: exportertest.NewNopTraceExporterOld()}
	ereqs, err := zi.v2ToTraceSpans(receiverInputJSON, nil)
	require.NoError(t, err)

	wantProtoRequests := []consumerdata.TraceData{
		{
			Node: &commonpb.Node{
				ServiceInfo: &commonpb.ServiceInfo{Name: "frontend"},
			},

			Spans: []*tracepb.Span{
				{
					TraceId:      []byte{0x4d, 0x1e, 0x00, 0xc0, 0xdb, 0x90, 0x10, 0xdb, 0x86, 0x15, 0x4a, 0x4b, 0xa6, 0xe9, 0x13, 0x85},
					ParentSpanId: []byte{0x86, 0x15, 0x4a, 0x4b, 0xa6, 0xe9, 0x13, 0x85},
					SpanId:       []byte{0x4d, 0x1e, 0x00, 0xc0, 0xdb, 0x90, 0x10, 0xdb},
					Kind:         tracepb.Span_CLIENT,
					Name:         &tracepb.TruncatableString{Value: "get"},
					StartTime:    internal.TimeToTimestamp(time.Unix(int64(1472470996199000)/1e6, 1e3*(int64(1472470996199000)%1e6))),
					EndTime:      internal.TimeToTimestamp(time.Unix(int64(1472470996199000+207000)/1e6, 1e3*(int64(1472470996199000+207000)%1e6))),
					TimeEvents: &tracepb.Span_TimeEvents{
						TimeEvent: []*tracepb.Span_TimeEvent{
							{
								Time: internal.TimeToTimestamp(time.Unix(int64(1472470996238000)/1e6, 1e3*(int64(1472470996238000)%1e6))),
								Value: &tracepb.Span_TimeEvent_Annotation_{
									Annotation: &tracepb.Span_TimeEvent_Annotation{
										Description: &tracepb.TruncatableString{Value: "foo"},
									},
								},
							},
							{
								Time: internal.TimeToTimestamp(time.Unix(int64(1472470996403000)/1e6, 1e3*(int64(1472470996403000)%1e6))),
								Value: &tracepb.Span_TimeEvent_Annotation_{
									Annotation: &tracepb.Span_TimeEvent_Annotation{
										Description: &tracepb.TruncatableString{Value: "bar"},
									},
								},
							},
						},
					},
					Attributes: &tracepb.Span_Attributes{
						AttributeMap: map[string]*tracepb.AttributeValue{
							zipkin.LocalEndpointIPv6:         {Value: &tracepb.AttributeValue_StringValue{StringValue: &tracepb.TruncatableString{Value: "7::80:807f"}}},
							zipkin.RemoteEndpointServiceName: {Value: &tracepb.AttributeValue_StringValue{StringValue: &tracepb.TruncatableString{Value: "backend"}}},
							zipkin.RemoteEndpointIPv4:        {Value: &tracepb.AttributeValue_StringValue{StringValue: &tracepb.TruncatableString{Value: "192.168.99.101"}}},
							zipkin.RemoteEndpointPort:        {Value: &tracepb.AttributeValue_StringValue{StringValue: &tracepb.TruncatableString{Value: "9000"}}},
							"http.path":                      {Value: &tracepb.AttributeValue_StringValue{StringValue: &tracepb.TruncatableString{Value: "/api"}}},
							"clnt/finagle.version":           {Value: &tracepb.AttributeValue_StringValue{StringValue: &tracepb.TruncatableString{Value: "6.45.0"}}},
						},
					},
				},
				{
					TraceId:      []byte{0x4d, 0x1e, 0x00, 0xc0, 0xdb, 0x90, 0x10, 0xdb, 0x86, 0x15, 0x4a, 0x4b, 0xa6, 0xe9, 0x13, 0x85},
					SpanId:       []byte{0x4d, 0x1e, 0x00, 0xc0, 0xdb, 0x90, 0x10, 0xdb},
					ParentSpanId: []byte{0x86, 0x15, 0x4a, 0x4b, 0xa6, 0xe9, 0x13, 0x86},
					Kind:         tracepb.Span_SERVER,
					Name:         &tracepb.TruncatableString{Value: "put"},
					StartTime:    internal.TimeToTimestamp(time.Unix(int64(1472470996199000)/1e6, 1e3*(int64(1472470996199000)%1e6))),
					EndTime:      internal.TimeToTimestamp(time.Unix(int64(1472470996199000+207000)/1e6, 1e3*(int64(1472470996199000+207000)%1e6))),
					TimeEvents: &tracepb.Span_TimeEvents{
						TimeEvent: []*tracepb.Span_TimeEvent{
							{
								Time: internal.TimeToTimestamp(time.Unix(int64(1472470996238000)/1e6, 1e3*(int64(1472470996238000)%1e6))),
								Value: &tracepb.Span_TimeEvent_Annotation_{
									Annotation: &tracepb.Span_TimeEvent_Annotation{
										Description: &tracepb.TruncatableString{Value: "foo"},
									},
								},
							},
							{
								Time: internal.TimeToTimestamp(time.Unix(int64(1472470996403000)/1e6, 1e3*(int64(1472470996403000)%1e6))),
								Value: &tracepb.Span_TimeEvent_Annotation_{
									Annotation: &tracepb.Span_TimeEvent_Annotation{
										Description: &tracepb.TruncatableString{Value: "bar"},
									},
								},
							},
						},
					},
					Attributes: &tracepb.Span_Attributes{
						AttributeMap: map[string]*tracepb.AttributeValue{
							zipkin.LocalEndpointIPv6:         {Value: &tracepb.AttributeValue_StringValue{StringValue: &tracepb.TruncatableString{Value: "7::80:807f"}}},
							zipkin.RemoteEndpointServiceName: {Value: &tracepb.AttributeValue_StringValue{StringValue: &tracepb.TruncatableString{Value: "frontend"}}},
							zipkin.RemoteEndpointIPv4:        {Value: &tracepb.AttributeValue_StringValue{StringValue: &tracepb.TruncatableString{Value: "192.168.99.101"}}},
							zipkin.RemoteEndpointPort:        {Value: &tracepb.AttributeValue_StringValue{StringValue: &tracepb.TruncatableString{Value: "9000"}}},
							"http.path":                      {Value: &tracepb.AttributeValue_StringValue{StringValue: &tracepb.TruncatableString{Value: "/api"}}},
							"clnt/finagle.version":           {Value: &tracepb.AttributeValue_StringValue{StringValue: &tracepb.TruncatableString{Value: "6.45.0"}}},
						},
					},
				},
			},
		},
	}

	require.Equal(t, wantProtoRequests, ereqs)

	// Now the last phase is to transmit them over the wire and then compare the JSONs

	buf := new(bytes.Buffer)
	// This will act as the final Zipkin server.
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(buf, r.Body)
		_ = r.Body.Close()
	}))
	defer backend.Close()

	factory := &zipkinexporter.Factory{}
	config := &zipkinexporter.Config{
		ExporterSettings: configmodels.ExporterSettings{},
		URL:              backend.URL,
		Format:           "json",
	}
	ze, err := factory.CreateTraceExporter(zap.NewNop(), config)
	require.NoError(t, err)
	require.NotNil(t, ze)
	require.NoError(t, ze.Start(context.Background(), componenttest.NewNopHost()))

	for _, treq := range ereqs {
		require.NoError(t, ze.ConsumeTraceData(context.Background(), treq))
	}

	// Shutdown the exporter so it can flush any remaining data.
	assert.NoError(t, ze.Shutdown(context.Background()))
	backend.Close()

	// The received JSON messages are inside arrays, so reading then directly will
	// fail with error. Use a small hack to transform the multiple arrays into a
	// single one.
	accumulatedJSONMsgs := strings.Replace(buf.String(), "][", ",", -1)
	gj := testutils.GenerateNormalizedJSON(t, accumulatedJSONMsgs)
	wj := testutils.GenerateNormalizedJSON(t, string(receiverInputJSON))
	assert.Equal(t, wj, gj)
}

func TestStartTraceReception(t *testing.T) {
	tests := []struct {
		name    string
		host    component.Host
		wantErr bool
	}{
		{
			name:    "nil_host",
			wantErr: true,
		},
		{
			name: "valid_host",
			host: componenttest.NewNopHost(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sink := new(exportertest.SinkTraceExporterOld)
			zr, err := New(zipkinReceiver, "localhost:0", sink)
			require.Nil(t, err)
			require.NotNil(t, zr)

			err = zr.Start(context.Background(), tt.host)
			assert.Equal(t, tt.wantErr, err != nil)
			if !tt.wantErr {
				require.Nil(t, zr.Shutdown(context.Background()))
			}
		})
	}
}

func TestSpanKindTranslation(t *testing.T) {
	tests := []struct {
		zipkinKind zipkinmodel.Kind
		ocKind     tracepb.Span_SpanKind
		otKind     tracetranslator.OpenTracingSpanKind
	}{
		{
			zipkinKind: zipkinmodel.Client,
			ocKind:     tracepb.Span_CLIENT,
			otKind:     "",
		},
		{
			zipkinKind: zipkinmodel.Server,
			ocKind:     tracepb.Span_SERVER,
			otKind:     "",
		},
		{
			zipkinKind: zipkinmodel.Producer,
			ocKind:     tracepb.Span_SPAN_KIND_UNSPECIFIED,
			otKind:     tracetranslator.OpenTracingSpanKindProducer,
		},
		{
			zipkinKind: zipkinmodel.Consumer,
			ocKind:     tracepb.Span_SPAN_KIND_UNSPECIFIED,
			otKind:     tracetranslator.OpenTracingSpanKindConsumer,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.zipkinKind), func(t *testing.T) {
			zs := &zipkinmodel.SpanModel{
				SpanContext: zipkinmodel.SpanContext{
					TraceID: zipkinmodel.TraceID{Low: 123},
					ID:      456,
				},
				Kind:      tt.zipkinKind,
				Timestamp: time.Now(),
			}
			ocSpan, _ := zipkinSpanToTraceSpan(zs)
			assert.EqualValues(t, tt.ocKind, ocSpan.Kind)
			if tt.otKind != "" {
				otSpanKind := ocSpan.Attributes.AttributeMap[tracetranslator.TagSpanKind]
				assert.EqualValues(t, tt.otKind, otSpanKind.GetStringValue().Value)
			} else {
				assert.True(t, ocSpan.Attributes == nil)
			}
		})
	}
}

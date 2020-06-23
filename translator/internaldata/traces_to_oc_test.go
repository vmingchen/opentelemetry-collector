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

	occommon "github.com/census-instrumentation/opencensus-proto/gen-go/agent/common/v1"
	ocresource "github.com/census-instrumentation/opencensus-proto/gen-go/resource/v1"
	octrace "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/collector/consumer/consumerdata"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/internal/data/testdata"
	tracetranslator "go.opentelemetry.io/collector/translator/trace"
)

func TestInternalTraceStateToOC(t *testing.T) {
	assert.Equal(t, (*octrace.Span_Tracestate)(nil), traceStateToOC(pdata.TraceState("")))

	ocTracestate := &octrace.Span_Tracestate{
		Entries: []*octrace.Span_Tracestate_Entry{
			{
				Key:   "abc",
				Value: "def",
			},
		},
	}
	assert.EqualValues(t, ocTracestate, traceStateToOC(pdata.TraceState("abc=def")))

	ocTracestate.Entries = append(ocTracestate.Entries,
		&octrace.Span_Tracestate_Entry{
			Key:   "123",
			Value: "4567",
		})
	assert.EqualValues(t, ocTracestate, traceStateToOC(pdata.TraceState("abc=def,123=4567")))
}

func TestAttributesMapToOC(t *testing.T) {
	assert.EqualValues(t, (*octrace.Span_Attributes)(nil), attributesMapToOCSpanAttributes(pdata.NewAttributeMap(), 0))

	ocAttrs := &octrace.Span_Attributes{
		DroppedAttributesCount: 123,
	}
	assert.EqualValues(t, ocAttrs, attributesMapToOCSpanAttributes(pdata.NewAttributeMap(), 123))

	ocAttrs = &octrace.Span_Attributes{
		AttributeMap: map[string]*octrace.AttributeValue{
			"abc": {
				Value: &octrace.AttributeValue_StringValue{StringValue: &octrace.TruncatableString{Value: "def"}},
			},
		},
		DroppedAttributesCount: 234,
	}
	assert.EqualValues(t, ocAttrs,
		attributesMapToOCSpanAttributes(
			pdata.NewAttributeMap().InitFromMap(map[string]pdata.AttributeValue{
				"abc": pdata.NewAttributeValueString("def"),
			}),
			234))

	ocAttrs.AttributeMap["intval"] = &octrace.AttributeValue{
		Value: &octrace.AttributeValue_IntValue{IntValue: 345},
	}
	ocAttrs.AttributeMap["boolval"] = &octrace.AttributeValue{
		Value: &octrace.AttributeValue_BoolValue{BoolValue: true},
	}
	ocAttrs.AttributeMap["doubleval"] = &octrace.AttributeValue{
		Value: &octrace.AttributeValue_DoubleValue{DoubleValue: 4.5},
	}
	assert.EqualValues(t, ocAttrs,
		attributesMapToOCSpanAttributes(pdata.NewAttributeMap().InitFromMap(
			map[string]pdata.AttributeValue{
				"abc":       pdata.NewAttributeValueString("def"),
				"intval":    pdata.NewAttributeValueInt(345),
				"boolval":   pdata.NewAttributeValueBool(true),
				"doubleval": pdata.NewAttributeValueDouble(4.5),
			}),
			234))
}

func TestSpanKindToOC(t *testing.T) {
	tests := []struct {
		kind   pdata.SpanKind
		ocKind octrace.Span_SpanKind
	}{
		{
			kind:   pdata.SpanKindCLIENT,
			ocKind: octrace.Span_CLIENT,
		},
		{
			kind:   pdata.SpanKindSERVER,
			ocKind: octrace.Span_SERVER,
		},
		{
			kind:   pdata.SpanKindCONSUMER,
			ocKind: octrace.Span_SPAN_KIND_UNSPECIFIED,
		},
		{
			kind:   pdata.SpanKindPRODUCER,
			ocKind: octrace.Span_SPAN_KIND_UNSPECIFIED,
		},
		{
			kind:   pdata.SpanKindUNSPECIFIED,
			ocKind: octrace.Span_SPAN_KIND_UNSPECIFIED,
		},
		{
			kind:   pdata.SpanKindINTERNAL,
			ocKind: octrace.Span_SPAN_KIND_UNSPECIFIED,
		},
	}

	for _, test := range tests {
		t.Run(test.kind.String(), func(t *testing.T) {
			got := spanKindToOC(test.kind)
			assert.EqualValues(t, test.ocKind, got, "Expected "+test.ocKind.String()+", got "+got.String())
		})
	}
}

func TestSpanKindToOCAttribute(t *testing.T) {
	tests := []struct {
		kind        pdata.SpanKind
		ocAttribute *octrace.AttributeValue
	}{
		{
			kind: pdata.SpanKindCONSUMER,
			ocAttribute: &octrace.AttributeValue{
				Value: &octrace.AttributeValue_StringValue{
					StringValue: &octrace.TruncatableString{
						Value: string(tracetranslator.OpenTracingSpanKindConsumer),
					},
				},
			},
		},
		{
			kind: pdata.SpanKindPRODUCER,
			ocAttribute: &octrace.AttributeValue{
				Value: &octrace.AttributeValue_StringValue{
					StringValue: &octrace.TruncatableString{
						Value: string(tracetranslator.OpenTracingSpanKindProducer),
					},
				},
			},
		},
		{
			kind: pdata.SpanKindINTERNAL,
			ocAttribute: &octrace.AttributeValue{
				Value: &octrace.AttributeValue_StringValue{
					StringValue: &octrace.TruncatableString{
						Value: string(tracetranslator.OpenTracingSpanKindInternal),
					},
				},
			},
		},
		{
			kind:        pdata.SpanKindUNSPECIFIED,
			ocAttribute: nil,
		},
		{
			kind:        pdata.SpanKindSERVER,
			ocAttribute: nil,
		},
		{
			kind:        pdata.SpanKindCLIENT,
			ocAttribute: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.kind.String(), func(t *testing.T) {
			got := spanKindToOCAttribute(test.kind)
			assert.EqualValues(t, test.ocAttribute, got, "Expected "+test.ocAttribute.String()+", got "+got.String())
		})
	}
}

func TestInternalToOC(t *testing.T) {
	ocNode := &occommon.Node{}
	ocResource1 := &ocresource.Resource{Labels: map[string]string{"resource-attr": "resource-attr-val-1"}}
	ocResource2 := &ocresource.Resource{Labels: map[string]string{"resource-attr": "resource-attr-val-2"}}

	startTime, err := ptypes.TimestampProto(testdata.TestSpanStartTime)
	assert.NoError(t, err)
	eventTime, err := ptypes.TimestampProto(testdata.TestSpanEventTime)
	assert.NoError(t, err)
	endTime, err := ptypes.TimestampProto(testdata.TestSpanEndTime)
	assert.NoError(t, err)

	ocSpan1 := &octrace.Span{
		Name:      &octrace.TruncatableString{Value: "operationA"},
		StartTime: startTime,
		EndTime:   endTime,
		TimeEvents: &octrace.Span_TimeEvents{
			TimeEvent: []*octrace.Span_TimeEvent{
				{
					Time: eventTime,
					Value: &octrace.Span_TimeEvent_Annotation_{
						Annotation: &octrace.Span_TimeEvent_Annotation{
							Description: &octrace.TruncatableString{Value: "event-with-attr"},
							Attributes: &octrace.Span_Attributes{
								AttributeMap: map[string]*octrace.AttributeValue{
									"span-event-attr": {
										Value: &octrace.AttributeValue_StringValue{
											StringValue: &octrace.TruncatableString{Value: "span-event-attr-val"},
										},
									},
								},
								DroppedAttributesCount: 2,
							},
						},
					},
				},
				{
					Time: eventTime,
					Value: &octrace.Span_TimeEvent_Annotation_{
						Annotation: &octrace.Span_TimeEvent_Annotation{
							Description: &octrace.TruncatableString{Value: "event"},
							Attributes: &octrace.Span_Attributes{
								DroppedAttributesCount: 2,
							},
						},
					},
				},
			},
			DroppedAnnotationsCount: 1,
		},
		Attributes: &octrace.Span_Attributes{
			DroppedAttributesCount: 1,
		},
		Status: &octrace.Status{Message: "status-cancelled", Code: 1},
	}

	ocSpan2 := &octrace.Span{
		Name:      &octrace.TruncatableString{Value: "operationB"},
		StartTime: startTime,
		EndTime:   endTime,
		Links: &octrace.Span_Links{
			Link: []*octrace.Span_Link{
				{
					Attributes: &octrace.Span_Attributes{
						AttributeMap: map[string]*octrace.AttributeValue{
							"span-link-attr": {
								Value: &octrace.AttributeValue_StringValue{
									StringValue: &octrace.TruncatableString{Value: "span-link-attr-val"},
								},
							},
						},
						DroppedAttributesCount: 4,
					},
				},
				{
					Attributes: &octrace.Span_Attributes{
						DroppedAttributesCount: 4,
					},
				},
			},
			DroppedLinksCount: 3,
		},
	}

	ocSpan3 := &octrace.Span{
		Name:      &octrace.TruncatableString{Value: "operationC"},
		StartTime: startTime,
		EndTime:   endTime,
		// TODO: Set resource here and put it in the same TraceDataOld.
		Attributes: &octrace.Span_Attributes{
			AttributeMap: map[string]*octrace.AttributeValue{
				"span-attr": {
					Value: &octrace.AttributeValue_StringValue{
						StringValue: &octrace.TruncatableString{Value: "span-attr-val"},
					},
				},
			},
			DroppedAttributesCount: 5,
		},
	}

	tests := []struct {
		name string
		td   pdata.Traces
		oc   []consumerdata.TraceData
	}{
		{
			name: "empty",
			td:   testdata.GenerateTraceDataEmpty(),
			oc:   []consumerdata.TraceData(nil),
		},

		{
			name: "one-empty-resource-spans",
			td:   testdata.GenerateTraceDataOneEmptyResourceSpans(),
			oc: []consumerdata.TraceData{
				{
					Node:         nil,
					Resource:     nil,
					Spans:        []*octrace.Span(nil),
					SourceFormat: sourceFormat,
				},
			},
		},

		{
			name: "one-empty-one-nil-resource-spans",
			td:   testdata.GenerateTraceDataOneEmptyOneNilResourceSpans(),
			oc: []consumerdata.TraceData{
				{
					Node:         nil,
					Resource:     nil,
					Spans:        []*octrace.Span(nil),
					SourceFormat: sourceFormat,
				},
			},
		},

		{
			name: "no-libraries",
			td:   testdata.GenerateTraceDataNoLibraries(),
			oc: []consumerdata.TraceData{
				{
					Node:         ocNode,
					Resource:     ocResource1,
					Spans:        []*octrace.Span(nil),
					SourceFormat: sourceFormat,
				},
			},
		},

		{
			name: "one-empty-instrumentation-library",
			td:   testdata.GenerateTraceDataOneEmptyInstrumentationLibrary(),
			oc: []consumerdata.TraceData{
				{
					Node:         ocNode,
					Resource:     ocResource1,
					Spans:        []*octrace.Span{},
					SourceFormat: sourceFormat,
				},
			},
		},

		{
			name: "one-empty-one-nil-instrumentation-library",
			td:   testdata.GenerateTraceDataOneEmptyOneNilInstrumentationLibrary(),
			oc: []consumerdata.TraceData{
				{
					Node:         ocNode,
					Resource:     ocResource1,
					Spans:        []*octrace.Span{},
					SourceFormat: sourceFormat,
				},
			},
		},

		{
			name: "one-span-no-resource",
			td:   testdata.GenerateTraceDataOneSpanNoResource(),
			oc: []consumerdata.TraceData{
				{
					Node:         nil,
					Resource:     nil,
					Spans:        []*octrace.Span{ocSpan1},
					SourceFormat: sourceFormat,
				},
			},
		},

		{
			name: "one-span",
			td:   testdata.GenerateTraceDataOneSpan(),
			oc: []consumerdata.TraceData{
				{
					Node:         ocNode,
					Resource:     ocResource1,
					Spans:        []*octrace.Span{ocSpan1},
					SourceFormat: sourceFormat,
				},
			},
		},

		{
			name: "one-span-one-nil",
			td:   testdata.GenerateTraceDataOneSpanOneNil(),
			oc: []consumerdata.TraceData{
				{
					Node:         ocNode,
					Resource:     ocResource1,
					Spans:        []*octrace.Span{ocSpan1},
					SourceFormat: sourceFormat,
				},
			},
		},

		{
			name: "two-spans-same-resource",
			td:   testdata.GenerateTraceDataTwoSpansSameResource(),
			oc: []consumerdata.TraceData{
				{
					Node:         ocNode,
					Resource:     ocResource1,
					Spans:        []*octrace.Span{ocSpan1, ocSpan2},
					SourceFormat: sourceFormat,
				},
			},
		},

		{
			name: "two-spans-same-resource-one-different",
			td:   testdata.GenerateTraceDataTwoSpansSameResourceOneDifferent(),
			oc: []consumerdata.TraceData{
				{
					Node:         ocNode,
					Resource:     ocResource1,
					Spans:        []*octrace.Span{ocSpan1, ocSpan2},
					SourceFormat: sourceFormat,
				},
				{
					Node:         ocNode,
					Resource:     ocResource2,
					Spans:        []*octrace.Span{ocSpan3},
					SourceFormat: sourceFormat,
				},
			},
		},
	}

	assert.EqualValues(t, testdata.NumTraceTests, len(tests))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.EqualValues(t, test.oc, TraceDataToOC(test.td))
		})
	}
}

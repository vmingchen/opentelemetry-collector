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

package testdata

import (
	otlpcommon "go.opentelemetry.io/collector/internal/data/opentelemetry-proto-gen/common/v1"

	"go.opentelemetry.io/collector/consumer/pdata"
)

var (
	resourceAttributes1 = map[string]pdata.AttributeValue{"resource-attr": pdata.NewAttributeValueString("resource-attr-val-1")}
	resourceAttributes2 = map[string]pdata.AttributeValue{"resource-attr": pdata.NewAttributeValueString("resource-attr-val-2")}
	spanEventAttributes = map[string]pdata.AttributeValue{"span-event-attr": pdata.NewAttributeValueString("span-event-attr-val")}
	spanLinkAttributes  = map[string]pdata.AttributeValue{"span-link-attr": pdata.NewAttributeValueString("span-link-attr-val")}
	spanAttributes      = map[string]pdata.AttributeValue{"span-attr": pdata.NewAttributeValueString("span-attr-val")}
)

const (
	TestLabelKey        = "label"
	TestLabelKey1       = "label-1"
	TestLabelValue1     = "label-value-1"
	TestLabelKey2       = "label-2"
	TestLabelValue2     = "label-value-2"
	TestLabelKey3       = "label-3"
	TestLabelValue3     = "label-value-3"
	TestAttachmentKey   = "exemplar-attachment"
	TestAttachmentValue = "exemplar-attachment-value"
)

func initResourceAttributes1(dest pdata.AttributeMap) {
	dest.InitFromMap(resourceAttributes1)
}

func generateOtlpResourceAttributes1() []*otlpcommon.AttributeKeyValue {
	return []*otlpcommon.AttributeKeyValue{
		{
			Key:         "resource-attr",
			StringValue: "resource-attr-val-1",
		},
	}
}

func initResourceAttributes2(dest pdata.AttributeMap) {
	dest.InitFromMap(resourceAttributes2)
}

func generateOtlpResourceAttributes2() []*otlpcommon.AttributeKeyValue {
	return []*otlpcommon.AttributeKeyValue{
		{
			Key:         "resource-attr",
			StringValue: "resource-attr-val-2",
		},
	}
}

func initSpanAttributes(dest pdata.AttributeMap) {
	dest.InitFromMap(spanAttributes)
}

func generateOtlpSpanAttributes() []*otlpcommon.AttributeKeyValue {
	return []*otlpcommon.AttributeKeyValue{
		{
			Key:         "span-attr",
			StringValue: "span-attr-val",
		},
	}
}

func initSpanEventAttributes(dest pdata.AttributeMap) {
	dest.InitFromMap(spanEventAttributes)
}

func generateOtlpSpanEventAttributes() []*otlpcommon.AttributeKeyValue {
	return []*otlpcommon.AttributeKeyValue{
		{
			Key:         "span-event-attr",
			StringValue: "span-event-attr-val",
		},
	}
}

func initSpanLinkAttributes(dest pdata.AttributeMap) {
	dest.InitFromMap(spanLinkAttributes)
}

func generateOtlpSpanLinkAttributes() []*otlpcommon.AttributeKeyValue {
	return []*otlpcommon.AttributeKeyValue{
		{
			Key:         "span-link-attr",
			StringValue: "span-link-attr-val",
		},
	}
}

func initMetricLabels1(dest pdata.StringMap) {
	dest.InitFromMap(map[string]string{TestLabelKey1: TestLabelValue1})
}

func generateOtlpMetricLabels1() []*otlpcommon.StringKeyValue {
	return []*otlpcommon.StringKeyValue{
		{
			Key:   TestLabelKey1,
			Value: TestLabelValue1,
		},
	}
}

func initMetricLabelValue1(dest pdata.StringMap) {
	dest.InitFromMap(map[string]string{TestLabelKey: TestLabelValue1})
}

func generateOtlpMetricLabelValue1() []*otlpcommon.StringKeyValue {
	return []*otlpcommon.StringKeyValue{
		{
			Key:   TestLabelKey,
			Value: TestLabelValue1,
		},
	}
}

func initMetricLabels12(dest pdata.StringMap) {
	dest.InitFromMap(map[string]string{TestLabelKey1: TestLabelValue1, TestLabelKey2: TestLabelValue2}).Sort()
}

func generateOtlpMetricLabels12() []*otlpcommon.StringKeyValue {
	return []*otlpcommon.StringKeyValue{
		{
			Key:   TestLabelKey1,
			Value: TestLabelValue1,
		},
		{
			Key:   TestLabelKey2,
			Value: TestLabelValue2,
		},
	}
}

func initMetricLabels13(dest pdata.StringMap) {
	dest.InitFromMap(map[string]string{TestLabelKey1: TestLabelValue1, TestLabelKey3: TestLabelValue3}).Sort()
}

func generateOtlpMetricLabels13() []*otlpcommon.StringKeyValue {
	return []*otlpcommon.StringKeyValue{
		{
			Key:   TestLabelKey1,
			Value: TestLabelValue1,
		},
		{
			Key:   TestLabelKey3,
			Value: TestLabelValue3,
		},
	}
}

func initMetricLabels2(dest pdata.StringMap) {
	dest.InitFromMap(map[string]string{TestLabelKey2: TestLabelValue2})
}

func generateOtlpMetricLabels2() []*otlpcommon.StringKeyValue {
	return []*otlpcommon.StringKeyValue{
		{
			Key:   TestLabelKey2,
			Value: TestLabelValue2,
		},
	}
}

func initMetricLabelValue2(dest pdata.StringMap) {
	dest.InitFromMap(map[string]string{TestLabelKey: TestLabelValue2})
}

func generateOtlpMetricLabelValue2() []*otlpcommon.StringKeyValue {
	return []*otlpcommon.StringKeyValue{
		{
			Key:   TestLabelKey,
			Value: TestLabelValue2,
		},
	}
}

func initMetricAttachment(dest pdata.StringMap) {
	dest.InitFromMap(map[string]string{TestAttachmentKey: TestAttachmentValue})
}

func generateOtlpMetricAttachment() []*otlpcommon.StringKeyValue {
	return []*otlpcommon.StringKeyValue{
		{
			Key:   TestAttachmentKey,
			Value: TestAttachmentValue,
		},
	}
}

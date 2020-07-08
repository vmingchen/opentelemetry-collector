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
	"go.opentelemetry.io/collector/consumer/pdata"
)

// labels

const (
	deviceLabelName    = "device"
	directionLabelName = "direction"
)

// direction label values

const (
	readDirectionLabelValue  = "read"
	writeDirectionLabelValue = "write"
)

// descriptors

var diskIODescriptor = func() pdata.MetricDescriptor {
	descriptor := pdata.NewMetricDescriptor()
	descriptor.InitEmpty()
	descriptor.SetName("system.disk.io")
	descriptor.SetDescription("Disk bytes transferred.")
	descriptor.SetUnit("bytes")
	descriptor.SetType(pdata.MetricTypeMonotonicInt64)
	return descriptor
}()

var diskOpsDescriptor = func() pdata.MetricDescriptor {
	descriptor := pdata.NewMetricDescriptor()
	descriptor.InitEmpty()
	descriptor.SetName("system.disk.ops")
	descriptor.SetDescription("Disk operations count.")
	descriptor.SetUnit("1")
	descriptor.SetType(pdata.MetricTypeMonotonicInt64)
	return descriptor
}()

var diskTimeDescriptor = func() pdata.MetricDescriptor {
	descriptor := pdata.NewMetricDescriptor()
	descriptor.InitEmpty()
	descriptor.SetName("system.disk.time")
	descriptor.SetDescription("Time spent in disk operations.")
	descriptor.SetUnit("ms")
	descriptor.SetType(pdata.MetricTypeMonotonicInt64)
	return descriptor
}()

var diskMergedDescriptor = func() pdata.MetricDescriptor {
	descriptor := pdata.NewMetricDescriptor()
	descriptor.InitEmpty()
	descriptor.SetName("system.disk.merged")
	descriptor.SetDescription("The number of disk reads merged into single physical disk access operations.")
	descriptor.SetUnit("1")
	descriptor.SetType(pdata.MetricTypeMonotonicInt64)
	return descriptor
}()

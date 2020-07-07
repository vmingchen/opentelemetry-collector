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

package processscraper

import (
	"go.opentelemetry.io/collector/consumer/pdata"
)

// labels

const (
	cmdlineLabelName   = "command_line"
	directionLabelName = "direction"
	pidLabelName       = "pid"
	processLabelName   = "process"
	stateLabelName     = "state"
	usernameLabelName  = "username"
)

// direction label values

const (
	readDirectionLabelValue  = "read"
	writeDirectionLabelValue = "write"
)

// state label values

const (
	userStateLabelValue   = "user"
	systemStateLabelValue = "system"
	waitStateLabelValue   = "wait"
)

// descriptors

var cpuTimeDescriptor = func() pdata.MetricDescriptor {
	descriptor := pdata.NewMetricDescriptor()
	descriptor.InitEmpty()
	descriptor.SetName("process.cpu.time")
	descriptor.SetDescription("Total CPU seconds broken down by different states.")
	descriptor.SetUnit("s")
	descriptor.SetType(pdata.MetricTypeMonotonicDouble)
	return descriptor
}()

var memoryUsageDescriptor = func() pdata.MetricDescriptor {
	descriptor := pdata.NewMetricDescriptor()
	descriptor.InitEmpty()
	descriptor.SetName("process.memory.usage")
	descriptor.SetDescription("Bytes of memory in use.")
	descriptor.SetUnit("bytes")
	descriptor.SetType(pdata.MetricTypeInt64)
	return descriptor
}()

var diskIODescriptor = func() pdata.MetricDescriptor {
	descriptor := pdata.NewMetricDescriptor()
	descriptor.InitEmpty()
	descriptor.SetName("process.disk.io")
	descriptor.SetDescription("Disk bytes transferred.")
	descriptor.SetUnit("bytes")
	descriptor.SetType(pdata.MetricTypeMonotonicInt64)
	return descriptor
}()

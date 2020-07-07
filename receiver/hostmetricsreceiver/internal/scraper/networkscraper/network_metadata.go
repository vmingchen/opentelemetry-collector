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

package networkscraper

import (
	"go.opentelemetry.io/collector/consumer/pdata"
)

// network metric constants

const (
	directionLabelName = "direction"
	stateLabelName     = "state"
)

// direction label values

const (
	receiveDirectionLabelValue  = "receive"
	transmitDirectionLabelValue = "transmit"
)

// descriptors

var networkPacketsDescriptor = func() pdata.MetricDescriptor {
	descriptor := pdata.NewMetricDescriptor()
	descriptor.InitEmpty()
	descriptor.SetName("system.network.packets")
	descriptor.SetDescription("The number of packets transferred.")
	descriptor.SetUnit("1")
	descriptor.SetType(pdata.MetricTypeMonotonicInt64)
	return descriptor
}()

var networkDroppedPacketsDescriptor = func() pdata.MetricDescriptor {
	descriptor := pdata.NewMetricDescriptor()
	descriptor.InitEmpty()
	descriptor.SetName("system.network.dropped_packets")
	descriptor.SetDescription("The number of packets dropped.")
	descriptor.SetUnit("1")
	descriptor.SetType(pdata.MetricTypeMonotonicInt64)
	return descriptor
}()

var networkErrorsDescriptor = func() pdata.MetricDescriptor {
	descriptor := pdata.NewMetricDescriptor()
	descriptor.InitEmpty()
	descriptor.SetName("system.network.errors")
	descriptor.SetDescription("The number of errors encountered")
	descriptor.SetUnit("1")
	descriptor.SetType(pdata.MetricTypeMonotonicInt64)
	return descriptor
}()

var networkIODescriptor = func() pdata.MetricDescriptor {
	descriptor := pdata.NewMetricDescriptor()
	descriptor.InitEmpty()
	descriptor.SetName("system.network.io")
	descriptor.SetDescription("The number of bytes transmitted and received")
	descriptor.SetUnit("bytes")
	descriptor.SetType(pdata.MetricTypeMonotonicInt64)
	return descriptor
}()

var networkTCPConnectionsDescriptor = func() pdata.MetricDescriptor {
	descriptor := pdata.NewMetricDescriptor()
	descriptor.InitEmpty()
	descriptor.SetName("system.network.tcp_connections")
	descriptor.SetDescription("The number of tcp connections")
	descriptor.SetUnit("bytes")
	descriptor.SetType(pdata.MetricTypeInt64)
	return descriptor
}()

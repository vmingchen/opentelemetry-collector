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

// +build !linux,!windows

package processscraper

import (
	"github.com/shirou/gopsutil/cpu"

	"go.opentelemetry.io/collector/consumer/pdata"
)

const cpuStatesLen = 2

func appendCPUTimeStateDataPoints(ddps pdata.DoubleDataPointSlice, startTime pdata.TimestampUnixNano, cpuTime *cpu.TimesStat) {
	initializeCPUTimeDataPoint(ddps.At(0), startTime, cpuTime.User, userStateLabelValue)
	initializeCPUTimeDataPoint(ddps.At(1), startTime, cpuTime.System, systemStateLabelValue)
}

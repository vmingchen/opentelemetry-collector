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

// +build !linux

package memoryscraper

import (
	"github.com/shirou/gopsutil/mem"

	"go.opentelemetry.io/collector/consumer/pdata"
)

const availableStateLabelValue = "available"

const memStatesLen = 2

func appendMemoryUsageStateDataPoints(idps pdata.Int64DataPointSlice, memInfo *mem.VirtualMemoryStat) {
	initializeMemoryUsageDataPoint(idps.At(0), usedStateLabelValue, int64(memInfo.Used))
	initializeMemoryUsageDataPoint(idps.At(1), availableStateLabelValue, int64(memInfo.Available))
}

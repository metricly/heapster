// Copyright 2018 Google Inc. All Rights Reserved.
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
package metricly

import (
	"github.com/golang/glog"
	"github.com/metricly/go-client/model/core"
)

//Calculate usage percent metrics for an element
func CreateComputedMetric(ts int64, e *core.Element) {
	createCpuUsageComputedMetric(ts, e)
	createMemUsageComputedMetric(ts, e)
}

func createCpuUsageComputedMetric(ts int64, e *core.Element) {
	limit := -1.0
	request := -1.0
	usageRate := -1.0
	usage := -1.0
	for _, s := range e.Samples() {
		switch s.MetricId() {
		case "cpu.request":
			request = s.Val()
		case "cpu.limit":
			limit = s.Val()
		case "cpu.usage":
			usage = s.Val()
		case "cpu.usage_rate":
			usageRate = s.Val()
		}
	}
	usagePercent := calculateCpuUsagePercent(limit, request, usageRate, usage)
	glog.V(4).Infof("calculated cpu usage percent: %f for element: %s using limit: %f, request: %f, usageRate: %f, usage: %f", usagePercent, e.Id, limit, request, usageRate, usage)
	if usagePercent != -1.0 {
		sample, _ := core.NewSample("cpu.usage.percent", ts, usagePercent)
		e.AddSample(sample)
	}
}

func createMemUsageComputedMetric(ts int64, e *core.Element) {
	limit := -1.0
	request := -1.0
	usage := -1.0
	for _, s := range e.Samples() {
		switch s.MetricId() {
		case "memory.request":
			request = s.Val()
		case "memory.limit":
			limit = s.Val()
		case "memory.usage":
			usage = s.Val()
		}
	}
	usagePercent := calculateMemUsagePercent(limit, request, usage)
	glog.V(4).Infof("calculated memory usage percent: %f for element: %s using limit: %f, request: %f, usage: %f", usagePercent, e.Id, limit, request, usage)
	if usagePercent != -1.0 {
		sample, _ := core.NewSample("memory.usage.percent", ts, usagePercent)
		e.AddSample(sample)
	}
}

func calculateCpuUsagePercent(limit, request, usageRate, usage float64) float64 {
	if limit != 0.0 && limit != -1.0 && usageRate != -1.0 {
		return 100.0 * (usageRate / limit)
	}
	if request != 0.0 && request != -1.0 && usageRate != -1.0 {
		return 100.0 * (usageRate / request)
	}
	if usage != -1.0 {
		return 100.0 * usage / (60 * 1000000000)
	}
	return -1.0
}

func calculateMemUsagePercent(limit, request, usage float64) float64 {
	if limit != 0.0 && limit != -1.0 && usage != -1.0 {
		return 100.0 * (usage / limit)
	}
	if request != 0.0 && request != -1.0 && usage != -1.0 {
		return 100.0 * (usage / request)
	}
	if usage == 0.0 {
		return 0.0
	}
	return -1.0
}

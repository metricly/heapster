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
	"testing"
	"time"

	core "github.com/metricly/go-client/model/core"
)

func TestPodComputedMetric(t *testing.T) {
	//given
	pod := core.NewElement("kube-system/heapster", "kube-system/heapster", "Kubernetes Pod", "")
	now := time.Now()
	cpuUsageRate, _ := core.NewSample("cpu.usage_rate", now, 500.0)
	cpuUsage, _ := core.NewSample("cpu.usage", now, 0.0)
	cpuRequest, _ := core.NewSample("cpu.request", now, 0.0)
	cpuLimit, _ := core.NewSample("cpu.limit", now, 1000.0)
	pod.AddSample(cpuUsageRate)
	pod.AddSample(cpuUsage)
	pod.AddSample(cpuRequest)
	pod.AddSample(cpuLimit)
	//when
	CreateComputedMetric(now.Unix()*1000, &pod)
	//then
	if len(pod.Samples()) != 5 {
		t.Errorf("there should be 5 samples including a computed 'cpu.usage.percent', but actually has %d", len(pod.Samples()))
	}
	for _, s := range pod.Samples() {
		if s.MetricId() == "cpu.usage.percent" {
			if s.Val() != 50.0 {
				t.Errorf("cpu usage percent should be 50.0 and the actual is %f", s.Val())
			}
		}
	}
}

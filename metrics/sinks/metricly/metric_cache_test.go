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

	"github.com/metricly/go-client/model/core"
)

func TestContainsMetric(t *testing.T) {
	//given
	mc := NewMetricCache(60)
	//when
	found := mc.ContainsMetric("cpu.usage")
	//then
	if !found {
		t.Errorf("'cpu.usage' should exist as a counter metric")
	}
}

func TestNotContainsMetric(t *testing.T) {
	//given
	mc := NewMetricCache(60)
	//when
	found := mc.ContainsMetric("cpu.request")
	//then
	if found {
		t.Errorf("'cpu.request' should not exist as a counter metric")
	}

}

func TestAddSample(t *testing.T) {
	//given
	mc := NewMetricCache(60)
	//when
	sample, _ := core.NewSample("cpu.usage", time.Now(), 0.0)
	mc.addSample("elementId", sample)
	//then
	result, ok := mc.getSample("elementId", "cpu.usage")

	if !ok || result.val != 0.0 {
		t.Errorf("'cpu.usage' sample should be added to metric cache")
	}
}

func TestAddNonCounterSample(t *testing.T) {
	//given
	mc := NewMetricCache(60)
	//when
	sample, _ := core.NewSample("cpu.request", time.Now(), 0.0)
	mc.addSample("elementId", sample)
	//then
	_, ok := mc.getSample("elementId", "cpu.request")

	if ok {
		t.Errorf("'cpu.request' sample should not be added to metric cache")
	}
	if mc.Size() != 0 {
		t.Errorf("metric cache should be empty")
	}
}

func TestAddSampleExpiration(t *testing.T) {
	//given
	mc := NewMetricCache(1)
	//when
	sample, _ := core.NewSample("cpu.usage", time.Now(), 0.0)
	mc.addSample("elementId", sample)
	duration := time.Duration(3) * time.Second
	time.Sleep(duration)
	//then
	_, ok := mc.getSample("elementId", "cpu.usage")
	if ok {
		t.Errorf("'cpu.usage' sample should no longer be in metric cache due to ttl")
	}
	if mc.Size() != 0 {
		t.Errorf("metric cache should be empty")
	}
}

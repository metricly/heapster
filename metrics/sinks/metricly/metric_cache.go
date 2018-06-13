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
	"sync"
	"time"

	"github.com/metricly/go-client/model/core"
)

type metricKey struct {
	elementId string
	metricId  string
}

type sample struct {
	timestamp int64
	val       float64
}

//MetricCache contains a collection of samples for COUNTER metrics
// counters contains predefined metrics that are allowed to be added to the cache
type MetricCache struct {
	counters map[string]struct{}
	cache    map[metricKey]sample
	lock     sync.Mutex
}

//NewMetricCache create a new MetricCache
func NewMetricCache(ttl int) (mc *MetricCache) {
	mc = &MetricCache{cache: make(map[metricKey]sample),
		counters: map[string]struct{}{
			"cpu.usage": struct{}{},
		}}
	go func() {
		for now := range time.Tick(time.Second) {
			mc.lock.Lock()
			for k, v := range mc.cache {
				if now.Unix()*1000-v.timestamp > int64(ttl*1000) {
					delete(mc.cache, k)
				}
			}
			mc.lock.Unlock()
		}
	}()
	return
}

// returns the size of the metric cache
func (mc *MetricCache) Size() int {
	mc.lock.Lock()
	size := len(mc.cache)
	mc.lock.Unlock()
	return size
}

func (mc *MetricCache) addSample(elementId string, s core.Sample) {
	if _, ok := mc.counters[s.MetricId()]; ok {
		mc.lock.Lock()
		mc.cache[metricKey{elementId, s.MetricId()}] = sample{s.Timestamp(), s.Val()}
		mc.lock.Unlock()
	}
}

func (mc *MetricCache) getSample(elementId, metricId string) (sample, bool) {
	mc.lock.Lock()
	s, ok := mc.cache[metricKey{elementId, metricId}]
	mc.lock.Unlock()
	return s, ok
}

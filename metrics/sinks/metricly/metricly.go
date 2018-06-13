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
	"net/url"
	"strings"

	"github.com/golang/glog"
	"github.com/metricly/go-client/api"
	metricly_core "github.com/metricly/go-client/model/core"
	"k8s.io/heapster/common/metricly"
	"k8s.io/heapster/metrics/core"
)

const (
	defaultElementsPayloadSize = 20
	elementTypePrefix          = "Kubernetes "
)

type MetriclyMetricsSink struct {
	client api.Client
	config metricly.MetriclyConfig
	cache  *MetricCache
}

func (sink *MetriclyMetricsSink) Name() string {
	return "Metricly Metrics Sink"
}

func (sink *MetriclyMetricsSink) Stop() {
}

type chunk struct {
	start, end int
}

func (sink *MetriclyMetricsSink) ExportData(batch *core.DataBatch) {
	glog.Info("Start exporting data batch to Metricly ...")
	elements := DataBatchToElements(sink.config, sink.cache, batch)
	elementsPayloadSize := defaultElementsPayloadSize
	if sink.config.ElementBatchSize > 0 {
		elementsPayloadSize = sink.config.ElementBatchSize
	}
	total := len(elements)
	chunks := partition(total, elementsPayloadSize)
	jobs := make(chan chunk, len(chunks))
	count := make(chan int)
	for _, c := range chunks {
		jobs <- c
	}
	close(jobs)

	for c := range jobs {
		go func(c chunk) {
			if err := sink.send(elements[c.start:c.end]); err == nil {
				count <- c.end - c.start
			} else {
				glog.Warningf("Error occurred during exporting %d elements with response:  %v", c.end-c.start, err)
				count <- 0
			}
		}(c)
	}

	var sent int
	for i := 0; i < len(chunks); i++ {
		sent += <-count
	}
	glog.Infof("Exported %d out of %d elements using %d workers", sent, len(elements), len(chunks))
}

func NewMetriclySink(uri *url.URL) (core.DataSink, error) {
	config, _ := metricly.Config(uri)
	glog.Info("Create Metricly sink using config: ", config)
	return &MetriclyMetricsSink{client: api.NewClient(config.ApiURL, config.ApiKey), config: config, cache: NewMetricCache(300)}, nil
}

func DataBatchToElements(config metricly.MetriclyConfig, cache *MetricCache, batch *core.DataBatch) []metricly_core.Element {
	ts := batch.Timestamp.Unix() * 1000
	var elements []metricly_core.Element
	for key, ms := range batch.MetricSets {
		if !filter(config.InclusionFilters, config.ExclusionFilters, ms) {
			glog.V(1).Info("metric set is dropped due to filtering, key: ", key)
			continue
		}
		etype := ms.Labels["type"]
		element := metricly_core.NewElement(key, shortenName(key), prettyElementType(etype), "")
		// metric set labels to element tags
		for lname, lvalue := range ms.Labels {
			if lname == "labels" {
				for _, l := range strings.Split(lvalue, ",") {
					kv := strings.SplitN(l, ":", 2)
					element.AddTag(kv[0], kv[1])
				}
			} else {
				element.AddTag(lname, lvalue)
			}
		}
		// metrics
		for mname, mvalue := range ms.MetricValues {
			if sample, err := metricly_core.NewSample(sanitizeMetricId(mname), ts, mvalue.GetValue()); err == nil {
				if cache.ContainsMetric(sample.MetricId()) {
					if lastSample, found := cache.getSample(element.Id, sample.MetricId()); found {
						if delta, err := metricly_core.NewSample(sample.MetricId(), ts, sample.Val()-lastSample.val); err == nil {
							element.AddSample(delta)
						}
					}
					cache.addSample(element.Id, sample)
				} else {
					element.AddSample(sample)
				}
			}
		}
		// labeled metrics
		for _, lmetric := range ms.LabeledMetrics {
			instanceMetricName := sanitizeMetricId(lmetric.Name) + ":" + lmetric.Labels["resource_id"]
			if sample, err := metricly_core.NewSample(instanceMetricName, ts, lmetric.GetValue()); err == nil {
				element.AddSample(sample)
			}
		}
		elements = append(elements, element)
	}
	LinkElements(elements)
	CreateComputedMetrics(ts, elements)
	return elements
}

func LinkElements(elements []metricly_core.Element) {
	var elementsById = make(map[string]*metricly_core.Element)
	for idx := range elements {
		switch e := elements[idx]; e.Type {
		case "Kubernetes Pod":
			if id, ok := elements[idx].Tag("pod_id"); ok {
				elementsById[id.Value] = &e
			}
		case "Kubernetes Node":
			if id, ok := elements[idx].Tag("host_id"); ok {
				elementsById[id.Value] = &e
			}
		case "Kubernetes Namespace":
			if id, ok := elements[idx].Tag("namespace_id"); ok {
				elementsById[id.Value] = &e
			}
		}
	}

	for idx := range elements {
		switch e := elements[idx]; e.Type {
		case "Kubernetes Pod Container":
			if podId, ok := e.Tag("pod_id"); ok {
				if pod, ok := elementsById[podId.Value]; ok {
					pod.AddRelation(e.Id)
				}
			}
			if hostId, ok := elements[idx].Tag("host_id"); ok {
				if host, ok := elementsById[hostId.Value]; ok {
					host.AddRelation(e.Id)
				}
			}
			if nsId, ok := elements[idx].Tag("namespace_id"); ok {
				if ns, ok := elementsById[nsId.Value]; ok {
					ns.AddRelation(e.Id)
				}
			}
		case "Kubernetes Pod":
			if hostId, ok := elements[idx].Tag("host_id"); ok {
				if host, ok := elementsById[hostId.Value]; ok {
					host.AddRelation(e.Id)
				}
			}
			if nsId, ok := elements[idx].Tag("namespace_id"); ok {
				if ns, ok := elementsById[nsId.Value]; ok {
					ns.AddRelation(e.Id)
				}
			}
		case "Kubernetes Sys Container":
			if hostId, ok := elements[idx].Tag("host_id"); ok {
				if host, ok := elementsById[hostId.Value]; ok {
					host.AddRelation(e.Id)
				}
			}
		}
	}
}

//CreateComputedMetrics creates new computed metrics/samples based on current collected metrics
func CreateComputedMetrics(timestamp int64, elements []metricly_core.Element) {
	for _, e := range elements {
		CreateComputedMetric(timestamp, &e)
	}
}

func shortenName(fqn string) string {
	var names []string
	for _, s := range strings.Split(fqn, "/") {
		kv := strings.SplitN(s, ":", 2)
		if len(kv) == 2 {
			names = append(names, kv[1])
		} else {
			names = append(names, kv[0])
		}
	}
	return strings.Join(names, "/")
}

func prettyElementType(etype string) string {
	switch etype {
	case "ns":
		return elementTypePrefix + "Namespace"
	}
	return elementTypePrefix + strings.Title(strings.Replace(etype, "_", " ", -1))
}

//filter MetricSet against inclusion/exclusion filters and return true if it passes
func filter(inf []metricly.Filter, exf []metricly.Filter, ms *core.MetricSet) bool {
	return include(inf, ms) && !exclude(exf, ms)
}

func exclude(filters []metricly.Filter, ms *core.MetricSet) bool {
	if len(filters) == 0 {
		return false
	}
	for k, v := range ms.Labels {
		for _, f := range filters {
			if f.Type == "label" && k == f.Name && f.Regex.MatchString(v) {
				return true
			}
		}
	}
	return false
}

func include(filters []metricly.Filter, ms *core.MetricSet) bool {
	if len(filters) == 0 {
		return true
	}
	for k, v := range ms.Labels {
		for _, f := range filters {
			if f.Type == "label" && k == f.Name && f.Regex.MatchString(v) {
				return true
			}
		}
	}
	return false
}

func sanitizeMetricId(metricId string) string {
	return strings.Replace(metricId, "/", ".", -1)
}

func partition(total, batch int) []chunk {
	partitions := total / batch
	var chunks []chunk
	var i int
	for i = 0; i < partitions; i++ {
		chunks = append(chunks, chunk{start: i * batch, end: (i + 1) * batch})
	}
	if total%batch != 0 {
		chunks = append(chunks, chunk{start: i * batch, end: total})
	}
	return chunks
}

func (sink *MetriclyMetricsSink) send(elements []metricly_core.Element) error {
	return sink.client.PostElements(elements)
}

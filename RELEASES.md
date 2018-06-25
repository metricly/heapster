# Metrilcy Heapster Docker Container Release History
======================================

## Version 0.0.8
---------------------------
- fix a bug on negative computed cpu.usage.percent values
- fix a bug on 0 values for cpu.usage_rate for pod, namespace and cluster
- add a configurable cluster name in Metricly Sink to avoid element name collisions

## Version 0.0.7
---------------------------
- support creating computed metrics for cpu/memeory usage percent for all element types

## Version 0.0.6
---------------------------
- update default api uri to use kubernetes datasource
- fix element relationships broken by improved element type names

## Version 0.0.5
---------------------------
- improve element type names

## Version 0.0.4
---------------------------
- fix a bug when shortening a name that has no type prepended, i.e. 'cluster'

## Version 0.0.3
---------------------------
- add filter support to include/exclude metricsets by entity labels
- break concatenated custom labels to indivdual tags for better grouping elments
- shorten display element name by removing the types in its fqn

## Version 0.0.2
---------------------------
- add batch support to send element payload to Metricly

## Version 0.0.1
---------------------------
- initial release
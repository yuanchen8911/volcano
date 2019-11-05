/*
Copyright 2015 The Archimedes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metricsprovider

import (
	"archimedes-metrics/pkg/types"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/kubernetes/pkg/scheduler/cache"
	"time"
)

type MetricsClient interface {
	GetALlNodeMetrics(selector labels.Selector) (NodeMetricsCache, time.Time, error)
	GetNodeMetrics(name string) (*cache.Resource, error)
	GetAllNodeAllocatable() (NodeMetricsCache, time.Time, error)
	GetNodeAllocatable(name string) (*cache.Resource, error)

	GetAllNodeRIMetrics() (NodeMetricsCache, error)
	GetNodeRIMetrics(name string) (*types.NodeRIMetirc, error)
}

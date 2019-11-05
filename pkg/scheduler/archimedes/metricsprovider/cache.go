/*
Copyright 2018 The Archimedes Authors.

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
	"sync"

	"archimedes-metrics/pkg/types"
	"fmt"
	"k8s.io/kubernetes/pkg/scheduler/cache"
)

type nodeMetricsCache struct {
	lock              sync.RWMutex
	nodeMetricsInfo   map[string]*cache.Resource
	nodeRIMetricsInfo map[string]*types.NodeRIMetirc
}

type NodeMetricsCache interface {
	Get(string) (*cache.Resource, error)
	Set(string, *cache.Resource)
	Delete(string)
	GetRIMetric(name string) (*types.NodeRIMetirc, error)
	SetRIMetric(name string, resource *types.NodeRIMetirc)
	DeleteRIMetric(name string)
}

func NewNodeMetricsCache() NodeMetricsCache {
	return &nodeMetricsCache{
		nodeMetricsInfo:   map[string]*cache.Resource{},
		nodeRIMetricsInfo: map[string]*types.NodeRIMetirc{}}
}

func (c *nodeMetricsCache) Get(name string) (*cache.Resource, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	n, ok := c.nodeMetricsInfo[name]
	if !ok {
		return nil, fmt.Errorf("no metrics found for node: %v", name)
	}
	return n, nil
}

func (c *nodeMetricsCache) Set(name string, resource *cache.Resource) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.nodeMetricsInfo[name] = resource
}

func (c *nodeMetricsCache) Delete(name string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.nodeMetricsInfo, name)
}

func (c *nodeMetricsCache) GetRIMetric(name string) (*types.NodeRIMetirc, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	n, ok := c.nodeRIMetricsInfo[name]
	if !ok {
		return nil, fmt.Errorf("no ri metrics found for node: %v", name)
	}
	return n, nil
}

func (c *nodeMetricsCache) SetRIMetric(name string, resource *types.NodeRIMetirc) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.nodeRIMetricsInfo[name] = resource
}

func (c *nodeMetricsCache) DeleteRIMetric(name string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.nodeRIMetricsInfo, name)
}

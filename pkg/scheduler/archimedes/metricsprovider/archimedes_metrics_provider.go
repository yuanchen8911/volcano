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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"archimedes-metrics/pkg/types"
	"github.com/golang/glog"
	"io"
	"k8s.io/apimachinery/pkg/labels"
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"
	metricsapi "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"volcano.sh/volcano/cmd/scheduler/app/options"
)

const (
	ArchimedesMetricsProvider              = "archimedes"
	ArchimedesPredictNodeAllocatablePrefix = "predictNodeAllocatable-"
)

//func init() {
//	RegisterMetricsProviders(ArchimedesMetricsProvider)
//}

type ArchimedesMetricsClient struct {
	client            http.Client
	URL               string
	NodeMetricsAPI    string
	NodeAllocateAPI   string
	MetricsExpiration time.Duration
	nodeMetricsCache  NodeMetricsCache

	MetricsNodeRIEndpoint string
}

func NewArchimedesMetricsClient(options *options.ArchimedesMetricsOption) MetricsClient {
	return &ArchimedesMetricsClient{
		client: http.Client{
			Timeout: options.MetricsTimeout * time.Second,
		},
		URL:               options.MetricsServerURL,
		NodeMetricsAPI:    options.NodeMetricsAPI,
		NodeAllocateAPI:   options.NodeAllocateAPI,
		MetricsExpiration: options.MetricsExpiration,
		nodeMetricsCache:  NewNodeMetricsCache(),

		MetricsNodeRIEndpoint: options.MetricsNodeRIEndpoint,
	}
}

func (h *ArchimedesMetricsClient) GetAllNodeRIMetrics() (NodeMetricsCache, error) {
	metricPath := h.MetricsNodeRIEndpoint
	glog.V(4).Infof("Archimedes RI metrics url url: %s", metricPath)

	resultRaw, err := h.httpGet(metricPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod resource metrics: %v", err)
	}

	glog.V(10).Infof("Archimedes RI metrics result: %s", string(resultRaw))

	var metricsList []types.NodeRIMetirc
	err = json.Unmarshal(resultRaw, &metricsList)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal archimedes RI metrics response: %v", err)
	}

	if len(metricsList) == 0 {
		return nil, fmt.Errorf("no archimedes RI metrics returned from archimedes")
	}

	for _, m := range metricsList {
		glog.V(4).Infof("Add node metrics to Cache for node: %s,", m.ObjectMeta.Name)
		h.nodeMetricsCache.SetRIMetric(m.ObjectMeta.Name, &m)
	}

	return h.nodeMetricsCache, nil
}

func (h *ArchimedesMetricsClient) GetNodeRIMetrics(name string) (*types.NodeRIMetirc, error) {
	n, err := h.nodeMetricsCache.GetRIMetric(name)
	if err != nil {
		return nil, fmt.Errorf("no archimedes metrics found for node: %v，err: %v", name, err)
	}
	return n, nil
}

func (h *ArchimedesMetricsClient) GetALlNodeMetrics(selector labels.Selector) (NodeMetricsCache, time.Time, error) {
	params := map[string]string{"labelSelector": selector.String()}
	glog.V(4).Infof("Archimedes metrics url params: %s", params)

	resultRaw, err := h.httpGet(h.NodeMetricsAPI)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to get pod resource metrics: %v", err)
	}

	glog.V(10).Infof("Archimedes metrics result: %s", string(resultRaw))

	metricslist := metricsapi.NodeMetricsList{}
	err = json.Unmarshal([]byte(resultRaw), &metricslist)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to unmarshal archimedes response: %v", err)
	}

	if len(metricslist.Items) == 0 {
		return nil, time.Time{}, fmt.Errorf("no metrics returned from archimedes")
	}

	now := time.Now()
	for _, m := range metricslist.Items {
		if now.Sub(m.Timestamp.Time) > h.MetricsExpiration {
			h.nodeMetricsCache.Delete(m.ObjectMeta.Name)
			glog.V(4).Infof("Delete node metrics for node: %s,now: %s, timestamp: %s", m.ObjectMeta.Name, now, m.Timestamp.Time)
			continue
		}
		result := schedulercache.Resource{}
		result.Memory = m.Usage.Memory().Value()
		result.MilliCPU = m.Usage.Cpu().MilliValue()
		glog.V(4).Infof("Add node metrics to Cache for node: %s,", m.ObjectMeta.Name)
		h.nodeMetricsCache.Set(m.ObjectMeta.Name, &result)
	}

	timestamp := metricslist.Items[0].Timestamp.Time
	return h.nodeMetricsCache, timestamp, nil
}

func (h *ArchimedesMetricsClient) GetNodeMetrics(name string) (*schedulercache.Resource, error) {
	n, err := h.nodeMetricsCache.Get(name)
	if err != nil {
		return nil, fmt.Errorf("no archimedes metrics found for node: %v，err: %v", name, err)
	}
	return n, nil
}

func (h *ArchimedesMetricsClient) GetAllNodeAllocatable() (NodeMetricsCache, time.Time, error) {
	resultRaw, err := h.httpGet(h.NodeAllocateAPI)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to get node allocatable metrics: %v", err)
	}

	glog.V(10).Infof("Archimedes node allocatable metrics result: %s", string(resultRaw))

	items := NodeAllocatableMetricsList{}.Items
	err = json.Unmarshal([]byte(resultRaw), &items)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to unmarshal archimedes node allocatable response: %v", err)
	}

	if len(items) == 0 {
		return nil, time.Time{}, fmt.Errorf("no node allocatable metrics returned from archimedes")
	}

	now := time.Now()
	for _, m := range items {
		if now.Sub(m.Timestamp.Time) > h.MetricsExpiration {
			h.nodeMetricsCache.Delete(ArchimedesPredictNodeAllocatablePrefix + m.ObjectMeta.Name)
			continue
		}
		result := schedulercache.Resource{}
		result.Memory = m.Allocatable.Memory().Value()
		result.MilliCPU = m.Allocatable.Cpu().MilliValue()
		h.nodeMetricsCache.Set(ArchimedesPredictNodeAllocatablePrefix+m.ObjectMeta.Name, &result)
	}

	timestamp := items[0].Timestamp.Time
	return h.nodeMetricsCache, timestamp, nil
}

func (h *ArchimedesMetricsClient) GetNodeAllocatable(name string) (*schedulercache.Resource, error) {
	n, err := h.nodeMetricsCache.Get(ArchimedesPredictNodeAllocatablePrefix + name)
	if err != nil {
		return nil, fmt.Errorf("no archimedes allocatable metrics found for node: %v,err: %v", name, err)
	}
	return n, nil
}

func (h *ArchimedesMetricsClient) httpGet(action string) ([]byte, error) {
	url := h.URL + "/" + action
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		glog.Errorf("Http new request failed")
		return []byte{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		glog.Errorf("Http client failed")
		return []byte{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Http read body failed")
		return []byte{}, err
	}

	return body, nil
}

func (h *ArchimedesMetricsClient) httpPost(action string, requestBody io.Reader) (string, error) {
	url := h.URL + "/" + action
	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		glog.Errorf("Http new request failed")
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		glog.Errorf("Http client failed")
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Http read body failed")
		return "", err
	}

	return string(body), nil
}

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

package metrics

import (
	"github.com/golang/glog"
	"time"

	"archimedes-metrics/pkg/types"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"
	"volcano.sh/volcano/cmd/scheduler/app/options"
	"volcano.sh/volcano/pkg/scheduler/archimedes/metricsprovider"
)

var Metrics *ArchimedesMetrics

type ArchimedesMetrics struct {
	Ready                  bool
	nodeMetricsEnabled     bool
	nodeAllocatableEnabled bool
	metricsSyncPeriod      time.Duration
	EvaluateType           string
	MemUsageRatio          float64
	CpuUsageRatio          float64
	ReservedMemory         string
	ReservedCPU            int64
	MemAllocatableRatio    float32
	CpuAllocatableRatio    float32
	metricsClient          metricsprovider.MetricsClient

	DynamicRuntimeInfoScheduleEnabled bool
	CpuMaxUsageRatio                  float64
	MemMaxUsageRatio                  float64
}

func NewArchimedesMetrics(option *options.ArchimedesMetricsOption) {
	Metrics = &ArchimedesMetrics{
		Ready:                  false,
		nodeMetricsEnabled:     option.UseNodeMetrics,
		nodeAllocatableEnabled: option.UseNodeAllocatable,
		metricsSyncPeriod:      option.MetricsSyncPeriod,
		EvaluateType:           option.EvaluateType,
		MemUsageRatio:          option.MemUsageRatio,
		CpuUsageRatio:          option.CpuUsageRatio,
		ReservedMemory:         option.ReservedMemory,
		ReservedCPU:            option.ReservedCPU,
		MemAllocatableRatio:    option.MemAllocatableRatio,
		CpuAllocatableRatio:    option.CpuAllocatableRatio,

		DynamicRuntimeInfoScheduleEnabled: option.UseRuntimeInfoSchedule,
		CpuMaxUsageRatio:                  option.CpuMaxUsageRatio,
		MemMaxUsageRatio:                  option.MemMaxUsageRatio,
	}
	if option.UseNodeAllocatable || option.UseNodeMetrics {

		Metrics.metricsClient = metricsprovider.NewArchimedesMetricsClient(option)

		err := Metrics.startAllNodeMetrics()
		if err != nil {
			glog.V(3).Infof("scheduler app failed to start with metrics client: %v", err)
		} else {
			Metrics.Ready = true
			go wait.Until(Metrics.syncAllNodeMetrics, Metrics.metricsSyncPeriod, wait.NeverStop)
			go wait.Until(Metrics.syncAllNodeRiMetrics, Metrics.metricsSyncPeriod, wait.NeverStop)
		}
	}
}

func (am *ArchimedesMetrics) startAllNodeMetrics() error {

	if am.nodeMetricsEnabled {
		_, timestamp, err := am.GetALlNodeMetrics()
		if err != nil {
			glog.Errorf("get all node metrics failed, error : %s", err)
			return err
		}
		glog.V(4).Infof("get all node metrics success at time: %s", timestamp)
	}
	if am.nodeAllocatableEnabled {
		_, timestamp, err := am.GetAllNodeAllocatable()
		if err != nil {
			glog.Errorf("get all node allocatable metrics failed, error : %s", err)
			return err
		}
		glog.V(4).Infof("get all node allocatable metrics success at time: %s", timestamp)
	}
	if am.DynamicRuntimeInfoScheduleEnabled {
		_, err := am.GetAllNodeRINodeMetrics()
		if err != nil {
			glog.Errorf("get all node RI schedule metrics failed, error : %s", err)
			return err
		}
		glog.V(4).Infof("get all node RI schedule metrics success")
	}
	return nil
}

func (am *ArchimedesMetrics) syncAllNodeMetrics() {
	if am.nodeMetricsEnabled {
		am.metricsClient.GetALlNodeMetrics(labels.NewSelector())
	}
	if am.nodeAllocatableEnabled {
		am.metricsClient.GetAllNodeAllocatable()
	}
}

func (am *ArchimedesMetrics) syncAllNodeRiMetrics() {
	am.metricsClient.GetAllNodeRIMetrics()
	if am.DynamicRuntimeInfoScheduleEnabled {
		am.metricsClient.GetAllNodeRIMetrics()
	}
}

func (am *ArchimedesMetrics) GetALlNodeMetrics() (metricsprovider.NodeMetricsCache, time.Time, error) {
	return am.metricsClient.GetALlNodeMetrics(labels.NewSelector())
}

func (am *ArchimedesMetrics) GetNodeMetrics(name string) (*schedulercache.Resource, error) {
	return am.metricsClient.GetNodeMetrics(name)
}

func (am *ArchimedesMetrics) GetAllNodeAllocatable() (metricsprovider.NodeMetricsCache, time.Time, error) {
	return am.metricsClient.GetAllNodeAllocatable()
}

func (am *ArchimedesMetrics) GetNodeRIMetrics(name string) (*types.NodeRIMetirc, error) {
	return am.metricsClient.GetNodeRIMetrics(name)
}

func (am *ArchimedesMetrics) GetAllNodeRINodeMetrics() (metricsprovider.NodeMetricsCache, error) {
	return am.metricsClient.GetAllNodeRIMetrics()
}

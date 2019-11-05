package archimedesmetrics

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/resource"
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"
	"volcano.sh/volcano/pkg/scheduler/api"
	archimedes "volcano.sh/volcano/pkg/scheduler/archimedes"
	"volcano.sh/volcano/pkg/scheduler/framework"
)

type archimedesmetricsPlugin struct {
	pluginArguments framework.Arguments
}

func (pp *archimedesmetricsPlugin) Name() string {
	return "archimedesmetrics"
}

func New(arguments framework.Arguments) framework.Plugin {
	return &archimedesmetricsPlugin{
		pluginArguments: arguments,
	}
}

func (pp *archimedesmetricsPlugin) OnSessionOpen(ssn *framework.Session) {
	if archimedes.Metrics.Ready {
		for k, n := range ssn.Nodes {
			//k="11.5.83.231"
			nodeUsage, err := archimedes.Metrics.GetNodeMetrics(k)
			if err != nil {
				glog.Errorf("get node: %s metrics info failed, error : %s", n.Name, err)
			} else {
				calcRequestedResource(n, nodeUsage)
			}

		}
	}
}

func (pp *archimedesmetricsPlugin) OnSessionClose(ssn *framework.Session) {
}

func calcRequestedResource(nodeInfo *api.NodeInfo, nodeMetrics *schedulercache.Resource) {
	glog.V(5).Infof(" nodeInfo.Allocatable cpu %f mem %f", nodeInfo.Allocatable.MilliCPU, nodeInfo.Allocatable.Memory)
	glog.V(5).Infof(" nodeInfo.Idle cpu %f mem %f", nodeInfo.Idle.MilliCPU, nodeInfo.Idle.Memory)
	glog.V(5).Infof(" nodeInfo.Used cpu %f mem %f", nodeInfo.Used.MilliCPU, nodeInfo.Used.Memory)

	cpuUsage := nodeMetrics.MilliCPU
	memUsage := nodeMetrics.Memory
	reservedCPU := resource.NewQuantity(archimedes.Metrics.ReservedCPU, resource.DecimalSI)
	reservedMemory := resource.MustParse(archimedes.Metrics.ReservedMemory)

	requestedResource := nodeInfo.Used.Clone()

	if archimedes.Metrics.EvaluateType == "max" {
		requestedResource.MilliCPU = max(float64(cpuUsage)*(archimedes.Metrics.CpuUsageRatio)+float64(reservedCPU.MilliValue()), nodeInfo.Used.MilliCPU)
		requestedResource.Memory = max(float64(memUsage)*(archimedes.Metrics.MemUsageRatio)+float64(reservedMemory.Value()), nodeInfo.Used.Memory)

		requestedCPU := fmt.Sprintf("%f = max(%d * %f + %d, %f)", requestedResource.MilliCPU, cpuUsage, archimedes.Metrics.CpuUsageRatio, reservedCPU.MilliValue(), nodeInfo.Used.MilliCPU)
		requestedMemory := fmt.Sprintf("%f = max(%d * %f + %d, %f)", requestedResource.Memory, memUsage, archimedes.Metrics.MemUsageRatio, reservedMemory.Value(), nodeInfo.Used.Memory)
		glog.V(5).Infof("Node now requested: %s, cpu: %s, mem: %s", nodeInfo.Name, requestedCPU, requestedMemory)
	}
	if archimedes.Metrics.EvaluateType == "weighted" {
		requestedResource.MilliCPU = float64(cpuUsage)*(archimedes.Metrics.CpuUsageRatio) + float64(reservedCPU.MilliValue())
		requestedResource.Memory = float64(memUsage)*(archimedes.Metrics.MemUsageRatio) + float64(reservedMemory.Value())

		requestedCPU := fmt.Sprintf("%f = %d * %f + %d", requestedResource.MilliCPU, cpuUsage, archimedes.Metrics.CpuUsageRatio, reservedCPU.MilliValue())
		requestedMemory := fmt.Sprintf("%f = %d * %f + %d", requestedResource.Memory, memUsage, archimedes.Metrics.MemUsageRatio, reservedMemory.Value())
		glog.V(5).Infof("Node now requested: %s, cpu: %s, mem: %s", nodeInfo.Name, requestedCPU, requestedMemory)
	}
	nodeInfo.Used = requestedResource
	nodeInfo.Idle.Memory = nodeInfo.Allocatable.Memory - nodeInfo.Used.Memory
	nodeInfo.Idle.MilliCPU = nodeInfo.Allocatable.MilliCPU - nodeInfo.Used.MilliCPU
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

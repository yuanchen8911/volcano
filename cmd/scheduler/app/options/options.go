/*
Copyright 2017 The Kubernetes Authors.

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

package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

const (
	defaultSchedulerName   = "volcano"
	defaultSchedulerPeriod = time.Second
	defaultQueue           = "default"
	defaultListenAddress   = ":8080"

	defaultHealthzBindAddress = "127.0.0.1:11251"

	defaultQPS   = 50.0
	defaultBurst = 100

	// Default parameters to control the number of feasible nodes to find and score
	defaultMinPercentageOfNodesToFind = 5
	defaultMinNodesToFind             = 100
	defaultPercentageOfNodesToFind    = 100
)

// ServerOption is the main context object for the controller manager.
type ServerOption struct {
	Master               string
	Kubeconfig           string
	SchedulerName        string
	SchedulerConf        string
	SchedulePeriod       time.Duration
	EnableLeaderElection bool
	LockObjectNamespace  string
	DefaultQueue         string
	PrintVersion         bool
	ListenAddress        string
	EnablePriorityClass  bool
	KubeAPIBurst         int
	KubeAPIQPS           float32
	// HealthzBindAddress is the IP address and port for the health check server to serve on
	// defaulting to 127.0.0.1:11251
	HealthzBindAddress string

	// Parameters for scheduling tuning: the number of feasible nodes to find and score
	MinNodesToFind             int32
	MinPercentageOfNodesToFind int32
	PercentageOfNodesToFind    int32

	//JDOS runtime metrics
	ArchimedesMetricsOption ArchimedesMetricsOption
}

type ArchimedesMetricsOption struct {
	UseNodeMetrics     bool
	UseNodeAllocatable bool

	MetricsSyncPeriod time.Duration
	MetricsServerURL  string
	MetricsTimeout    time.Duration
	EvaluateType      string
	NodeMetricsAPI    string
	NodeAllocateAPI   string
	MetricsExpiration time.Duration

	MemUsageRatio float64
	CpuUsageRatio float64

	ReservedMemory string
	ReservedCPU    int64

	MemAllocatableRatio float32
	CpuAllocatableRatio float32

	UseRuntimeInfoSchedule bool
	MetricsNodeRIEndpoint  string
	CpuMaxUsageRatio       float64
	MemMaxUsageRatio       float64
}

// ServerOpts server options
var ServerOpts *ServerOption

// NewServerOption creates a new CMServer with a default config.
func NewServerOption() *ServerOption {
	s := ServerOption{}
	return &s
}

func (am *ArchimedesMetricsOption) AddFlags(fs *pflag.FlagSet) {
	fs.DurationVar(&am.MetricsSyncPeriod, "archimedes-metrics-sync-period", 120*time.Second, "The archimedes metrics sync period")
	fs.BoolVar(&am.UseNodeAllocatable, "archimedes-metrics-use-node-allocatable", false, "use predict node allocatable")
	fs.BoolVar(&am.UseNodeMetrics, "archimedes-metrics-use-node-metrics", false, "Use Node metrics for scheduling")
	fs.StringVar(&am.MetricsServerURL, "archimedes-metrics-url", "", "The archimedes metrics server url")
	fs.DurationVar(&am.MetricsTimeout, "archimedes-metrics-timeout", 60*time.Second, "The archimedes metrics time out")
	fs.Float64Var(&am.MemUsageRatio, "archimedes-memory-usage-ratio", 1.5, "Memory usage ratio")
	fs.Float64Var(&am.CpuUsageRatio, "archimedes-cpu-usage-ratio", 1.0, "CPU usage ratio")
	fs.StringVar(&am.ReservedMemory, "archimedes-node-reserved-memory", "2Gi", "The archimedes node reserved memory (Gi).")
	fs.Int64Var(&am.ReservedCPU, "archimedes-node-reserved-cpu", 0, "The archimedes node reserved cpu cores.")
	fs.StringVar(&am.EvaluateType, "archimedes-request-evaluate-type", "max", "Request Resource evaluate type, e.g.: max/weighted")
	fs.Float32Var(&am.MemAllocatableRatio, "archimedes-memory-allocatable-ratio", 1.0, "Memory allocatable ratio")
	fs.Float32Var(&am.CpuAllocatableRatio, "archimedes-cpu-allocatable-ratio", 1.0, "CPU allocatable ratio")
	fs.StringVar(&am.NodeMetricsAPI, "archimedes-metrics-node-metrics-api", "node/metrics", "The archimedes metrics root endpoint of metrics api")
	fs.StringVar(&am.NodeAllocateAPI, "archimedes-metrics-node-allocate-api", "node/allocatable", "The archimedes metrics of node allocate")
	fs.DurationVar(&am.MetricsExpiration, "archimedes-metrics-expiration", 300*time.Second, "The archimedes metrics expiration")

	fs.BoolVar(&am.UseRuntimeInfoSchedule, "archimedes-node-runtime-info-schedule-enabled", false, "use runtime info of node to schedule")
	fs.StringVar(&am.MetricsNodeRIEndpoint, "archimedes-metrics-runtimeinfo-endpoints", "/node/rimetrics",
		"The archimedes runtime info metrics endpoint")

}

// AddFlags adds flags for a specific CMServer to the specified FlagSet
func (s *ServerOption) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Master, "master", s.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	fs.StringVar(&s.Kubeconfig, "kubeconfig", s.Kubeconfig, "Path to kubeconfig file with authorization and master location information")
	// kube-batch will ignore pods with scheduler names other than specified with the option
	fs.StringVar(&s.SchedulerName, "scheduler-name", defaultSchedulerName, "kube-batch will handle pods whose .spec.SchedulerName is same as scheduler-name")
	fs.StringVar(&s.SchedulerConf, "scheduler-conf", "", "The absolute path of scheduler configuration file")
	fs.DurationVar(&s.SchedulePeriod, "schedule-period", defaultSchedulerPeriod, "The period between each scheduling cycle")
	fs.StringVar(&s.DefaultQueue, "default-queue", defaultQueue, "The default queue name of the job")
	fs.BoolVar(&s.EnableLeaderElection, "leader-elect", s.EnableLeaderElection,
		"Start a leader election client and gain leadership before "+
			"executing the main loop. Enable this when running replicated kube-batch for high availability")
	fs.BoolVar(&s.PrintVersion, "version", false, "Show version and quit")
	fs.StringVar(&s.LockObjectNamespace, "lock-object-namespace", s.LockObjectNamespace, "Define the namespace of the lock object that is used for leader election")
	fs.StringVar(&s.ListenAddress, "listen-address", defaultListenAddress, "The address to listen on for HTTP requests.")
	fs.BoolVar(&s.EnablePriorityClass, "priority-class", true,
		"Enable PriorityClass to provide the capacity of preemption at pod group level; to disable it, set it false")
	fs.Float32Var(&s.KubeAPIQPS, "kube-api-qps", defaultQPS, "QPS to use while talking with kubernetes apiserver")
	fs.IntVar(&s.KubeAPIBurst, "kube-api-burst", defaultBurst, "Burst to use while talking with kubernetes apiserver")
	fs.StringVar(&s.HealthzBindAddress, "healthz-bind-address", defaultHealthzBindAddress, "The address to listen on for /healthz HTTP requests.")

	// Minimum number of feasible nodes to find and score
	fs.Int32Var(&s.MinNodesToFind, "minimum-feasible-nodes", defaultMinNodesToFind, "The minimum number of feasible nodes to find and score")

	// Minimum percentage of nodes to find and score
	fs.Int32Var(&s.MinPercentageOfNodesToFind, "minimum-percentage-nodes-to-find", defaultMinPercentageOfNodesToFind, "The minimum percentage of nodes to find and score")

	// The percentage of nodes that would be scored in each scheduling cycle; if <= 0, an adpative percentage will be calcuated
	fs.Int32Var(&s.PercentageOfNodesToFind, "percentage-nodes-to-find", defaultPercentageOfNodesToFind, "The percentage of nodes to find and score, if <=0 will be calcuated based on the cluster size")

	s.ArchimedesMetricsOption.AddFlags(fs)
}

// CheckOptionOrDie check lock-object-namespace when LeaderElection is enabled
func (s *ServerOption) CheckOptionOrDie() error {
	if s.EnableLeaderElection && s.LockObjectNamespace == "" {
		return fmt.Errorf("lock-object-namespace must not be nil when LeaderElection is enabled")
	}

	return nil
}

// RegisterOptions registers options
func (s *ServerOption) RegisterOptions() {
	ServerOpts = s
}

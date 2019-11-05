package metricsprovider

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MetricsMeta struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// The following fields define time interval from which metrics were
	// collected from the interval [Timestamp-Window, Timestamp].
	Timestamp metav1.Time     `json:"timestamp"`
	Window    metav1.Duration `json:"window"`
}

type NodeMetrics struct {
	MetricsMeta `json:",inline"`
	RatedUtil   string          `json:"ratedUtil,omitempty"`
	Allocatable v1.ResourceList `json:"allocatable,omitempty"`
}

// NodeMetricsList is a list of NodeMetrics.
type NodeAllocatableMetricsList struct {
	// List of node metrics.
	Items []NodeMetrics `json:"items"`
}

package types

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

type RIUsage map[string]float64

type NodeRIMetirc struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// The following fields define time interval from which metrics were
	// collected from the interval [Timestamp-Window, Timestamp].
	Timestamp unversioned.Time `json:"timestamp"`

	// The memory usage is the memory working set.
	Usage map[string]RIUsage `json:"usage"`
}

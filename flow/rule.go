package flow

import (
	"encoding/json"
	"fmt"
)

const (
	LimitOriginDefault = "default"
	LimitOriginOther   = "other"
)

type MetricType int32

const (
	Concurrency MetricType = iota
	Qps
)

type RelationStrategy int32

const (
	Direct RelationStrategy = iota
	AssociatedResource
)

type ControlBehavior int32

const (
	Reject ControlBehavior = iota
	WarmUp
	Throttling
	WarmUpThrottling
)

type ClusterThresholdMode uint32

const (
	AvgLocalThreshold ClusterThresholdMode = iota
	GlobalThreshold
)

type ClusterRuleConfig struct {
	ThresholdType ClusterThresholdMode `json:"thresholdType"`
}

type FlowRule struct {
	Id uint64 `json:"id,omitempty"`

	Resource         string           `json:"resource"`
	LimitOrigin      string           `json:"limitApp"`
	MetricType       MetricType       `json:"grade"`
	Count            float64          `json:"count"`
	RelationStrategy RelationStrategy `json:"strategy"`
	ControlBehavior  ControlBehavior  `json:"controlBehavior"`

	RefResource       string            `json:"refResource,omitempty"`
	WarmUpPeriodSec   uint32            `json:"warmUpPeriodSec"`
	MaxQueueingTimeMs uint32            `json:"maxQueueingTimeMs"`
	ClusterMode       bool              `json:"clusterMode"`
	ClusterConfig     ClusterRuleConfig `json:"clusterConfig"`
}

func (f *FlowRule) String() string {
	b, err := json.Marshal(f)
	if err != nil {
		// Return the fallback string
		return fmt.Sprintf("FlowRule{resource=%s, id=%d, metricType=%d, threshold=%.2f",
			f.Resource, f.Id, f.MetricType, f.Count)
	}
	return string(b)
}

func (f *FlowRule) ResourceName() string {
	return f.Resource
}

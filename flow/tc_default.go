package flow

import (
	"github.com/sentinel-group/sentinel-golang/core"
)

type DefaultTrafficShapingCalculator struct {
	threshold float64
}

func NewDefaultTrafficShapingCalculator(threshold float64) *DefaultTrafficShapingCalculator {
	return &DefaultTrafficShapingCalculator{threshold: threshold}
}

func (d *DefaultTrafficShapingCalculator) CalculateAllowedTokens(core.StatNode, uint32, int32) float64 {
	return d.threshold
}

type DefaultTrafficShapingChecker struct {
	metricType MetricType
}

func NewDefaultTrafficShapingChecker(metricType MetricType) *DefaultTrafficShapingChecker {
	return &DefaultTrafficShapingChecker{metricType: metricType}
}

func (d *DefaultTrafficShapingChecker) DoCheck(node core.StatNode, acquireCount uint32, threshold float64) *core.TokenResult {
	if node == nil {
		return core.NewTokenResultPass()
	}
	var curCount float64
	if d.metricType == Concurrency {
		curCount = float64(node.CurrentGoroutineNum())
	} else {
		curCount = node.PassQPS()
	}
	if curCount+float64(acquireCount) > threshold {
		return core.NewTokenResultBlocked(core.BlockTypeFlow, "Flow")
	}
	return core.NewTokenResultPass()
}

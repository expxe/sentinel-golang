package flow

import "github.com/sentinel-group/sentinel-golang/core"

type TrafficShapingCalculator interface {
	CalculateAllowedTokens(node core.StatNode, acquireCount uint32, flag int32) float64
}

type TrafficShapingChecker interface {
	DoCheck(node core.StatNode, acquireCount uint32, threshold float64) *core.TokenResult
}

type TrafficShapingController struct {
	flowCalculator TrafficShapingCalculator
	flowChecker    TrafficShapingChecker

	rule *FlowRule
}

func NewTrafficShapingController(flowCalculator TrafficShapingCalculator, flowChecker TrafficShapingChecker, rule *FlowRule) *TrafficShapingController {
	return &TrafficShapingController{flowCalculator: flowCalculator, flowChecker: flowChecker, rule: rule}
}

func (t *TrafficShapingController) Rule() *FlowRule {
	return t.rule
}

func (t *TrafficShapingController) FlowChecker() TrafficShapingChecker {
	return t.flowChecker
}

func (t *TrafficShapingController) FlowCalculator() TrafficShapingCalculator {
	return t.flowCalculator
}

func (t *TrafficShapingController) CanPass(node core.StatNode, acquireCount uint32, flag int32) *core.TokenResult {
	allowedTokens := t.flowCalculator.CalculateAllowedTokens(node, acquireCount, flag)
	return t.flowChecker.DoCheck(node, acquireCount, allowedTokens)
}

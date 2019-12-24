package flow

import (
	"github.com/sentinel-group/sentinel-golang/core"
	"github.com/sentinel-group/sentinel-golang/statistic"
	"time"
)

type FlowSlot struct {
}

func (s *FlowSlot) Check(ctx *core.EntryContext) *core.TokenResult {
	tcs := getTrafficControllerListFor(ctx.Resource.Name())
	if len(tcs) == 0 {
		return core.NewTokenResultPass()
	}

	// Check rules in order
	for _, tc := range tcs {
		if tc == nil {
			// TODO: should warn or panic?
			continue
		}
		r := canPassCheck(tc, ctx.StatNode, ctx.Input.AcquireCount)
		if r.Status() == core.ResultStatusBlocked {
			return r
		}
		if r.Status() == core.ResultStatusShouldWait {
			if waitMs := r.WaitMs(); waitMs > 0 {
				// Handle waiting action.
				time.Sleep(time.Duration(waitMs) * time.Millisecond)
			}
			continue
		}
	}
	return core.NewTokenResultPass()
}

func canPassCheck(tc *TrafficShapingController, node core.StatNode, acquireCount uint32) *core.TokenResult {
	return canPassCheckWithFlag(tc, node, acquireCount, 0)
}

func canPassCheckWithFlag(tc *TrafficShapingController, node core.StatNode, acquireCount uint32, flag int32) *core.TokenResult {
	if tc.rule.ClusterMode {
		// TODO: support cluster mode
	}
	return checkInLocal(tc, node, acquireCount, flag)
}

func selectNodeByRelStrategy(rule *FlowRule, node core.StatNode) core.StatNode {
	if rule.RelationStrategy == AssociatedResource {
		return statistic.GetResourceNode(rule.RefResource)
	}
	return node
}

func checkInLocal(tc *TrafficShapingController, node core.StatNode, acquireCount uint32, flag int32) *core.TokenResult {
	actual := selectNodeByRelStrategy(tc.rule, node)
	if actual == nil {
		return core.NewTokenResultPass()
	}
	return tc.CanPass(node, acquireCount, flag)
}

package flow

import (
	"github.com/pkg/errors"
	"github.com/sentinel-group/sentinel-golang/datasource"
	"github.com/sentinel-group/sentinel-golang/log"
	"reflect"
	"sync"
)

// const
var (
	logger = log.GetDefaultLogger()

	ruleListener = &flowRulePropertyListener{}
)

type TrafficControllerGenFunc func(*FlowRule) *TrafficShapingController

type TrafficControllerMap map[string][]*TrafficShapingController

var (
	tcGenFuncMap = make(map[ControlBehavior]TrafficControllerGenFunc)
	tcMap        = make(TrafficControllerMap, 0)
	tcMux        = new(sync.RWMutex)

	ruleProperty datasource.PropertyPublisher = datasource.NewSentinelProperty()
	propertyInit sync.Once
)

func init() {
	propertyInit.Do(func() {
		ruleProperty.AddListener(ruleListener)
	})

	// Initialize the traffic shaping controller generator map for existing control behaviors.
	tcGenFuncMap[Reject] = func(rule *FlowRule) *TrafficShapingController {
		return NewTrafficShapingController(NewDefaultTrafficShapingCalculator(rule.Count), NewDefaultTrafficShapingChecker(rule.MetricType), rule)
	}
}

func LoadRules(rules []*FlowRule) (bool, error) {
	return ruleProperty.UpdateValue(rules, 0)
}

func SetTrafficShapingGenerator(cb ControlBehavior, generator TrafficControllerGenFunc) error {
	if generator == nil {
		return errors.New("nil generator")
	}
	if cb >= Reject && cb <= WarmUpThrottling {
		return errors.New("not allowed to replace the generator for default control behaviors")
	}
	tcMux.Lock()
	defer tcMux.Unlock()

	tcGenFuncMap[cb] = generator
	return nil
}

func RemoveTrafficShapingGenerator(cb ControlBehavior) error {
	if cb >= Reject && cb <= WarmUpThrottling {
		return errors.New("not allowed to replace the generator for default control behaviors")
	}
	tcMux.Lock()
	defer tcMux.Unlock()

	delete(tcGenFuncMap, cb)
	return nil
}

func RegisterRuleProperty(p datasource.PropertyPublisher) error {
	if p == nil {
		return errors.New("nil PropertyPublisher is not allowed")
	}
	if p == ruleProperty {
		return nil
	}
	tcMux.Lock()
	defer tcMux.Unlock()

	ruleProperty.RemoveListener(ruleListener)
	p.AddListener(ruleListener)
	ruleProperty = p
	return nil
}

func getTrafficControllerListFor(name string) []*TrafficShapingController {
	tcMux.RLock()
	defer tcMux.RUnlock()

	return tcMap[name]
}

type flowRulePropertyListener struct {
}

func (l *flowRulePropertyListener) OnConfigUpdate(value interface{}, flag int32) error {
	rules, ok := value.([]*FlowRule)
	if !ok {
		// nil is also not allowed here
		return errors.New("invalid type, expected []FlowRule but got: " + reflect.TypeOf(value).Name())
	}
	tcMux.Lock()
	defer tcMux.Unlock()

	m := buildFlowMap(rules)

	// TODO: check potential leak here?
	tcMap = m
	return nil
}

// NotThreadSafe
func buildFlowMap(rules []*FlowRule) TrafficControllerMap {
	if len(rules) == 0 {
		return make(TrafficControllerMap, 0)
	}
	m := make(TrafficControllerMap, 0)
	for _, rule := range rules {
		if !IsValidFlowRule(rule) {
			logger.Warnf("Ignoring invalid flow rule: %v", rule)
			continue
		}
		if rule.LimitOrigin == "" {
			rule.LimitOrigin = LimitOriginDefault
		}
		generator, supported := tcGenFuncMap[rule.ControlBehavior]
		if !supported {
			logger.Warnf("Ignoring the rule due to unsupported control behavior: %v", rule)
			continue
		}
		tsc := generator(rule)

		rulesOfRes, exists := m[rule.Resource]
		if !exists {
			m[rule.Resource] = []*TrafficShapingController{tsc}
		} else {
			m[rule.Resource] = append(rulesOfRes, tsc)
		}
	}
	return m
}

func IsValidFlowRule(rule *FlowRule) bool {
	if rule == nil || rule.Resource == "" || rule.Count < 0 {
		return false
	}
	if rule.MetricType < 0 || rule.RelationStrategy < 0 || rule.ControlBehavior < 0 {
		return false
	}

	if rule.RelationStrategy == AssociatedResource && rule.RefResource == "" {
		return false
	}

	return checkClusterField(rule) && checkControlBehaviorField(rule)
}

func checkClusterField(rule *FlowRule) bool {
	if rule.ClusterMode && rule.Id <= 0 {
		return false
	}
	return true
}

func checkControlBehaviorField(rule *FlowRule) bool {
	switch rule.ControlBehavior {
	case WarmUp:
		return rule.WarmUpPeriodSec > 0
	case Throttling:
		return true
	case WarmUpThrottling:
		return rule.WarmUpPeriodSec > 0 && rule.MaxQueueingTimeMs > 0
	default:
		return true
	}
}

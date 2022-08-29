package fctrl

import (
	"github.com/pkg/errors"
	"github.com/tigbox/fctrl/internal/lib"
)

// LoadConfig important
func LoadConfig(resourceConfig *ResourceConfig, opts ...OptLimiter) (ILimiter, error) {
	if resourceConfig == nil {
		return nil, errors.New("ResourceConfig is nil")
	}

	// fields字段组合哈希map
	fHashMap := make(map[string][]string)
	// 规则哈希map
	rHashMap := make(map[string]strategy)
	// 字段哈希到规则哈希数组
	fieldRuleHashMap := make(map[string][]string)
	for _, rule := range resourceConfig.Rules {
		fields := rule.getSortFields()
		fieldsHash := lib.GetStrListHash(fields)
		if len(fields) > 0 {
			if _, exists := fHashMap[fieldsHash]; !exists {
				fHashMap[fieldsHash] = fields
			}

			if _, exists := fieldRuleHashMap[fieldsHash]; !exists {
				fieldRuleHashMap[fieldsHash] = make([]string, 0)
			}
		}

		// 将滑动窗口时间字符串转为对应的毫秒值 int64 类型
		var period int64
		period, err := timeStringToMilliSecond(rule.Period)
		if err != nil {
			return nil, err
		}

		strategy := newStrategy(rule.Name, period, int64(rule.Threshold), rule.Code)
		strategyHash := strategy.getHash()

		if _, exists := rHashMap[strategyHash]; !exists {
			rHashMap[strategyHash] = strategy
		}

		if _, ok := fieldRuleHashMap[fieldsHash]; ok {
			if !lib.ContainInStringSlice(fieldRuleHashMap[fieldsHash], strategyHash) {
				fieldRuleHashMap[fieldsHash] = append(fieldRuleHashMap[fieldsHash], strategyHash)
			}
		}

	}

	var strategySets []strategySet
	for k, v := range fieldRuleHashMap {
		var strategies []strategy
		if fields, ok := fHashMap[k]; ok {
			for _, rHash := range v {
				if strategy, exists := rHashMap[rHash]; exists {
					strategies = append(strategies, strategy)
				}
			}
			if len(strategies) > 0 {
				ss := newStrategySet(fields, k, strategies)
				strategySets = append(strategySets, ss)
			}
		}
	}

	// limiter基本的构造函数
	limiter := newLimiter(resourceConfig.Resource, strategySets, resourceConfig.RecordFields, opts...)
	return limiter, nil
}

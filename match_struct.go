package fctrl

// 频率控制规则
type MatchDetail struct {
	RuleName  string
	Period    int64
	Threshold int64
	DataList  DataList
}

// 频率控制返回结果
type MatchResult struct {
	Code int64
	Data interface{}
}

func defaultMatchResult() *MatchResult {
	return &MatchResult{
		Code: 0,
		Data: struct{}{},
	}
}

package fctrl

import (
	"fmt"

	"github.com/tigbox/fctrl/internal/lib"
)

type strategy struct {
	Name      string
	Period    int64
	Threshold int64
	Code      int64
}

func newStrategy(name string, period int64, threshold int64, code int64) strategy {
	return strategy{
		Name:      name,
		Period:    period,
		Threshold: threshold,
		Code:      code,
	}
}

func (s *strategy) toString() string {
	if s == nil {
		return ""
	}
	result := fmt.Sprintf("Name:%v", s.Name)
	result += fmt.Sprintf(" Period:%v", s.Period)
	result += fmt.Sprintf(" Threshold:%v", s.Threshold)
	result += fmt.Sprintf(" Code:%v", s.Code)
	return result
}

func (s *strategy) getHash() string {
	if s == nil {
		return ""
	}
	return lib.GetMD5Bytes([]byte(s.toString()))
}

package fctrl

import "context"

type ILimiter interface {
	FrequencyControl(ctx context.Context, entry *Entry) (matchResult *MatchResult, err error)
}

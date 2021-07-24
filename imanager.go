package fctrl

import "context"

type IManager interface {
	FrequencyControl(ctx context.Context, entity Entity) (MatchResult, error)
}

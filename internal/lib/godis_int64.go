package lib

import (
	"fmt"
	"runtime"

	"github.com/tigbox/godis/interface/redis"
	"github.com/tigbox/godis/redis/reply"
)

func GodisReply2int64(rr redis.Reply) (int64, error) {
	intResult, ok := rr.(*reply.IntReply)
	if !ok {
		return 0, fmt.Errorf("expected int64 reply, but acutally %s,%s", rr.ToBytes(), printStack())
	}
	return intResult.Code, nil
}

func printStack() string {
	_, file, no, ok := runtime.Caller(2)
	if ok {
		return fmt.Sprintf("at %s#%d", file, no)
	}
	return ""
}

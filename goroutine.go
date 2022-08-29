package fctrl

import (
	"context"
	"fmt"
	"runtime"
)

func RecoverGoroutineFunc(ctx context.Context, f func()) {
	go func() {
		defer Recover(ctx)
		f()
	}()
}

func Recover(_ context.Context) {
	err := recover()
	if err != nil {
		for i := 1; i < 64; i++ {
			_, file, line, ok := runtime.Caller(i)
			if ok {
				Loger.Error("stack_index", i, file, fmt.Sprintf("%v:%v", file, line))
			} else {
				break
			}
		}
	}

}

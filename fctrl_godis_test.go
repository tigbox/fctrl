package fctrl

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gogf/gf/util/gconv"
	"github.com/stretchr/testify/assert"
	"github.com/tigbox/fctrl/internal/lib"
	"github.com/tigbox/godis/lib/utils"
)

var godisLimiter ILimiter
var godisOnceInit sync.Once

func LoadConfigByGodisMode() {
	godisOnceInit.Do(func() {
		// 资源线规则
		resourceConfig := mockResourceConfig()
		var err error
		godisLimiter, err = LoadConfig(resourceConfig)
		if err != nil {
			panic(err)
		}
		if godisLimiter == nil {
			panic("testGlobalLimiter is nil")
		}
	})
}

// 基础测试 单机写入有明细返回的模式
func BenchmarkGodisWriteWithDetial(b *testing.B) {
	LoadConfigByGodisMode()
	var err error
	for n := 0; n < b.N; n++ {
		// 构造测试数据
		input := generateInput()
		entry := NewEntry(input, OptionMode(WriteMode_ResponseWithDetail))
		_, err = godisLimiter.FrequencyControl(context.Background(), entry)
		if err != nil {
			panic(err)
		}
	}
}

// 基础测试 单机写入无明细返回的模式
func BenchmarkGodisWriteWithoutDetial(b *testing.B) {
	LoadConfigByGodisMode()
	var err error
	for n := 0; n < b.N; n++ {
		// 构造测试数据
		input := generateInput()
		entry := NewEntry(input, OptionMode(WriteMode_ResponseWithoutDetail))
		_, err = godisLimiter.FrequencyControl(context.Background(), entry)
		if err != nil {
			panic(err)
		}
	}
}

// 基础测试 单机只读有明细返回的模式
func BenchmarkGodisOnlyReadWithDetial(b *testing.B) {
	LoadConfigByGodisMode()

	// 因为是写入模式，构造一些数据
	debugGodisDB.Flush()
	writeGoids()

	var err error
	for n := 0; n < b.N; n++ {
		// 构造测试数据
		input := generateInput()
		entry := NewEntry(input, OptionMode(ReadMode_ResponseWithDetail))
		_, err = godisLimiter.FrequencyControl(context.Background(), entry)
		if err != nil {
			panic(err)
		}
	}
}

// 基础测试 单机只读无明细返回的模式
func BenchmarkGodisOnlyReadWithoutDetial(b *testing.B) {
	LoadConfigByGodisMode()

	// 因为是写入模式，构造一些数据
	debugGodisDB.Flush()
	writeGoids()

	var err error
	for n := 0; n < b.N; n++ {
		// 构造测试数据
		input := generateInput()
		entry := NewEntry(input, OptionMode(ReadMode_ResponseWithoutDetail))
		_, err = godisLimiter.FrequencyControl(context.Background(), entry)
		if err != nil {
			panic(err)
		}
	}
}

// 写入模式并且返回的时候不附带明细
func TestFCGodisWithWriteMode(t *testing.T) {
	modes := []int{WriteMode_ResponseWithDetail, WriteMode_ResponseWithoutDetail}
	for _, mode := range modes {
		fmt.Printf("-------start testing mode=%d\n", mode)
		// 写入模式
		FCtrlModeByGodis(t, mode)
		debugGodisDB.Flush() // 需要刷新数据
	}
}

func TestFCGodisWithReadMode(t *testing.T) {
	modes := []int{ReadMode_ResponseWithDetail, ReadMode_ResponseWithoutDetail}
	for _, mode := range modes {
		fmt.Printf("-------start testing mode=%d\n", mode)
		// 读取模式，每次每种模式都需要先构造数据
		writeGoids()
		FCtrlModeByGodis(t, mode)
		debugGodisDB.Flush()
	}

}

// 构造数据的部分
func writeGoids() {
	key := "fc:test:15000:ip#uid:114.19.201.23#456"
	currentTimestamp := TimeStamp()
	for i := 0; i < 20; i++ {
		temp := currentTimestamp - int64(i)
		score := gconv.String(temp)
		member := gconv.String(map[string]interface{}{
			"a":  "this is value a",
			"b":  "this is value b",
			"ts": temp,
		})
		// write
		debugGodisDB.Exec(nil, utils.ToCmdLine("zadd", key, score, member))
	}

}

// 写入模式并且命中的时候返回明细
func FCtrlModeByGodis(t *testing.T, mode int) {
	LoadConfigByGodisMode()
	testMode := mode

	for i := 0; i < 20; i++ {
		// 构造测试数据
		input := generateInput()
		entry := NewEntry(input, OptionMode(testMode))

		mResult, err := godisLimiter.FrequencyControl(context.Background(), entry)
		assert.Nil(t, err)
		time.Sleep(1000 * time.Millisecond)
		fmt.Printf("第%d次请求\n", i)

		if mResult.Code != 0 {
			fmt.Printf("%#v\n", mResult)
		}

		debug := false
		if debug {
			// for debug
			currentKey := "fc:test:15000:ip#uid:114.19.201.23#456"
			temp := debugGodisDB.Exec(nil, utils.ToCmdLine("zrange", currentKey, "0", "-1", "WITHSCORES"))
			fmt.Println("----------")
			zs, err := lib.Bytes2redisZ(temp.ToBytes())
			if err != nil {
				spew.Dump(err)
			}
			if len(zs) > 0 {
				spew.Dump(zs)
			}
			fmt.Printf("当前key下有%d个member\n", len(zs))
			fmt.Println("==========")
		}
	}

}

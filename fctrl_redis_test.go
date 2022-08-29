package fctrl

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

var redisLimiter ILimiter
var redisOnceInit sync.Once

func LoadConfigByRedisMode() {
	redisOnceInit.Do(func() {
		var err error
		options := &redis.Options{
			Addr:         "127.0.0.1:6379",
			Password:     "",
			DB:           0,
			PoolSize:     16,
			DialTimeout:  100 * time.Millisecond,
			WriteTimeout: 500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
		}
		redisCLient := redis.NewClient(options)
		// 资源线规则
		resourceConfig := mockResourceConfig()
		redisLimiter, err = LoadConfig(resourceConfig, OptRedisClient(redisCLient))
		if err != nil {
			panic(err)
		}
	})
}

// 基础测试 redis主从版写入有明细返回的模式
func BenchmarkRedisWriteWithDetail(b *testing.B) {
	LoadConfigByRedisMode()
	var err error
	for n := 0; n < b.N; n++ {
		// 构造测试数据
		input := generateInput()
		entry := NewEntry(input, OptionMode(WriteMode_ResponseWithDetail))
		_, err = redisLimiter.FrequencyControl(context.Background(), entry)
		if err != nil {
			panic(err)
		}
	}
}

// 基础测试 redis主从版写入无明细返回的模式
func BenchmarkRedisWriteWithoutDetail(b *testing.B) {
	LoadConfigByRedisMode()
	var err error
	for n := 0; n < b.N; n++ {
		// 构造测试数据
		input := generateInput()
		entry := NewEntry(input, OptionMode(WriteMode_ResponseWithoutDetail))
		_, err = redisLimiter.FrequencyControl(context.Background(), entry)
		if err != nil {
			panic(err)
		}
	}
}

// 基础测试 redis主从版只读有明细返回的模式
func BenchmarkRedisReadWithDetail(b *testing.B) {
	LoadConfigByRedisMode()
	var err error
	for n := 0; n < b.N; n++ {
		// 构造测试数据
		input := generateInput()
		entry := NewEntry(input, OptionMode(ReadMode_ResponseWithDetail))
		_, err = redisLimiter.FrequencyControl(context.Background(), entry)
		if err != nil {
			panic(err)
		}
	}
}

// 基础测试 redis主从版只读无明细返回的模式
func BenchmarkRedisReadWithoutDetail(b *testing.B) {
	LoadConfigByRedisMode()
	var err error
	for n := 0; n < b.N; n++ {
		// 构造测试数据
		input := generateInput()
		entry := NewEntry(input, OptionMode(ReadMode_ResponseWithoutDetail))
		_, err = redisLimiter.FrequencyControl(context.Background(), entry)
		if err != nil {
			panic(err)
		}
	}
}

func TestFCRedisWithWriteMode(t *testing.T) {
	LoadConfigByRedisMode()
	modes := []int{WriteMode_ResponseWithDetail, WriteMode_ResponseWithoutDetail}
	for _, mode := range modes {
		fmt.Printf("-------start testing mode=%d\n", mode)
		// 写入模式，不需要构造数据
		FCtrlModeByRedis(t, mode)
	}
}

func FCtrlModeByRedis(t *testing.T, mode int) {
	testMode := mode

	for i := 0; i < 20; i++ {
		// 构造测试数据
		input := generateInput()
		entry := NewEntry(input, OptionMode(testMode))

		mResult, err := redisLimiter.FrequencyControl(context.Background(), entry)
		assert.Nil(t, err)
		time.Sleep(1000 * time.Millisecond)
		fmt.Printf("第%d次请求\n", i)

		if mResult.Code != 0 {
			fmt.Printf("%#v\n", mResult)
		}

	}
}

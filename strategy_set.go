package fctrl

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/gogf/gf/util/gconv"
	"github.com/pkg/errors"
	"github.com/tigbox/fctrl/internal/lib"
	"github.com/tigbox/godis/lib/utils"
)

type strategySet struct {
	Fields       []string
	fieldsHash   *string
	maxPeriod    int64
	maxThreshold int64
	Strategies   []strategy
}

// newStrategySet 创建规则集
func newStrategySet(fields []string, fieldHash string, strategies []strategy) strategySet {
	object := strategySet{Fields: fields, Strategies: strategies}
	if fieldHash != "" {
		object.fieldsHash = &fieldHash
	}
	(&object).format()
	return object
}

// format format StrategySet
func (ss *strategySet) format() {
	for _, rule := range ss.Strategies {
		if rule.Period > ss.maxPeriod {
			ss.maxPeriod = rule.Period
		}
		if rule.Threshold > ss.maxThreshold {
			ss.maxThreshold = rule.Threshold
		}
	}
}

// isFit 当前规则集的字段是一个数组， 如果这个数组中的全部元素都包含在传入的map数组中， 说明规则集应该
func (ss *strategySet) isFit(inputData map[string]interface{}) bool {
	if ss == nil {
		return false
	}
	for _, fieldName := range ss.Fields {
		if _, exists := inputData[fieldName]; !exists {
			return false
		}
	}
	return true
}

// getCurrentFields 获取当前规则集的字段字符串
func (ss *strategySet) getCurrentFields() string {
	return strings.Join(ss.Fields, redis_key_inner_split)
}

// prepareRedisKey 准备redis的key
func (ss *strategySet) prepareRedisKey(
	_ context.Context,
	resourceName string,
	inputData map[string]interface{},
) string {
	var result string
	var buf bytes.Buffer

	// 拼凑fcData
	normalizedData := make([]string, len(ss.Fields))
	for index, field := range ss.Fields {
		normalizedData[index] = fmt.Sprintf("%v", inputData[field])
	}
	fcData := strings.Join(normalizedData, redis_key_inner_split)

	//前缀:资源线:保留时长:组合id:fc数据
	buf.WriteString(redis_key_prefix) //前缀
	buf.WriteString(redis_key_split)
	buf.WriteString(resourceName) //业务资源名称
	buf.WriteString(redis_key_split)
	buf.WriteString(strconv.FormatInt(ss.maxPeriod, 10)) //最大保留时长
	buf.WriteString(redis_key_split)
	buf.WriteString(ss.getCurrentFields()) //字段组合str ，当前ruleSet对应的Fields是排好序的
	buf.WriteString(redis_key_split)
	buf.WriteString(fcData) // 对应的字段value组合成的一段string

	result = buf.String()
	return result
}

func (ss *strategySet) writeRedis(ctx context.Context, limiter *Limiter, entry *Entry, memberData string) (err error) {
	if ss == nil {
		err = errors.New("strategySet is nil")
		return
	}

	// 1. 判断输入内容是否有规定的字段
	if strong_match {
		if !ss.isFit(entry.Input) {
			return
		}
	}

	// 2. prepare redis key
	redisKey := ss.prepareRedisKey(ctx, limiter.name, entry.Input)

	// 3. prepare member with score
	// redis sorted set，可以简单理解成跳表。每个节点由score和member组成
	scoreMember := redis.Z{Score: float64(entry.Ts), Member: memberData}

	// 4. ZAdd operation
	err = ss.zAdd(ctx, redisKey, scoreMember, limiter)
	if err != nil {
		return
	}
	RecoverGoroutineFunc(ctx, func() {
		// 5. expire operation with new goroutine
		ss.expire(ctx, redisKey, limiter)
		// 6. remove member with new goroutine
		lowBoundInt64 := getLowerBound(TimeStamp(), ss.maxPeriod)
		remBefore := strconv.FormatInt(lowBoundInt64, 10)
		//fmt.Printf("redisKey = %v,ts = %v, remBefore = %v,duration=%v\n",redisKey,entry.Ts,remBefore,entry.Ts-lowBoundInt64)

		ss.zRemRangeByScore(ctx, redisKey, "-inf", remBefore, limiter)
	})

	return
}

func (ruleSet *strategySet) zAdd(_ context.Context, redisKey string, member redis.Z, limiter *Limiter) (err error) {
	switch limiter.redisMode {
	case godisMode:
		limiter.godisDB.Exec(nil, utils.ToCmdLine2("zadd", redisKey, gconv.String(member.Score), gconv.String(member.Member)))
	case redisMode:
		_, err = limiter.redisClient.ZAdd(redisKey, member).Result()
	case redisClusterMode:
		_, err = limiter.redisClusterClient.ZAdd(redisKey, member).Result()
	}

	if err != nil {
		err = errors.Wrapf(err, "ZAdd failed,[key]%v", redisKey)
		return
	}

	return
}

func (ss *strategySet) expire(_ context.Context, redisKey string, limiter *Limiter) {
	var err error
	switch limiter.redisMode {
	case godisMode:
		// godis
		limiter.godisDB.Expire(redisKey, time.Now().Add(time.Duration(ss.maxPeriod)*time.Millisecond))
	case redisMode:
		// redis
		_, err = limiter.redisClient.PExpire(redisKey, time.Duration(ss.maxPeriod)*time.Millisecond).Result()
	case redisClusterMode:
		// redis cluster
		_, err = limiter.redisClusterClient.PExpire(redisKey, time.Duration(ss.maxPeriod)*time.Millisecond).Result()
	default:
		err = fmt.Errorf("key=%v, redisMode=%v", redisKey, limiter.redisMode)
	}
	if err != nil {
		Loger.Warnf("PExpire failed,[key]%v [epiration]%v", redisKey, ss.maxPeriod)
		return
	}

}

func (ss *strategySet) zRemRangeByScore(_ context.Context, redisKey string, min, max string, limiter *Limiter) {
	switch limiter.redisMode {
	case godisMode:
		limiter.godisDB.Exec(nil, utils.ToCmdLine("ZRemRangeByScore", redisKey, min, max))
	case redisMode:
		limiter.redisClient.ZRemRangeByScore(redisKey, min, max).Result()
	case redisClusterMode:
		limiter.redisClusterClient.ZRemRangeByScore(redisKey, min, max).Result()
	}

}

////////// read options ////////
func (ss *strategySet) readRedis(ctx context.Context,
	limiter *Limiter,
	entry *Entry,
	upBound int64,
	lowBound int64) (dataList DataList, err error) {
	if ss == nil {
		err = errors.New("RuleSet is nil")
		return
	}
	if !ss.isFit(entry.Input) {
		return
	}

	// 1. redis key
	redisKey := ss.prepareRedisKey(ctx, limiter.name, entry.Input)
	// 2. zrangeby 查询条件
	upperBoundStr := strconv.FormatInt(upBound, 10)
	lowerBoundStr := strconv.FormatInt(lowBound, 10)
	zRangeBy := redis.ZRangeBy{Min: lowerBoundStr, Max: upperBoundStr, Count: ss.maxThreshold}

	var readRedisResult []redis.Z
	switch limiter.redisMode {
	case godisMode:
		// 注意godis使用的是命令，ZRevRangeByScore，要先max后min
		godiReply := limiter.godisDB.Exec(nil, utils.ToCmdLine("ZRevRangeByScore", redisKey, upperBoundStr, lowerBoundStr, "WithScores"))
		readRedisResult, err = lib.Bytes2redisZ(godiReply.ToBytes())
	case redisMode:
		readRedisResult, err = limiter.redisClient.ZRevRangeByScoreWithScores(redisKey, zRangeBy).Result()
	case redisClusterMode:
		readRedisResult, err = limiter.redisClusterClient.ZRevRangeByScoreWithScores(redisKey, zRangeBy).Result()
	}

	if err != nil {
		return DataList{Data{}}, err
	}
	dataList = redisToDataList(readRedisResult)

	return
}

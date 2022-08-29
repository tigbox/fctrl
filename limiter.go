package fctrl

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/tigbox/fctrl/internal/lib"
	"github.com/tigbox/godis"
	"github.com/tigbox/godis/lib/utils"
)

type Limiter struct {
	dependence   // 依赖的一些存储方面的东西
	name         string
	strategySets []strategySet
	recordFields []string // 要额外记录的字段
}

type dependence struct {
	redisMode          int                  // 0/1/2 分别表示 单机版/redis/redis集群版
	redisClient        *redis.Client        // redis实例
	redisClusterClient *redis.ClusterClient // redis集群实例
	godisDB            *godis.DB
}

func newDependenceByGoids() dependence {
	// 默认是godis， 即本机模式
	return dependence{
		redisMode: godisMode,
		godisDB:   godis.MakeDB(),
	}
}

type OptLimiter func(limiter *Limiter)

func OptRedisClient(rdsClient *redis.Client) OptLimiter {
	return func(limiter *Limiter) {
		limiter.redisMode = redisMode
		limiter.redisClient = rdsClient
	}
}

func OptRedisClusterClient(rdsClusterClient *redis.ClusterClient) OptLimiter {
	return func(limiter *Limiter) {
		limiter.redisMode = redisClusterMode
		limiter.redisClusterClient = rdsClusterClient
	}
}

func newLimiter(resourceName string, strategySets []strategySet, records []string, opts ...OptLimiter) ILimiter {
	limiter := &Limiter{
		name:         resourceName,
		strategySets: strategySets,
		recordFields: records,
		dependence:   dependence{},
	}

	if len(opts) > 0 {
		for _, optFunc := range opts {
			optFunc(limiter)
		}
	}

	// 既不是redis主从模式也不是redis集群模式的话就搞成godis模式
	if limiter.redisMode != redisMode && limiter.redisMode != redisClusterMode {
		limiter.dependence = newDependenceByGoids()
		// for debug
		debugGodisDB = limiter.dependence.godisDB
	}

	return limiter
}

var debugGodisDB *godis.DB

func (limiter *Limiter) FrequencyControl(ctx context.Context, entry *Entry) (matchResult *MatchResult, err error) {
	// 1、检查并格式化Entry
	err = entry.checkAndFormat(ctx)
	if err != nil {
		return nil, err
	}

	// 2、如果是写入的模式，那么先进行写入
	if lib.ContainIntSlice(writeModes, entry.Mode) {
		err = limiter.writeRedis(ctx, entry)
		if err != nil {
			Loger.Warn(err.Error())
		}
	}

	// 3. 读取数据
	matchResult, err = limiter.readRedis(ctx, entry)
	return
}

func (limiter *Limiter) makeMemberData(ctx context.Context, entry *Entry) (result string, err error) {
	data := limiter.filterRecordData(entry.Input)
	_, exists := data[TS]
	if len(data) == 0 || !exists {
		// 如果以后有链路追踪的traceID的话，那么这个TS最好变成trace_id。不过现在并不支持，因为基础架构并不存在traceID。
		// 要支持traceID的话，需要对context做一些封装，并且保证每个服务都会调用封装好的context
		data[TS] = entry.Ts
	}
	memberBytes, err := json.Marshal(data)
	if err != nil {
		return
	}
	result = string(memberBytes[:])
	return
}

func (limiter *Limiter) filterRecordData(inputData map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	if limiter.recordFields != nil && len(limiter.recordFields) > 0 {
		for _, field := range limiter.recordFields {
			if val, ok := inputData[field]; ok {
				result[field] = val
			}
		}
	}
	return result
}

func (limiter *Limiter) writeRedis(ctx context.Context, entry *Entry) (err error) {
	if limiter == nil || entry == nil {
		err = errors.New("limiter or entry is nil")
		return
	}
	// 构造memberData
	memberData, err := limiter.makeMemberData(ctx, entry)
	if err != nil {
		return
	}

	// concurrent write
	wg := sync.WaitGroup{}
	wg.Add(len(limiter.strategySets))

	for _, item := range limiter.strategySets {
		ss := item
		RecoverGoroutineFunc(ctx, func() {
			defer wg.Done()
			err = ss.writeRedis(ctx, limiter, entry, memberData)
			if err != nil {
				return
			}
		})
	}
	wg.Wait()
	return
}

func (limiter *Limiter) readRedis(ctx context.Context, entry *Entry) (*MatchResult, error) {
	// 默认都走带详情的模式
	// 这两种模式并没有明显的优劣之分
	// detail mode是在多规则单一字段集的情况会和redis交互一次，优点是交互少，缺点是可能数据量较大从而占用网络IO
	// no detail mode是一个规则就会和redis交互一次，将count操作交给了redis，优点是对外传输数据量较小，但是如果规则多会引发redis的CPU负载升高。
	switch entry.Mode {
	case WriteMode_ResponseWithDetail:
		fallthrough
	case ReadMode_ResponseWithDetail:
		return limiter.responseWithDetail(ctx, entry)
	case WriteMode_ResponseWithoutDetail:
		fallthrough
	case ReadMode_ResponseWithoutDetail:
		return limiter.responseWithoutDetail(ctx, entry)
	default:
		return limiter.responseWithDetail(ctx, entry)
	}
}

func (limiter *Limiter) responseWithDetail(ctx context.Context, entry *Entry) (*MatchResult, error) {
	matchResult := defaultMatchResult()
	var err error
	if limiter == nil || entry == nil {
		err := errors.New("limiter or entry is nil")
		return nil, err
	}
	upperBound := entry.Ts
	// 遍历每个RuleSet
	for _, ruleSet := range limiter.strategySets {
		if !ruleSet.isFit(entry.Input) {
			continue
		}
		lowerBound := getLowerBound(upperBound, ruleSet.maxPeriod)
		var dataList DataList
		dataList, err = ruleSet.readRedis(ctx, limiter, entry, upperBound, lowerBound)
		if err != nil {
			return nil, err
		}

		// 遍历RuleSet下的Rule
		for _, rule := range ruleSet.Strategies {
			lowerRuleBound := getLowerBound(upperBound, rule.Period)
			var ruleDataList DataList
			for _, fcData := range dataList {
				if lowerRuleBound <= fcData.Score && fcData.Score <= upperBound {
					ruleDataList = append(ruleDataList, fcData)

					// 判断是否达到阈值
					if ruleDataList.isReachThreshold(rule.Threshold) {
						matchDetail := MatchDetail{rule.Name, rule.Period, rule.Threshold, ruleDataList}
						matchResult = &MatchResult{Code: rule.Code, Data: matchDetail}
						return matchResult, nil
					}
				}
			}
		}
	}
	return matchResult, err
}

// responseWithoutDetail ...
func (limiter *Limiter) responseWithoutDetail(ctx context.Context, entry *Entry) (matchResult *MatchResult, err error) {
	upperBound := entry.Ts
	matchResult = &MatchResult{}

	for _, ruleSet := range limiter.strategySets {
		if !ruleSet.isFit(entry.Input) {
			continue
		}
		redisKey := ruleSet.prepareRedisKey(ctx, limiter.name, entry.Input)

		for _, rule := range ruleSet.Strategies {
			lowerBound := getLowerBound(upperBound, rule.Period)
			upperBoundStr := strconv.FormatInt(upperBound, 10)
			lowerBoundStr := strconv.FormatInt(lowerBound, 10)
			var currentCount int64

			switch limiter.redisMode {
			case godisMode:
				godisReply := limiter.godisDB.Exec(nil, utils.ToCmdLine("zcount", redisKey, lowerBoundStr, upperBoundStr))
				currentCount, err = lib.GodisReply2int64(godisReply)
			case redisMode:
				// 主从版
				currentCount, err = limiter.redisClient.ZCount(redisKey, lowerBoundStr, upperBoundStr).Result()
			case redisClusterMode:
				// 集群版
				currentCount, err = limiter.redisClusterClient.ZCount(redisKey, lowerBoundStr, upperBoundStr).Result()
			}

			if err != nil {
				err = errors.Wrapf(err, "ZCount failed,[key]%v", redisKey)
				return
			}

			if reachThreshold(currentCount, rule.Threshold) {
				type matchDetail struct {
					RuleName  string
					Period    int64
					Threshold int64
				}
				md := matchDetail{
					RuleName:  rule.Name,
					Period:    rule.Period,
					Threshold: rule.Threshold,
				}
				matchResult = &MatchResult{Code: rule.Code, Data: md}
				return
			}
		}
	}
	return
}

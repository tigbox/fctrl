package fctrl

import (
	"sort"

	"github.com/go-redis/redis"
)

type Data struct {
	Score  int64       `json:"ts"`
	Detail interface{} `json:"detail"`
}

type DataList []Data

func (list DataList) Len() int           { return len(list) }
func (list DataList) Less(i, j int) bool { return list[i].Score < list[j].Score }
func (list DataList) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }

func rankByScoreDesc(data map[int64]interface{}) DataList {
	result := make(DataList, len(data))
	i := 0
	for k, v := range data {
		result[i] = Data{k, v}
		i++
	}
	// rank by score desc
	sort.Sort(sort.Reverse(result))
	return result
}

func (dataList *DataList) isReachThreshold(threshold int64) bool {
	return reachThreshold(int64(len(*dataList)), threshold)
}

func reachThreshold(currentCount, threshold int64) bool {
	// 很多人理解的阈值都不同，有的是大于有的是大于等于，在这里进行统一
	return currentCount > threshold
}

func convertFromZ(zs []redis.Z) DataList {
	res := make(map[int64]interface{})
	for _, item := range zs {
		res[int64(item.Score)] = item.Member
	}
	return rankByScoreDesc(res)
}

func redisToDataList(zs []redis.Z) DataList {
	result := make(DataList, len(zs))
	for k, v := range zs {
		result[k].Score = int64(v.Score)
		result[k].Detail = v.Member
	}
	return result
}

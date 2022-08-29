package fctrl

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	//频率控制支持的时间粒度
	MILLI  = 1e6
	SECOND = 1e9 //1e9是秒 ， 不要超过秒

	PRECISION             = MILLI                             //当前频率数据的时间粒度
	RULE_PERIOD_PRECISION = MILLI                             //规则数据中支持的滑动窗口的时间单位为毫秒
	RATE                  = RULE_PERIOD_PRECISION / PRECISION //比例尺：要转成当前时间粒度毫秒的比例尺，那就是1000.
)

func getLowerBound(upperBound int64, period int64) (result int64) {
	result = upperBound - convertTimeAccuracy(period)
	return
}

func convertTimeAccuracy(period int64) int64 {
	return period * RATE
}

func TimeStamp(times ...time.Time) (result int64) {
	if len(times) > 0 {
		result = times[0].UnixNano() / PRECISION
	} else {
		result = time.Now().UnixNano() / PRECISION
	}
	return
}

// timeStringToMilliSecond 涉及到正则，只在初始化频控规则的时候用，其他的地方慎用
func timeStringToMilliSecond(str string) (s int64, err error) {
	// 字符串拆分
	var numList []string
	var unitList []string

	numRe := regexp.MustCompile("[0-9]+")
	numList = numRe.FindAllString(str, -1)

	unitRe := regexp.MustCompile("[^0-9]+")
	unitList = unitRe.FindAllString(str, -1)

	if len(numList) == 0 || len(unitList) == 0 || len(numList) != len(unitList) {
		err = errors.New("period format error")
		return
	}

	for i, unit := range unitList {
		var timeNum int64
		timeNum, err = strconv.ParseInt(numList[i], 10, 64)
		if err != nil {
			return
		}
		switch {
		case strings.EqualFold(unit, "ms"):
			s += timeNum
		case strings.EqualFold(unit, "s"):
			s += timeNum * 1000
		case strings.EqualFold(unit, "min"):
			s += timeNum * 60000
		case strings.EqualFold(unit, "h"):
			s += timeNum * 3600000
		case strings.EqualFold(unit, "d"):
			s += timeNum * 86400000
		default:
			err = errors.New("period format error")
		}
	}

	return
}

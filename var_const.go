package fctrl

const (
	// 目前用是见
	TS = "ts" // 在这个阶段代表member里面的时间戳，如果有traceID的话更好，可惜现在并不支持，要是traceID，需要封装context
)

const (
	godisMode        = 0 // godis模式是0
	redisMode        = 1 // redis模式是1
	redisClusterMode = 2 // redis模式是2
)

const (
	// (写入，不写入) * (返回明细，不返回明细) 总共有4种模式
	WriteMode_ResponseWithDetail    = 0  //写入模式，返回明细， 默认也是这种情况
	WriteMode_ResponseWithoutDetail = 1  //写入模式，不返回明细
	ReadMode_ResponseWithDetail     = -1 //读取模式，返回明细
	ReadMode_ResponseWithoutDetail  = -2 //读取模式，不返回明细
)

const (
	redis_key_prefix      = "fc"
	redis_key_split       = ":"
	redis_key_inner_split = "#"

	// 是否强匹配，指的是当规定的规则里面包含a和b，可是传入的数据却不包含a和b，在强匹配的情况就不进行写入了，因为写入了也不会有频率控制命中的情况
	// 如果强匹配是false，当前规则里面包含a和b字段，即便传入的数据不包含a和b，依旧会进行写入
	// 默认强匹配
	// 目前不对外开放
	strong_match = true
)

var (
	// 目前支持的4种模式
	SupportModes = []int{
		WriteMode_ResponseWithDetail,    // 先写入后读取，返回明细
		WriteMode_ResponseWithoutDetail, // 先写入后读取，不返回明细
		ReadMode_ResponseWithDetail,     // 不进行写入直接读取，返回明细
		ReadMode_ResponseWithoutDetail,  // 不进行写入直接读取，不返回明细
	}

	// 写入的时候支持的模式，是为了判断是否就需要进行写入
	writeModes = []int{
		WriteMode_ResponseWithDetail,
		WriteMode_ResponseWithoutDetail,
	}

	// 默认的模式
	default_mode = WriteMode_ResponseWithDetail
)

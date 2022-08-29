package lib

import (
	"bytes"

	"github.com/go-redis/redis"
	"github.com/tigbox/fctrl/internal/proto"
)

func Bytes2redisZ(byteSlice []byte) ([]redis.Z, error) {
	rd := proto.NewReader(bytes.NewReader(byteSlice))
	result, err := reader2zs(rd)
	return result, err
}

func reader2zs(rd *proto.Reader) ([]redis.Z, error) {
	var err error
	var v interface{}
	v, err = rd.ReadArrayReply(zSliceParser)
	if err != nil {
		return v.([]redis.Z), err
	}

	return v.([]redis.Z), nil
}

func zSliceParser(rd *proto.Reader, n int64) (interface{}, error) {
	zz := make([]redis.Z, n/2)
	for i := int64(0); i < n; i += 2 {
		var err error

		z := &zz[i/2]

		z.Member, err = rd.ReadString()
		if err != nil {
			return nil, err
		}

		z.Score, err = rd.ReadFloatReply()
		if err != nil {
			return nil, err
		}
	}
	return zz, nil
}

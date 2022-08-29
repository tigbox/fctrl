package lib

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func ContainInStringSlice(source []string, find string) (result bool) {
	for _, item := range source {
		if item == find {
			result = true
			break
		}
	}
	return
}

func ContainIntSlice(source []int, find int) (result bool) {
	for _, item := range source {
		if item == find {
			result = true
			break
		}
	}
	return
}

func GetMD5Bytes(bytes []byte) string {
	hashObject := md5.New()
	hashObject.Write(bytes)
	return hex.EncodeToString(hashObject.Sum(nil))
}

func GetStrListHash(strList []string) string {
	str := strings.Join(strList, ",")
	return GetMD5Bytes([]byte(str))
}

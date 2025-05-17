package utils

import (
	"strconv"
)

// StringToInt 将字符串转换为整数，出错时返回默认值0
func StringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// StringToInt64 将字符串转换为int64，出错时返回默认值0
func StringToInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

// IntToString 将整数转换为字符串
func IntToString(i int) string {
	return strconv.Itoa(i)
}

// Int64ToString 将int64转换为字符串
func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

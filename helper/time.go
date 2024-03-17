package helper

import "time"

// TimestampToTime 将毫秒级时间戳转换为time.Time类型
func TimestampToTime(msTimestamp int64) time.Time {
	// 将毫秒转换为秒
	seconds := msTimestamp / 1000
	// 提取剩下的毫秒，并将其转换为纳秒（1毫秒 = 1,000,000纳秒）
	nanoseconds := (msTimestamp % 1000) * 1000000

	// 使用Unix函数将秒和纳秒转换为time.Time
	return time.Unix(seconds, nanoseconds)
}

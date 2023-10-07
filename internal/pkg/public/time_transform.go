package public

import (
	"time"
)

// 将时间戳转换为时间格式字符串
func TimestampToString(timestamp int64, layout string) string {
	return time.Unix(timestamp, 0).Format(layout)
}

// 将时间字符串转换为时间戳
func StringToTimestamp(str string, layout string) (int64, error) {
	t, err := time.ParseInLocation(layout, str, time.Local)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// 将时间字符串转换为时间对象
func StringToTime(str string, layout string) (time.Time, error) {
	return time.ParseInLocation(layout, str, time.Local)
}

// 将时间对象转换为时间字符串
func TimeToString(t time.Time, layout string) string {
	return t.Format(layout)
}

// 将时间对象转换为时间戳
func TimeToTimestamp(t time.Time) int64 {
	return t.Unix()
}

// 将时间戳转换为时间对象
func TimestampToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

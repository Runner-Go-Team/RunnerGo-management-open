package handler

import (
	"fmt"
	"strings"
)

// removeTopStruct 去除提示信息中的结构体名称
func removeTopStruct(fields map[string]string) string {
	res := map[string]string{}
	for field, err := range fields {
		res[field[strings.Index(field, ".")+1:]] = err
	}

	// 将键值对连接为字符串
	var sb strings.Builder
	for _, v := range res {
		sb.WriteString(fmt.Sprintf("%s,", v))
	}

	result := sb.String()
	// 去除末尾的逗号
	if len(result) > 0 {
		result = result[:len(result)-1]
	}

	return result
}

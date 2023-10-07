package public

import (
	"reflect"
)

// GetStringNum 获取字符串字符个数
func GetStringNum(stringData string) int {
	num := 0
	for range stringData {
		num++
	}
	return num
}

// CheckStructIsEmpty 判断结构体是否为空
func CheckStructIsEmpty(obj interface{}) bool {
	// 获取结构体的反射值
	value := reflect.ValueOf(obj)
	// 获取结构体的反射类型
	typ := value.Type()

	// 如果传入的不是结构体类型，则认为不为空
	if typ.Kind() != reflect.Struct {
		return false
	}

	// 遍历结构体的每个字段，判断是否有值
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		// 如果有任何一个字段有值，就认为结构体不为空
		if !field.IsZero() {
			return false
		}
	}

	return true
}

// ReverseSlice 切片翻转
func ReverseSlice[T any](s []T) []T {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// SliceUnique 切片去重通过map键的唯一性去重
func SliceUnique[T any](s []T) []T {
	result := make([]T, 0, len(s))

	m := map[any]struct{}{}
	for _, v := range s {
		if _, ok := m[v]; !ok {
			result = append(result, v)
			m[v] = struct{}{}
		}
	}

	return result
}

// ContainsStringSlice 切片中是否包含改元素
func ContainsStringSlice(s []string, elem string) bool {
	for _, a := range s {
		if a == elem {
			return true
		}
	}
	return false
}

// SliceDiff 两个切片的差集
func SliceDiff[T any](a []T, b []T) []T {
	var c []T
	temp := map[any]struct{}{}

	for _, val := range b {
		if _, ok := temp[val]; !ok {
			temp[val] = struct{}{}
		}
	}
	for _, val := range a {
		if _, ok := temp[val]; !ok {
			c = append(c, val)
		}
	}

	return c
}

func ContainsSliceDiff(a, b []string) []string {
	var c []string
	for _, itemA := range a {
		for _, itemB := range b {
			if itemA == itemB {
				c = append(c, itemA)
				break
			}
		}
	}
	return c
}

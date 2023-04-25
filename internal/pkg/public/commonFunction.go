package public

import "reflect"

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

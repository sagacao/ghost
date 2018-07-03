package utils

import "strconv"

func GetInterfaceUint32(str string, qurey map[string]interface{}) uint32 {
	if qurey == nil {
		return 0
	}
	value, err := qurey[str]
	if !err {
		return 0
	}
	val, ok := value.(uint32)
	if ok {
		return val
	}
	return 0
}

func GetInterfaceString(str string, qurey map[string]interface{}) string {
	if qurey == nil {
		return "0"
	}
	value, err := qurey[str]
	if !err {
		return "0"
	}
	val, ok := value.(string)
	if ok {
		return val
	}
	return "0"
}

func ToUint32(str string) uint32 {
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return uint32(value)
}

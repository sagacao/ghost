package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

func FormatMapOfStrings(m map[string]string) string {
	s := fmt.Sprintf("%#v", m)
	s = strings.Replace(s, "\\\"", "\"", -1)
	s = strings.Replace(s, "\"{", "{", -1)
	s = strings.Replace(s, "}\"", "}", -1)
	s = strings.Replace(s, "\\n", "", -1)
	s = strings.Replace(s, "\\t", "", -1)
	return strings.Trim(s, "map[string]")
}

func CatchError(flag string, err error) bool {
	return strings.Contains(err.Error(), flag)
}

func DeepCopy(value interface{}) interface{} {
	if valueMap, ok := value.(map[string]interface{}); ok {
		newMap := make(map[string]interface{})
		for k, v := range valueMap {
			newMap[k] = DeepCopy(v)
		}

		return newMap
	} else if valueSlice, ok := value.([]interface{}); ok {
		newSlice := make([]interface{}, len(valueSlice))
		for k, v := range valueSlice {
			newSlice[k] = DeepCopy(v)
		}

		return newSlice
	}

	return value
}

func WeaklyString(value interface{}) (out string) {
	switch value.(type) {
	case int:
		out = fmt.Sprintf("%v", value)
	default:
		out = fmt.Sprintf("unknown type")
	}

	return
}

func SplitModel(model string) string {
	ret := ""
	counter := 0
	value := strings.Split(model, ";")
	for _, v := range value {
		if counter < 3 {
			ret += (v + ";")
		}
		counter++
	}
	return ret
}

func Md5sum(str []byte) string {
	l := md5.New()
	l.Write(str)
	return hex.EncodeToString(l.Sum(nil))
}

func SqlStringCheck(input string) string {
	output := strings.Replace(input, "'", "''", -1)
	return output;
}

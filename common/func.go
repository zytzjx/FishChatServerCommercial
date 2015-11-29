package common

//数组中是否包含某个元素
func InArray(arr interface{}, element interface{}) bool {

	switch arr.(type) {
	case []string:
		for _, v := range arr.([]string) {
			if v == element.(string) {
				return true
			}
		}
	default:
	}
	return false
}

//从字符数组中删除
func DeleteChild(arr []string, element string) []string {
	temp := make([]string, 0)
	for _, v := range arr {
		if v == element {
			continue
		}
		temp = append(temp, v)
	}
	return temp
}

func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}

	return string(rs[start:end])
}

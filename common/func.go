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

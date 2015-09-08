package helper

func CopyMap(old map[string]string) map[string]string{
	newMap := map[string]string{}
	for k,v := range old {
		newMap[k] = v
	}
	return newMap
}
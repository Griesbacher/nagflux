package modGearman

import (
	"io/ioutil"
	"strings"
)

//DefaultModGearmanKeyLength length of an gearman key
const DefaultModGearmanKeyLength = 32

//GetSecret parses the mod_gearman secret/file and returns one key.
func GetSecret(secret, secretFile string) string {
	if secret != "" {
		return secret
	}
	if secretFile != "" {
		if data, err := ioutil.ReadFile(secretFile); err != nil {
			panic(err)
		} else {
			return strings.TrimSpace(string(data))
		}
	}
	return ""
}

//ShapeKey expands the key to length, or cuts it.
func ShapeKey(key string, length int) []byte {
	for i := 0; i <= length-len(key); i++ {
		key = key + string([]rune{'\x00'})
	}
	return []byte(key)[:length]
}

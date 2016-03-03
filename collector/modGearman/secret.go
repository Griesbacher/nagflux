package modGearman

import "io/ioutil"

//GetSecret parses the mod_gearman secret/file and returns one key.
func GetSecret(secret, secretFile string) string {
	if secret != "" {
		return secret
	}
	if secretFile != "" {
		if data, err := ioutil.ReadFile(secretFile); err != nil {
			panic(err)
		} else {
			return string(data)
		}
	}
	return ""
}

//FillKey expands the key to length.
func FillKey(key string, length int) []byte {
	return []byte(key) //TODO: to implement
}

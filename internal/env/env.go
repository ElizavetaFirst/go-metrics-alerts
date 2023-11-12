package env

import (
	"os"
	"strconv"
)

func GetEnvDuration(key string, defaultVal int) int {
	if envVal, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(envVal); err == nil {
			return i
		}
	}
	return defaultVal
}

func GetEnvString(key, defaultVal string) string {
	if envVal, exists := os.LookupEnv(key); exists {
		return envVal
	}
	return defaultVal
}

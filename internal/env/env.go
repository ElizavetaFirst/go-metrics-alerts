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

func GetEnvBool(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		boolVal, err := strconv.ParseBool(value)
		if err == nil {
			return boolVal
		}
	}
	return defaultVal
}

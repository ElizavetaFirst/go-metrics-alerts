package env

import(
	"os"
	"strconv"
	"time"
)

func GetEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if envVal, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(envVal); err == nil {
			return time.Duration(i) * time.Second
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
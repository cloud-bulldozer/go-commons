package utils

import (
	"os"
	"strings"
)

// EnvToMap returns the host environment variables as a map
func EnvToMap() map[string]interface{} {
	envMap := make(map[string]interface{})
	for _, v := range os.Environ() {
		envVar := strings.SplitN(v, "=", 2)
		envMap[envVar[0]] = envVar[1]
	}
	return envMap
}
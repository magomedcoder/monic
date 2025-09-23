package pkg

import (
	"os"
	"strconv"
)

func GetEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

func MustInt(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}

	return def
}

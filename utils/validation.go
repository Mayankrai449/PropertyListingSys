package utils

import (
	"strconv"
	"strings"
)

func IsValidExternalID(id string) bool {
	if !strings.HasPrefix(id, "PROP") {
		return false
	}
	numStr := strings.TrimPrefix(id, "PROP")
	num, err := strconv.Atoi(numStr)
	if err != nil || num < 1000 {
		return false
	}
	return true
}

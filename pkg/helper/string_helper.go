package helper

import (
	"strconv"
	"strings"
)

func StringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func TrimLecturerName(name string) string {
	if len(name) <= 3 {
		return name
	}

	// Split by spaces to handle multiple names
	names := strings.Split(name, " ")
	var result []string

	for _, n := range names {
		if len(n) <= 3 {
			result = append(result, n)
			continue
		}

		// Take first 3 chars and add asterisks
		first := n[:3]
		padding := ""
		for i := 0; i < len(n)-3; i++ {
			padding += "*"
		}
		result = append(result, first+padding)
	}

	return strings.Join(result, " ")
}

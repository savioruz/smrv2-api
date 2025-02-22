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

	// Split by possible lecturer separators
	lecturers := strings.Split(name, "Oktira")
	if len(lecturers) == 1 {
		return trimSingleLecturer(name)
	}

	// Handle multiple lecturers case
	var trimmedNames []string
	for _, lecturer := range lecturers {
		trimmedNames = append(trimmedNames, trimSingleLecturer(strings.TrimSpace(lecturer)))
	}
	return strings.Join(trimmedNames, " & ")
}

func trimSingleLecturer(name string) string {
	// Handle titles and prefixes
	name = strings.TrimPrefix(name, "Prof. ")
	name = strings.TrimPrefix(name, "Dr., ")
	name = strings.TrimPrefix(name, "Drs. ")
	name = strings.TrimPrefix(name, "Hj. ")
	name = strings.TrimPrefix(name, "H. ")

	// Get first 3 chars of actual name
	first := name[:3]

	// Find first comma, period or space
	var endIndex int
	for i := 3; i < len(name); i++ {
		if name[i] == ',' || name[i] == ' ' || name[i] == '.' {
			endIndex = i
			break
		}
	}

	// If no separator found, use full length
	if endIndex == 0 {
		endIndex = len(name)
	}

	// Create asterisk padding
	padding := ""
	for i := 0; i < endIndex-3; i++ {
		padding += "*"
	}

	// Combine first + padding + rest of string
	return first + padding + name[endIndex:]
}

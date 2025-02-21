package helper

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type TimeRange struct {
	StartTime string
	EndTime   string
}

var timeRanges = []TimeRange{
	{StartTime: "07:00", EndTime: "07:50"},
	{StartTime: "07:50", EndTime: "08:40"},
	{StartTime: "08:45", EndTime: "09:35"},
	{StartTime: "09:35", EndTime: "10:25"},
	{StartTime: "10:30", EndTime: "11:20"},
	{StartTime: "11:20", EndTime: "12:10"},
	{StartTime: "12:30", EndTime: "13:20"},
	{StartTime: "13:20", EndTime: "14:10"},
	{StartTime: "14:15", EndTime: "15:05"},
	{StartTime: "15:15", EndTime: "16:05"},
	{StartTime: "16:10", EndTime: "17:00"},
	{StartTime: "17:00", EndTime: "17:50"},
	{StartTime: "18:30", EndTime: "19:20"},
	{StartTime: "19:20", EndTime: "20:10"},
	{StartTime: "20:10", EndTime: "21:00"},
	{StartTime: "16:00", EndTime: "16:50"},
	{StartTime: "16:50", EndTime: "17:40"},
	{StartTime: "17:40", EndTime: "18:30"},
}

func CalculateTimeRange(times string) (time.Time, time.Time, error) {
	var startTime, endTime time.Time
	timeSlots := strings.Split(times, ",")

	if len(timeSlots) > 0 {
		// Get first slot for start time
		firstSlot := strings.TrimSpace(timeSlots[0])
		slotIndex, err := strconv.Atoi(firstSlot)
		if err != nil {
			return startTime, endTime, fmt.Errorf("invalid time slot format: %v", err)
		}

		if slotIndex < 1 || slotIndex > len(timeRanges) {
			return startTime, endTime, fmt.Errorf("time slot index out of range")
		}

		// Get last slot for end time
		lastSlot := strings.TrimSpace(timeSlots[len(timeSlots)-1])
		lastIndex, err := strconv.Atoi(lastSlot)
		if err != nil {
			return startTime, endTime, fmt.Errorf("invalid time slot format: %v", err)
		}

		if lastIndex < 1 || lastIndex > len(timeRanges) {
			return startTime, endTime, fmt.Errorf("time slot index out of range")
		}

		// Parse the start time of first slot
		startTime, err = time.Parse("15:04", timeRanges[slotIndex-1].StartTime)
		if err != nil {
			return startTime, endTime, fmt.Errorf("error parsing start time: %v", err)
		}

		// Parse the end time of last slot
		endTime, err = time.Parse("15:04", timeRanges[lastIndex-1].EndTime)
		if err != nil {
			return startTime, endTime, fmt.Errorf("error parsing end time: %v", err)
		}
	}

	return startTime, endTime, nil
}

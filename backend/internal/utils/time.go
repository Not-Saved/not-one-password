package utils

import "time"

func SecondsUntilTime(t time.Time) int {
	return int(time.Until(t).Seconds())
}

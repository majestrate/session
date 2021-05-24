package utils

import "time"

func TimeNow() int64 {
	now := time.Now()
	return now.Unix() * 1000
}

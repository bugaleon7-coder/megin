package utils

import "time"

// TimePtr 返回 time.Time 的指针
func TimePtr(t time.Time) *time.Time {
	return &t
}
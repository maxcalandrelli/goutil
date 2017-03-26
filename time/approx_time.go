package gu_time

import (
	"time"
)

func Round(t time.Time, units time.Duration) time.Time {
	base := time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location())
	a := t.Sub(base) / units
	return base.Add(a * units)
}

func Seconds(t time.Time) time.Time {
	return Round(t, time.Second)
}

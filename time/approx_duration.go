package gu_time

import (
	"time"
)

type (
	Duration time.Duration
)

func (d Duration) In(units time.Duration) time.Duration {
	a := time.Duration(d) / units
	return time.Duration(a) * units
}

func (d Duration) InAutoUnits(maximum_precision time.Duration) time.Duration {
	duration := time.Duration(d)
	mp := time.Duration(maximum_precision)
	prec := time.Nanosecond
	switch true {
	case duration < time.Microsecond:
		prec = time.Nanosecond
	case duration < time.Millisecond:
		prec = time.Microsecond
	case duration < time.Second:
		prec = time.Millisecond
	default:
		prec = time.Second
	}
	if prec > mp {
		prec = mp
	}
	return d.In(prec)
}

func (d Duration) InSeconds() time.Duration {
	return d.InAutoUnits(time.Second)
}

func SecondsSince(start time.Time) time.Duration {
	return Duration(time.Since(start)).InSeconds()
}

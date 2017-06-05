package gu_time

import (
	"time"
)

type ThrottleType int

const (
	ThrottleInstantValue = ThrottleType(iota)
	ThrottleAverageValue
	ThrottleBoth
)

type ThrottledQuantity interface {
	Name() string
	GetThrottler() Throttler
	GetMaxRate() float64
	GetValue() float64
	GetUnits() float64
	GetTimeUnits() time.Duration
	GetThrottleType() ThrottleType
	Update(increment float64)
	GetLastIncrement() float64
	SetUnits(name, speedName string, valueUnits float64, timeUnits time.Duration)
	GetUnitsName() string
	GetSpeedUnitsName() string
	GetInstantRate() float64
	GetAverageRate() float64
}

type ThrottlingPause interface {
	Qty() ThrottledQuantity
	Amount() time.Duration
	Reason() ThrottleType
	ActualRate() float64
	Wait()
}

type Throttler interface {
	NewThrottler(name string) Throttler
	Name() string
	DefineThrottledQuantity(name string, rate float64, mode ThrottleType) ThrottledQuantity
	GetThrottledQuantity(name string) ThrottledQuantity
	GetThrottledQuantityNames() []string
	StartOperation()
	StopOperation() ThrottlingPause
	IsOperationInProgress() bool
	Freeze()
	IsFrozen() bool
	GetElapsedTime() time.Duration
	SetDutyCycle(rate float64)
	GetDutyCycle() float64
	GetTotalWaitTime() time.Duration
	GetLastLapse() time.Duration
	Pause(time.Duration)
	Throttle()
	Parent() Throttler
	SetUpdateCallback(func(ThrottlingPause))
}

var (
	GlobalThrottler Throttler = globalThrottler()
)

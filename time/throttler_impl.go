package gu_time

import (
	"fmt"
	"sync"
	"time"

	"github.com/maxcalandrelli/goutil/log"
)

var (
	instant_rate_delay_factor float64       = 1 - 0
	minimum_wait              time.Duration = time.Duration(time.Microsecond) * 1
)

type throttledQuantity struct {
	throttler     Throttler
	name          string
	maxRate       float64
	currentValue  float64
	rateUnits     float64
	timeUnits     time.Duration
	mode          ThrottleType
	lastIncrement float64
	unitsName     string
	lastUpdate    time.Time
}

func (tq throttledQuantity) Name() string {
	return tq.name
}

func (tq throttledQuantity) GetThrottler() Throttler {
	return tq.throttler
}

func (tq throttledQuantity) GetMaxRate() float64 {
	return tq.maxRate
}

func (tq throttledQuantity) GetThrottleType() ThrottleType {
	return tq.mode
}

func (tq throttledQuantity) GetUnits() float64 {
	return tq.rateUnits
}

func (tq throttledQuantity) GetTimeUnits() time.Duration {
	return tq.timeUnits
}

func (tq *throttledQuantity) Update(increment float64) {
	tq.lastIncrement = increment
	tq.lastUpdate = time.Now()
}

func (tq throttledQuantity) GetValue() float64 {
	if tq.throttler.IsOperationInProgress() {
		return tq.currentValue + tq.lastIncrement
	}
	return tq.currentValue
}

func (tq throttledQuantity) GetLastIncrement() float64 {
	return tq.lastIncrement
}

func (tq throttledQuantity) GetInstantRate() float64 {
	if tq.throttler.IsOperationInProgress() {
		return tq.lastIncrement * float64(tq.timeUnits) / float64(time.Since(tq.lastUpdate))
	}
	return tq.lastIncrement * float64(tq.timeUnits) / float64(tq.throttler.GetLastLapse())
}

func (tq throttledQuantity) GetAverageRate() float64 {
	if tq.throttler.IsOperationInProgress() {
		return (tq.currentValue + tq.lastIncrement) * float64(tq.timeUnits) / float64(tq.throttler.GetElapsedTime())
	}
	return tq.currentValue * float64(tq.timeUnits) / float64(tq.throttler.GetElapsedTime())
}

func (tq throttledQuantity) GetUnitsName() string {
	return tq.unitsName
}

func (tq *throttledQuantity) SetUnits(name string, rateUnits float64, timeUnits time.Duration) {
	tq.unitsName = name
	tq.rateUnits = rateUnits
	tq.timeUnits = timeUnits
}

type throttledQuantityPause struct {
	qty        ThrottledQuantity
	amount     time.Duration
	reason     ThrottleType
	actualRate float64
}

func (tqp throttledQuantityPause) Qty() ThrottledQuantity {
	return tqp.qty
}

func (tqp throttledQuantityPause) Amount() time.Duration {
	return tqp.amount
}

func (tqp throttledQuantityPause) Reason() ThrottleType {
	return tqp.reason
}

func (tqp throttledQuantityPause) ActualRate() float64 {
	return tqp.actualRate
}

func (tqp *throttledQuantityPause) Wait() {
	if tqp.qty != nil {
		t := tqp.qty.GetThrottler()
		t.Pause(tqp.amount)
		tqp.amount = time.Duration(0)
	}
}

type throttler struct {
	name              string
	first_started     time.Time
	currently_started time.Time
	last_throttling   time.Time
	frozen_at         time.Time
	quantities        map[string]ThrottledQuantity
	lock              sync.Mutex
	parent            Throttler
	totalWait         time.Duration
	lastLapse         time.Duration
	timeQty           ThrottledQuantity
	callback          func(ThrottlingPause)
}

type global_throttler struct {
	throttler
}

func (t *throttler) SetUpdateCallback(callback func(ThrottlingPause)) {
	t.callback = callback
}

func (t *throttler) sendCallback(pause ThrottlingPause) {
	if t.callback != nil {
		t.callback(pause)
	}
	switch t.parent.(type) {
	case *throttler:
		t.parent.(*throttler).sendCallback(pause)
	case *global_throttler:
		t.parent.(*global_throttler).sendCallback(pause)
	}
}

func (t *throttler) SetDutyCycle(rate float64) {
	if rate < 0.0 || rate > 1.0 {
		panic("SetDutyCycle: out of range")
	}
	t.checkCallable("SetDutyCycle", true, false)
	if t.timeQty == nil {
		t.timeQty = t.DefineThrottledQuantity("DutyCycle", rate, ThrottleAverageValue)
		t.timeQty.SetUnits("duty", 1.0, time.Duration(1))
	} else {
		t.timeQty.(*throttledQuantity).maxRate = rate
	}
}

func (t *throttler) GetDutyCycle() float64 {
	if t.timeQty == nil {
		return 1.0
	}
	return t.timeQty.GetValue()
}

func (t *throttler) checkCallable(opname string, mustBeStopped, mustBeStarted bool) {
	if t.IsFrozen() {
		panic(opname + "() called on frozen throttler")
	}
	if mustBeStopped && t.IsOperationInProgress() {
		panic(opname + "() called on started throttler")
	}
	if mustBeStarted && !t.IsOperationInProgress() {
		panic(opname + "() called on stopped throttler")
	}
}

func (t throttler) Name() string {
	return t.name
}

func (t *throttler) DefineThrottledQuantity(name string, rate float64, throttleType ThrottleType) ThrottledQuantity {
	t.checkCallable("DefineThrottledQuantity", true, false)
	if _, ok := t.quantities[name]; ok {
		panic(fmt.Sprintf("DefineThrottledQuantity(%s): already defined", name))
	} else {
		retval := &throttledQuantity{
			throttler:    t,
			name:         name,
			maxRate:      rate,
			currentValue: 0.0,
			mode:         throttleType,
			rateUnits:    1.0,
			timeUnits:    time.Duration(1),
			unitsName:    "",
		}
		t.quantities[name] = retval
		return retval
	}
}

func (t throttler) GetThrottledQuantity(name string) ThrottledQuantity {
	if tq, ok := t.quantities[name]; ok {
		return tq
	}
	return nil
}

func (t throttler) GetThrottledQuantityNames() []string {
	ret := []string{}
	for n, _ := range t.quantities {
		ret = append(ret, n)
	}
	return ret
}

func (t throttler) GetLastLapse() time.Duration {
	return t.lastLapse
}

func updatePause(tq ThrottledQuantity, currentValue float64, interval time.Duration, instantValue bool, pause *throttledQuantityPause) {
	if maximumRate := tq.GetMaxRate(); maximumRate > 0.0 {
		currentRate := currentValue * float64(tq.GetTimeUnits()) / float64(interval)
		instant := map[bool]struct {
			v ThrottleType
			d string
		}{
			false: {v: ThrottleAverageValue, d: ""},
			true:  {v: ThrottleInstantValue, d: "instant "},
		}[instantValue]
		next_minWait := time.Duration(0)
		if currentRate > maximumRate {
			next_minWait = time.Duration(float64(interval) * (currentRate/maximumRate - 1.0))
			if instantValue && tq.GetThrottleType() == ThrottleBoth {
				next_minWait = time.Duration(instant_rate_delay_factor * float64(next_minWait))
			}
			if next_minWait > minimum_wait && next_minWait > pause.amount {
				pause.amount = next_minWait
				pause.qty = tq
				pause.actualRate = currentRate
				pause.reason = instant.v
			}
		}
		dbgl := gu_log.LOG_DEBUG + 8
		if next_minWait == time.Duration(0) {
			dbgl += 1
		}
		gu_log.DebugLog.Custom(
			gu_log.LogLevel(dbgl),
			fmt.Sprintf("%s %svalue=%f units=%f/%v lapse=%v rate=%f/%f %s => wait: %v",
				tq.Name(),
				instant.d,
				currentValue,
				tq.GetUnits(),
				tq.GetTimeUnits(),
				interval,
				currentRate,
				maximumRate,
				tq.GetUnitsName(),
				next_minWait,
			))
	}
}

func (t *throttler) StartOperation() {
	t.checkCallable("StartOperation", true, false)
	t.lock.Lock()
	t.currently_started = time.Now()
	t.last_throttling = t.currently_started
	for _, tq := range t.quantities {
		tq.(*throttledQuantity).lastIncrement = 0.0
	}
}

func (t *throttler) getPause() ThrottlingPause {
	pause := throttledQuantityPause{}
	now := time.Now()
	t.lastLapse = now.Sub(t.currently_started)
	total_lapse := now.Sub(t.first_started)
	if t.timeQty != nil {
		tm := t.timeQty.(*throttledQuantity)
		tm.lastIncrement = float64(now.Sub(t.last_throttling))
		tm.currentValue = float64(now.Sub(t.first_started) - t.totalWait)
	}
	for _, qty := range t.quantities {
		tq := qty.(*throttledQuantity)
		tq.currentValue += tq.lastIncrement
		switch tq.GetThrottleType() {
		case ThrottleAverageValue:
			updatePause(tq, tq.currentValue, total_lapse, false, &pause)
		case ThrottleInstantValue:
			updatePause(tq, tq.lastIncrement, t.lastLapse, true, &pause)
		case ThrottleBoth:
			updatePause(tq, tq.lastIncrement, t.lastLapse, true, &pause)
			updatePause(tq, tq.currentValue, total_lapse, false, &pause)
		}
	}
	t.sendCallback(&pause)
	if pause.Qty() != nil {
		gu_log.DebugLog.Custom(gu_log.LOG_DEBUG+7, fmt.Sprintf("throttling on %s for %v, reason: %srate=%f>%f %s",
			pause.Qty().Name(),
			pause.Amount(),
			map[bool]string{false: "", true: "instant "}[pause.Reason() != ThrottleAverageValue],
			pause.ActualRate(),
			pause.Qty().GetMaxRate(),
			pause.Qty().GetUnitsName(),
		))
	}
	return &pause
}

func (t *throttler) StopOperation() ThrottlingPause {
	t.checkCallable("StopOperation", false, true)
	pause := t.getPause()
	t.currently_started = time.Time{}
	t.last_throttling = time.Time{}
	t.lock.Unlock()
	return pause
}

func (t *throttler) Throttle() {
	t.checkCallable("Throttle", false, true)
	t.getPause().Wait()
	t.last_throttling = time.Now()
}

func (t *throttler) Freeze() {
	t.checkCallable("Freeze", true, false)
	t.frozen_at = time.Now()
}

func (t throttler) IsOperationInProgress() bool {
	return t.currently_started != time.Time{}
}

func (t throttler) IsFrozen() bool {
	return t.frozen_at != time.Time{}
}

func (t throttler) GetElapsedTime() time.Duration {
	if t.IsFrozen() {
		return t.frozen_at.Sub(t.first_started)
	}
	return time.Since(t.first_started)
}

func (t throttler) GetTotalWaitTime() time.Duration {
	return t.totalWait
}

func (t *global_throttler) Pause(duration time.Duration) {
	t.lock.Lock()
	time.Sleep(duration)
	t.totalWait += duration
	t.lock.Unlock()
}

func (t *throttler) Pause(duration time.Duration) {
	t.parent.Pause(duration)
	t.totalWait += duration
}

func (parent *throttler) NewThrottler(name string) Throttler {
	return newThrottler(name, parent)
}

func (parent *global_throttler) NewThrottler(name string) Throttler {
	return newThrottler(name, parent)
}

func (t throttler) Parent() Throttler {
	return t.parent
}

func newThrottler(name string, parent Throttler) *throttler {
	return &throttler{
		name:              name,
		first_started:     time.Now(),
		currently_started: time.Time{},
		last_throttling:   time.Time{},
		frozen_at:         time.Time{},
		quantities:        map[string]ThrottledQuantity{},
		parent:            parent,
	}
}

func globalThrottler() Throttler {
	return &global_throttler{
		throttler: *newThrottler("GLOBALTHROTTLER", nil),
	}
}

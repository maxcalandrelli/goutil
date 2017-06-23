package gu_io

import (
	"errors"
	"io"
	"time"

	"github.com/maxcalandrelli/goutil/time"
)

type throttledReader struct {
	byteRateThrottler gu_time.ThrottledQuantity
	underlying_reader io.Reader
}

func (th *throttledReader) Read(p []byte) (n int, err error) {
	throttler := th.byteRateThrottler.GetThrottler()
	wasStarted := throttler.IsOperationInProgress()
	if !wasStarted {
		throttler.StartOperation()
	}
	if n, err = th.underlying_reader.Read(p); n > 0 {
		th.byteRateThrottler.Update(float64(n) / th.byteRateThrottler.GetUnits())
		throttler.Throttle()
	}
	if !wasStarted {
		pause := throttler.StopOperation()
		if pause.Amount() > time.Duration(0) {
			pause.Wait()
		}
	}
	return
}

func (th *throttledReader) ReadAt(p []byte, offs int64) (n int, err error) {
	if ra, ok := th.underlying_reader.(io.ReaderAt); ok {
		if _, err := ra.ReadAt([]byte{}, offs); err != nil {
			return 0, err
		}
		return th.underlying_reader.Read(p)
	} else {
		return 0, errors.New("ReadAt not implemented")
	}
}

func NewThrottledReader(rdr io.Reader, limit gu_time.ThrottledQuantity) io.Reader {
	return &throttledReader{
		byteRateThrottler: limit,
		underlying_reader: rdr,
	}
}

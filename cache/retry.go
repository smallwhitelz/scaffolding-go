package cache

import "time"

type RetryStrategy interface {
	// Next
	// time.Duration 重试的间隔
	// bool 要不要继续重试
	Next() (time.Duration, bool)
}

type FixedIntervalRetryStrategy struct {
	Interval time.Duration
	MaxCnt   int
	Cnt      int
}

func (f *FixedIntervalRetryStrategy) Next() (time.Duration, bool) {
	if f.Cnt >= f.MaxCnt {
		return 0, false
	}
	return f.Interval, true
}

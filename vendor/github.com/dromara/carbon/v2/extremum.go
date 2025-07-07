package carbon

import (
	"time"
)

const (
	minDuration Duration = -1 << 63
	maxDuration Duration = 1<<63 - 1
)

// ZeroValue returns the zero value of Carbon instance.
func ZeroValue() *Carbon {
	return MinValue()
}

// EpochValue returns the unix epoch value of Carbon instance.
func EpochValue() *Carbon {
	return NewCarbon(time.Date(EpochYear, time.January, MinDay, MinHour, MinMinute, MinSecond, MinNanosecond, time.UTC))
}

// MaxValue returns the maximum value of Carbon instance.
func MaxValue() *Carbon {
	return NewCarbon(time.Date(MaxYear, time.December, MaxDay, MaxHour, MaxMinute, MaxSecond, MaxNanosecond, time.UTC))
}

// MinValue returns the minimum value of Carbon instance.
func MinValue() *Carbon {
	return NewCarbon(time.Date(MinYear, time.January, MinDay, MinHour, MinMinute, MinSecond, MinNanosecond, time.UTC))
}

// MaxDuration returns the maximum value of duration instance.
func MaxDuration() Duration {
	return maxDuration
}

// MinDuration returns the minimum value of duration instance.
func MinDuration() Duration {
	return minDuration
}

// Max returns the maximum Carbon instance from some given Carbon instances.
func Max(c1 *Carbon, c2 ...*Carbon) (c *Carbon) {
	c = c1
	if c.IsInvalid() {
		return
	}
	if len(c2) == 0 {
		return
	}
	for _, carbon := range c2 {
		if carbon.IsInvalid() {
			return carbon
		}
		if carbon.Gte(c) {
			c = carbon
		}
	}
	return
}

// Min returns the minimum Carbon instance from some given Carbon instances.
func Min(c1 *Carbon, c2 ...*Carbon) (c *Carbon) {
	c = c1
	if c.IsInvalid() {
		return
	}
	if len(c2) == 0 {
		return
	}
	for _, carbon := range c2 {
		if carbon.IsInvalid() {
			return carbon
		}
		if carbon.Lte(c) {
			c = carbon
		}
	}
	return
}

// Closest returns the closest Carbon instance from some given Carbon instances.
func (c *Carbon) Closest(c1 *Carbon, c2 ...*Carbon) *Carbon {
	if c.IsInvalid() {
		return c
	}
	if c1.IsInvalid() {
		return c1
	}
	if len(c2) == 0 {
		return c1
	}
	closest := c1
	minDiff := c.DiffAbsInSeconds(closest)
	for _, arg := range c2 {
		if arg.IsInvalid() {
			return arg
		}
		diff := c.DiffAbsInSeconds(arg)
		if diff < minDiff {
			minDiff = diff
			closest = arg
		}
	}
	return closest
}

// Farthest returns the farthest Carbon instance from some given Carbon instances.
func (c *Carbon) Farthest(c1 *Carbon, c2 ...*Carbon) *Carbon {
	if c.IsInvalid() {
		return c
	}
	if c1.IsInvalid() {
		return c1
	}
	if len(c2) == 0 {
		return c1
	}
	farthest := c1
	maxDiff := c.DiffAbsInSeconds(farthest)
	for _, arg := range c2 {
		if arg.IsInvalid() {
			return arg
		}
		diff := c.DiffAbsInSeconds(arg)
		if diff > maxDiff {
			maxDiff = diff
			farthest = arg
		}
	}
	return farthest
}

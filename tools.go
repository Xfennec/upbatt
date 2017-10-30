package main

import (
	"time"

	"github.com/hako/durafmt"
)

// DurationRound will round d to r "precision"
func DurationRound(d, r time.Duration) time.Duration {
	if r <= 0 {
		return d
	}
	neg := d < 0
	if neg {
		d = -d
	}
	if m := d % r; m+m < r {
		d = d - m
	} else {
		d = d + r - m
	}
	if neg {
		return -d
	}
	return d
}

// SinceFmt returns a string formatted duration since Now
func SinceFmt(t time.Time) string {
	since := time.Now().Sub(t)
	round := time.Minute
	if since < 2*time.Minute {
		round = time.Second
	}
	durfmt := durafmt.Parse(DurationRound(since, round))
	return durfmt.String()
}

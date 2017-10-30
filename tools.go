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

// DurationFmt return a string-formatted duration
func DurationFmt(d time.Duration) string {
	round := time.Minute
	if d < 2*time.Minute {
		round = time.Second
	}
	durfmt := durafmt.Parse(DurationRound(d, round))
	return durfmt.String()
}

// SinceFmt returns a string-formatted duration since now
func SinceFmt(t time.Time) string {
	since := time.Now().Sub(t)
	return DurationFmt(since)
}

// Decline word to its plural (adding a "s") if count != 1
func Decline(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

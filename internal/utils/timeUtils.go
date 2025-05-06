package utils

import (
	"fmt"
	"time"
	"yadro-biathlon/internal/config"
)

// FormatTimeString wraps a time.Time value in square brackets
// using the configured time format for event logging.
func FormatTimeString(t time.Time) string {
	return "[" + t.Format(config.TimeFormat) + "]"
}

// FormatDurationString formats a time.Duration into a human-readable
// string in the form HH:MM:SS.milliseconds, zero-padded as needed.
func FormatDurationString(d time.Duration) string {
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	d -= s * time.Second
	ms := d / time.Millisecond
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}

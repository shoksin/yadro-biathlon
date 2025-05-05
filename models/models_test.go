package models

import (
	"testing"
	"time"
	"yadro-test/config"
)

func TestFormatTimeString(t *testing.T) {
	tests := []struct {
		input    time.Time
		expected string
	}{
		{time.Date(0, 0, 0, 9, 30, 5, 0, time.UTC), "[09:30:05.000]"},
		{time.Date(0, 1, 1, 10, 15, 30, 500_000_000, time.UTC), "[10:15:30.500]"},
	}

	for _, test := range tests {
		t.Run(test.input.Format(config.TimeFormat), func(t *testing.T) {
			formatted := FormatTimeString(test.input)
			if formatted != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, formatted)
			}
		})
	}
}

func TestFormatDurationString(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{time.Hour + 5*time.Minute + 30*time.Second, "01:05:30.000"},
		{29*time.Minute + 3*time.Second + 872*time.Millisecond, "00:29:03.872"},
		{0, "00:00:00.000"},
	}

	for _, test := range tests {
		t.Run(test.input.String(), func(t *testing.T) {
			formatted := FormatDurationString(test.input)
			if formatted != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, formatted)
			}
		})
	}
}

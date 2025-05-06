package events

import (
	"testing"
	"yadro-test/internal/config"
	"yadro-test/internal/models"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[09:31:49.285]", "09:31:49.285"},
		{"[09:58:00.000]", "09:58:00.000"},
		{"[09:59:30.123]", "09:59:30.123"},
		{"[10:32:20.999]", "10:32:20.999"},
		{"[23:10:10.111]", "23:10:10.111"},
		{"[00:00:01.000]", "00:00:01.000"},
		{"[00:00:00.001]", "00:00:00.001"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			parsedTime, err := ParseTime(test.input)
			if err != nil {
				t.Errorf("ParseTime(%s) failed: %v", test.input, err)
			}

			formattedTime := parsedTime.Format(config.TimeFormat)
			if formattedTime != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, formattedTime)
			}

		})
	}
}

func TestParseEvent(t *testing.T) {
	tests := []struct {
		input                string
		expectedID           int
		expectedCompetitorID int
		expectedExtraParams  string
	}{
		{"[09:38:28.673] 1 1", 1, 1, ""},
		{"[09:58:00.000] 2 3 10:03:00.000", 2, 3, "10:03:00.000"},
		{"[10:10:22.273] 5 2 1", 5, 2, "1"},
		{"[10:26:38.368] 6 4 1", 6, 4, "1"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			parsedEvent, err := ParseEvent(test.input)
			if err != nil {
				t.Errorf("ParseEvent(%s) failed: %v", test.input, err)
			}

			if parsedEvent.Action != models.Action(test.expectedID) {
				t.Errorf("Expected ID=%d, got %d", test.expectedID, parsedEvent.Action)
			}
			if parsedEvent.CompetitorID != test.expectedCompetitorID {
				t.Errorf("Expected EventID=%d, got %d", test.expectedCompetitorID, parsedEvent.CompetitorID)
			}
			if parsedEvent.ExtraParams != test.expectedExtraParams {
				t.Errorf("Expected ExtraParams=%s, got %s", test.expectedExtraParams, parsedEvent.ExtraParams)
			}
		})
	}
}

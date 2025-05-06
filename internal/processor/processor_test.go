package processor

import (
	"strings"
	"testing"
	"time"
	"yadro-test/internal/config"
	"yadro-test/internal/models"
)

func TestCheckDisqualifications(t *testing.T) {
	conf := config.Configuration{
		StartDelta: "00:00:30",
	}

	processor := NewEventProcessor(conf)

	startTime := time.Date(0, 1, 1, 9, 30, 0, 0, time.UTC)
	processor.Competitors[1] = &models.Competitor{
		ID:           1,
		PlannedStart: startTime,
		Status:       models.Registered,
	}

	processor.CheckDisqualifications()

	comp := processor.Competitors[1]
	if comp.Status != models.NotStarted {
		t.Errorf("Expected status NotStarted, got %v", comp.Status)
	}
}

func TestGenerateReport(t *testing.T) {
	conf := config.Configuration{
		Laps:       2,
		LapLen:     3651,
		PenaltyLen: 50,
	}

	processor := NewEventProcessor(conf)

	processor.Competitors[1] = &models.Competitor{
		ID:            1,
		Status:        models.NotFinished,
		LapsResult:    []models.LapResult{{Time: 29*time.Minute + 3*time.Second + 872*time.Millisecond, Speed: 2.093}},
		PenaltyResult: models.PenaltyResult{Time: 1*time.Minute + 44*time.Second + 296*time.Millisecond, Speed: 0.481},
		Hits:          4,
		Shots:         5,
	}

	report := processor.GenerateReport()

	expectedParts := []string{
		"[NotFinished] 1",
		"{00:29:03.872, 2.093}",
		"{00:01:44.296, 0.481}",
		"4/5",
	}

	for _, part := range expectedParts {
		if !strings.Contains(report, part) {
			t.Errorf("Expected report to contain '%s', but it didn't. Report: %s", part, report)
		}
	}
}

package processor

import (
	"bytes"
	"io"
	"math"
	"os"
	"strings"
	"testing"
	"time"
	"yadro-biathlon/internal/config"
	"yadro-biathlon/internal/models"
	"yadro-biathlon/internal/utils"
)

func TestCheckDisqualifications(t *testing.T) {
	processor := createTestProcessor()

	if processor.Config.StartDelta != "00:00:30" {
		t.Errorf("Expected StartDelta to be '00:00:30', got '%s'", processor.Config.StartDelta)
	}

	startTime := time.Date(0, 1, 1, 9, 30, 0, 0, time.UTC)
	processor.Competitors[1] = &models.Competitor{
		ID:           1,
		PlannedStart: startTime,
		Status:       models.Registered,
	}

	output := captureOutput(func() {
		processor.CheckDisqualifications()
	})

	comp := processor.Competitors[1]
	if comp.Status != models.NotStarted {
		t.Errorf("Expected status NotStarted, got %v", comp.Status)
	}

	if !strings.Contains(output, "The competitor(1) is disqualified") {
		t.Errorf("Expected disqualification message, got: %s", output)
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

// captureOutput captures stdout for testing logged output
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// createTestProcessor creates a processor with standard test configuration
func createTestProcessor() *EventProcessor {
	return NewEventProcessor(config.Configuration{
		Laps:        2,
		LapLen:      3651,
		PenaltyLen:  50,
		FiringLines: 1,
		Start:       "09:30:00.000",
		StartDelta:  "00:00:30",
	})
}

// createTestEvent creates a standard test event
func createTestEvent(action models.Action, competitorID int, timeStr string, extraParams string) models.Event {
	t, _ := time.Parse(config.TimeFormat, timeStr)
	return models.Event{
		Time:         t,
		TimeString:   utils.FormatTimeString(t),
		Action:       action,
		CompetitorID: competitorID,
		ExtraParams:  extraParams,
	}
}

func TestHandleRegistered(t *testing.T) {
	processor := createTestProcessor()
	comp := &models.Competitor{ID: 1}
	event := createTestEvent(models.ActionRegistered, 1, "09:05:59.867", "")

	output := captureOutput(func() {
		processor.handleRegistered(event, comp)
	})

	if comp.Status != models.Registered {
		t.Errorf("Expected status Registered, got %v", comp.Status)
	}

	// Verify logging
	expectedLog := "The competitor(1) registered"
	if !strings.Contains(output, expectedLog) {
		t.Errorf("Expected log to contain '%s', got: %s", expectedLog, output)
	}
}

func TestHandleStartTimeSet(t *testing.T) {
	processor := createTestProcessor()
	comp := &models.Competitor{ID: 1}
	event := createTestEvent(models.ActionStartTimeSet, 1, "09:15:00.841", "09:30:00.000")

	output := captureOutput(func() {
		processor.handleStartTimeSet(event, comp)
	})

	expectedTime, _ := time.Parse(config.TimeFormat, "09:30:00.000")
	if !comp.PlannedStart.Equal(expectedTime) {
		t.Errorf("Expected planned start %v, got %v", expectedTime, comp.PlannedStart)
	}
	if comp.Status != models.Registered {
		t.Errorf("Expected status Registered, got %v", comp.Status)
	}

	expectedLog := "The start time for the competitor(1) was set by a draw to 09:30:00.000"
	if !strings.Contains(output, expectedLog) {
		t.Errorf("Expected log to contain '%s', got: %s", expectedLog, output)
	}
}

func TestHandleOnStartLine(t *testing.T) {
	processor := createTestProcessor()
	comp := &models.Competitor{ID: 1}
	event := createTestEvent(models.ActionOnStartLine, 1, "09:29:45.734", "")

	output := captureOutput(func() {
		processor.handleOnStartLine(event, comp)
	})

	if comp.Status != models.OnStartLine {
		t.Errorf("Expected status OnStartLine, got %v", comp.Status)
	}

	expectedLog := "The competitor(1) is on the start line"
	if !strings.Contains(output, expectedLog) {
		t.Errorf("Expected log to contain '%s', got: %s", expectedLog, output)
	}
}

func TestHandleStarted(t *testing.T) {
	processor := createTestProcessor()
	comp := &models.Competitor{ID: 1}
	startTime := "09:30:01.005"
	parsedTime, _ := time.Parse(config.TimeFormat, startTime)
	event := createTestEvent(models.ActionStarted, 1, startTime, "")

	output := captureOutput(func() {
		processor.handleStarted(event, comp)
	})

	if comp.Status != models.Started {
		t.Errorf("Expected status Started, got %v", comp.Status)
	}
	if !comp.ActualStart.Equal(parsedTime) {
		t.Errorf("Expected actual start %v, got %v", parsedTime, comp.ActualStart)
	}
	if comp.CurrentLap != 1 {
		t.Errorf("Expected current lap 1, got %d", comp.CurrentLap)
	}

	expectedLog := "The competitor(1) has started"
	if !strings.Contains(output, expectedLog) {
		t.Errorf("Expected log to contain '%s', got: %s", expectedLog, output)
	}
}

func TestHandleOnFiringRange(t *testing.T) {
	processor := createTestProcessor()
	comp := &models.Competitor{ID: 1}
	event := createTestEvent(models.ActionOnFiringRange, 1, "09:49:31.659", "1")

	output := captureOutput(func() {
		processor.handleOnFiringRange(event, comp)
	})

	if comp.Status != models.OnFiringRange {
		t.Errorf("Expected status OnFiringRange, got %v", comp.Status)
	}

	expectedLog := "The competitor(1) is on the firing range(1)"
	if !strings.Contains(output, expectedLog) {
		t.Errorf("Expected log to contain '%s', got: %s", expectedLog, output)
	}
}

func TestHandleHit(t *testing.T) {
	processor := createTestProcessor()
	comp := &models.Competitor{ID: 1, Hits: 0, LastFiringHits: 0}
	event := createTestEvent(models.ActionHit, 1, "09:49:33.123", "1")

	output := captureOutput(func() {
		processor.handleHit(event, comp)
	})

	// Verify state changes
	if comp.Hits != 1 {
		t.Errorf("Expected hits 1, got %d", comp.Hits)
	}
	if comp.LastFiringHits != 1 {
		t.Errorf("Expected last firing hits 1, got %d", comp.LastFiringHits)
	}

	expectedLog := "The target(1) has been hit by competitor(1)"
	if !strings.Contains(output, expectedLog) {
		t.Errorf("Expected log to contain '%s', got: %s", expectedLog, output)
	}
}

func TestHandleLeftFiringRange(t *testing.T) {
	processor := createTestProcessor()
	comp := &models.Competitor{ID: 1, Shots: 0}
	event := createTestEvent(models.ActionLeftFiringRange, 1, "09:49:38.339", "")

	output := captureOutput(func() {
		processor.handleLeftFiringRange(event, comp)
	})

	if comp.Status != models.LeftFiringRange {
		t.Errorf("Expected status LeftFiringRange, got %v", comp.Status)
	}
	if comp.Shots != 5 {
		t.Errorf("Expected shots 5, got %d", comp.Shots)
	}

	expectedLog := "The competitor(1) left the firing range"
	if !strings.Contains(output, expectedLog) {
		t.Errorf("Expected log to contain '%s', got: %s", expectedLog, output)
	}
}

func TestHandleOnPenaltyLaps(t *testing.T) {
	processor := createTestProcessor()
	comp := &models.Competitor{ID: 1}
	eventTime := "09:49:55.915"
	parsedTime, _ := time.Parse(config.TimeFormat, eventTime)
	event := createTestEvent(models.ActionOnPenaltyLaps, 1, eventTime, "")

	output := captureOutput(func() {
		processor.handleOnPenaltyLaps(event, comp)
	})

	if comp.Status != models.OnPenaltyLaps {
		t.Errorf("Expected status OnPenaltyLaps, got %v", comp.Status)
	}
	if !comp.PenaltyStartTime.Equal(parsedTime) {
		t.Errorf("Expected penalty start time %v, got %v", parsedTime, comp.PenaltyStartTime)
	}

	expectedLog := "The competitor(1) entered the penalty laps"
	if !strings.Contains(output, expectedLog) {
		t.Errorf("Expected log to contain '%s', got: %s", expectedLog, output)
	}
}

func TestHandleLeftPenaltyLaps(t *testing.T) {
	processor := createTestProcessor()

	startTime, _ := time.Parse(config.TimeFormat, "09:49:55.915")
	comp := &models.Competitor{
		ID:               1,
		Hits:             4,
		Shots:            5,
		PenaltyStartTime: startTime,
		Status:           models.OnPenaltyLaps,
	}

	eventTime := "09:51:48.391"
	endTime, _ := time.Parse(config.TimeFormat, eventTime)
	event := createTestEvent(models.ActionLeftPenaltyLaps, 1, eventTime, "")

	output := captureOutput(func() {
		processor.handleLeftPenaltyLaps(event, comp)
	})

	if comp.Status != models.LeftPenaltyLaps {
		t.Errorf("Expected status LeftPenaltyLaps, got %v", comp.Status)
	}

	expectedPenaltyTime := endTime.Sub(startTime)
	if comp.FullPenaltyTime != expectedPenaltyTime {
		t.Errorf("Expected full penalty time %v, got %v", expectedPenaltyTime, comp.FullPenaltyTime)
	}

	// With 1 miss, we expect 1 * 50 = 50m penalty distance
	// Speed should be 50m / penalty time
	expectedSpeed := float64(50) / expectedPenaltyTime.Seconds()
	if comp.PenaltyResult.Speed != expectedSpeed {
		t.Errorf("Expected penalty speed %.3f, got %.3f", expectedSpeed, comp.PenaltyResult.Speed)
	}

	expectedLog := "The competitor(1) left the penalty laps"
	if !strings.Contains(output, expectedLog) {
		t.Errorf("Expected log to contain '%s', got: %s", expectedLog, output)
	}
}

func TestHandleFinishedLap(t *testing.T) {
	processor := createTestProcessor()

	startTime, _ := time.Parse(config.TimeFormat, "09:30:01.005")
	lapStartTime, _ := time.Parse(config.TimeFormat, "09:30:01.005")
	comp := &models.Competitor{
		ID:             1,
		Status:         models.LeftPenaltyLaps,
		CurrentLap:     1,
		LapStartTime:   lapStartTime,
		PlannedStart:   startTime,
		LastFiringHits: 4, // 1 miss
		LapsResult:     make([]models.LapResult, 0, 2),
	}

	eventTime := "09:59:03.872"
	endTime, _ := time.Parse(config.TimeFormat, eventTime)
	event := createTestEvent(models.ActionFinishedLap, 1, eventTime, "")

	lapTime := endTime.Sub(lapStartTime)
	penaltyDistance := (5 - comp.LastFiringHits) * processor.Config.PenaltyLen // 1 miss = 50m
	totalDistance := float64(processor.Config.LapLen) + float64(penaltyDistance)
	expectedSpeed := totalDistance / lapTime.Seconds()

	output := captureOutput(func() {
		processor.handleFinishedLap(event, comp)
	})

	if comp.CurrentLap != 2 {
		t.Errorf("Expected current lap 2, got %d", comp.CurrentLap)
	}

	if comp.Status != models.FinishedLap {
		t.Errorf("Expected status FinishedLap, got %v", comp.Status)
	}

	if len(comp.LapsResult) != 1 {
		t.Errorf("Expected 1 lap result, got %d", len(comp.LapsResult))
	} else {
		if comp.LapsResult[0].Time != lapTime {
			t.Errorf("Expected lap time %v, got %v", lapTime, comp.LapsResult[0].Time)
		}

		if math.Abs(comp.LapsResult[0].Speed-expectedSpeed) > 0.001 {
			t.Errorf("Expected lap speed %.3f, got %.3f", expectedSpeed, comp.LapsResult[0].Speed)
		}
	}

	expectedLog := "The competitor(1) ended the main lap"
	if !strings.Contains(output, expectedLog) {
		t.Errorf("Expected log to contain '%s', got: %s", expectedLog, output)
	}
}

func TestHandleFinishedLap_Final(t *testing.T) {
	processor := createTestProcessor()

	startTime, _ := time.Parse(config.TimeFormat, "09:30:01.005")
	lapStartTime, _ := time.Parse(config.TimeFormat, "10:00:00.000")
	comp := &models.Competitor{
		ID:             1,
		Status:         models.LeftPenaltyLaps,
		CurrentLap:     processor.Config.Laps, // Last lap
		LapStartTime:   lapStartTime,
		PlannedStart:   startTime,
		LastFiringHits: 5,
		LapsResult:     make([]models.LapResult, 0, 2),
	}

	eventTime := "10:25:26.047"
	endTime, _ := time.Parse(config.TimeFormat, eventTime)
	event := createTestEvent(models.ActionFinishedLap, 1, eventTime, "")

	output := captureOutput(func() {
		processor.handleFinishedLap(event, comp)
	})

	if comp.Status != models.Finished {
		t.Errorf("Expected status Finished, got %v", comp.Status)
	}

	expectedTotalTime := endTime.Sub(startTime)
	if comp.TotalTime != expectedTotalTime {
		t.Errorf("Expected total time %v, got %v", expectedTotalTime, comp.TotalTime)
	}

	expectedLog := "The competitor(1) has finished"
	if !strings.Contains(output, expectedLog) {
		t.Errorf("Expected log to contain '%s', got: %s", expectedLog, output)
	}
}

func TestHandleCannotContinue(t *testing.T) {
	processor := createTestProcessor()
	comp := &models.Competitor{ID: 1}
	event := createTestEvent(models.ActionCannotContinue, 1, "09:59:05.321", "Lost in the forest")

	output := captureOutput(func() {
		processor.handleCannotContinue(event, comp)
	})

	if comp.Status != models.NotFinished {
		t.Errorf("Expected status NotFinished, got %v", comp.Status)
	}

	expectedLog := "The competitor(1) can`t continue: Lost in the forest"
	if !strings.Contains(output, expectedLog) {
		t.Errorf("Expected log to contain '%s', got: %s", expectedLog, output)
	}
}

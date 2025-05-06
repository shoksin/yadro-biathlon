package processor

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"yadro-biathlon/internal/config"
	"yadro-biathlon/internal/messages"
	"yadro-biathlon/internal/models"
	"yadro-biathlon/internal/utils"
)

// EventProcessor manages the lifecycle of competitor events in a biathlon race.
// It collects events, updates competitor state, logs progress, and generates the final report.
type EventProcessor struct {
	Config      config.Configuration
	Competitors map[int]*models.Competitor
	Events      []models.Event
	logFile     *os.File
	logWriter   *bufio.Writer
	mu          sync.Mutex
}

// NewEventProcessor creates an EventProcessor with the given configuration.
// Initializes internal maps and event slice.
func NewEventProcessor(config config.Configuration) *EventProcessor {
	return &EventProcessor{
		Config:      config,
		Competitors: make(map[int]*models.Competitor),
		Events:      []models.Event{},
	}
}

// WriteLog outputs a log line to stdout and, if enabled, to the log file.
func (ep *EventProcessor) WriteLog(logText string) {
	fmt.Println(logText)

	if ep.logWriter != nil {
		_, err := ep.logWriter.WriteString(logText + "\n")
		if err != nil {
			fmt.Printf("Warning: error writing to log file: %v\n", err)
		}
		ep.logWriter.Flush()
	}
}

// Close flushes and closes the log file if it was enabled.
func (ep *EventProcessor) Close() error {
	if ep.logFile == nil {
		return nil
	}

	if ep.logWriter != nil {
		err := ep.logWriter.Flush()
		if err != nil {
			return err
		}
	}

	return ep.logFile.Close()
}

// EnableLogFile opens the specified filename for logging and prepares a buffered writer.
func (ep *EventProcessor) EnableLogFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	ep.logFile = file
	ep.logWriter = bufio.NewWriter(file)
	return nil
}

// ProcessEvent routes a single event to its handler based on event.Action.
// Updates competitor state and appends the event to history.
func (ep *EventProcessor) ProcessEvent(event models.Event) {
	comp, exists := ep.Competitors[event.CompetitorID]
	if !exists {
		comp = &models.Competitor{
			ID:         event.CompetitorID,
			Status:     models.Registered,
			LapsResult: make([]models.LapResult, 0, ep.Config.Laps),
		}
		ep.Competitors[event.CompetitorID] = comp
	}

	switch event.Action {
	case models.ActionRegistered:
		ep.handleRegistered(event, comp)
	case models.ActionStartTimeSet:
		ep.handleStartTimeSet(event, comp)
	case models.ActionOnStartLine:
		ep.handleOnStartLine(event, comp)
	case models.ActionStarted:
		ep.handleStarted(event, comp)
	case models.ActionOnFiringRange:
		ep.handleOnFiringRange(event, comp)
	case models.ActionHit:
		ep.handleHit(event, comp)
	case models.ActionLeftFiringRange:
		ep.handleLeftFiringRange(event, comp)
	case models.ActionOnPenaltyLaps:
		ep.handleOnPenaltyLaps(event, comp)
	case models.ActionLeftPenaltyLaps:
		ep.handleLeftPenaltyLaps(event, comp)
	case models.ActionFinishedLap:
		ep.handleFinishedLap(event, comp)
	case models.ActionCannotContinue:
		ep.handleCannotContinue(event, comp)
	}

	ep.Events = append(ep.Events, event)
}

// Individual handlers update competitor state and log each specific action.
// These methods keep ProcessEvent concise and improve readability.

func (ep *EventProcessor) handleRegistered(event models.Event, comp *models.Competitor) {
	comp.Status = models.Registered
	ep.WriteLog(fmt.Sprintf(messages.Registered, event.TimeString, comp.ID))
}

func (ep *EventProcessor) handleStartTimeSet(event models.Event, comp *models.Competitor) {
	startTime, err := time.Parse(config.TimeFormat, event.ExtraParams)
	if err != nil {
		fmt.Printf("Error parsing start time: %v\n", err)
		return
	}
	comp.PlannedStart = startTime
	comp.LapStartTime = startTime
	comp.Status = models.Registered
	ep.WriteLog(fmt.Sprintf(messages.StartTimeSet, event.TimeString, comp.ID, event.ExtraParams))
}

func (ep *EventProcessor) handleOnStartLine(event models.Event, comp *models.Competitor) {
	comp.Status = models.OnStartLine
	ep.WriteLog(fmt.Sprintf(messages.OnStartLine, event.TimeString, comp.ID))
}

func (ep *EventProcessor) handleStarted(event models.Event, comp *models.Competitor) {
	comp.ActualStart = event.Time
	comp.CurrentLap = 1
	comp.Status = models.Started
	ep.WriteLog(fmt.Sprintf(messages.Started, event.TimeString, comp.ID))
}

func (ep *EventProcessor) handleOnFiringRange(event models.Event, comp *models.Competitor) {
	comp.Status = models.OnFiringRange
	ep.WriteLog(fmt.Sprintf(messages.OnFiringRange, event.TimeString, comp.ID, event.ExtraParams))
}

func (ep *EventProcessor) handleHit(event models.Event, comp *models.Competitor) {
	comp.Hits++
	comp.LastFiringHits++
	ep.WriteLog(fmt.Sprintf(messages.TargetHit, event.TimeString, event.ExtraParams, comp.ID))
}

func (ep *EventProcessor) handleLeftFiringRange(event models.Event, comp *models.Competitor) {
	comp.Shots += 5
	comp.Status = models.LeftFiringRange
	ep.WriteLog(fmt.Sprintf(messages.LeftFiringRange, event.TimeString, comp.ID))
}

func (ep *EventProcessor) handleOnPenaltyLaps(event models.Event, comp *models.Competitor) {
	comp.PenaltyStartTime = event.Time
	comp.Status = models.OnPenaltyLaps
	ep.WriteLog(fmt.Sprintf(messages.EnteredPenaltyLaps, event.TimeString, comp.ID))
}

func (ep *EventProcessor) handleLeftPenaltyLaps(event models.Event, comp *models.Competitor) {
	penaltyTime := event.Time.Sub(comp.PenaltyStartTime)
	allMisses := comp.Shots - comp.Hits
	penaltyDistance := allMisses * ep.Config.PenaltyLen

	comp.FullPenaltyTime += penaltyTime

	var speed float64
	if comp.FullPenaltyTime.Seconds() > 0 {
		speed = float64(penaltyDistance) / comp.FullPenaltyTime.Seconds()
	}
	comp.PenaltyResult = models.PenaltyResult{Time: comp.FullPenaltyTime, Speed: speed}
	comp.Status = models.LeftPenaltyLaps
	ep.WriteLog(fmt.Sprintf(messages.LeftPenaltyLaps, event.TimeString, comp.ID))
}

func (ep *EventProcessor) handleFinishedLap(event models.Event, comp *models.Competitor) {
	lapTime := event.Time.Sub(comp.LapStartTime)
	lastPenaltyDistance := (5 - comp.LastFiringHits) * ep.Config.PenaltyLen
	comp.LastFiringHits = 0

	speed := (float64(ep.Config.LapLen) + float64(lastPenaltyDistance)) / lapTime.Seconds()
	comp.LapsResult = append(comp.LapsResult, models.LapResult{Time: lapTime, Speed: speed})
	comp.Status = models.FinishedLap
	ep.WriteLog(fmt.Sprintf(messages.MainLapEnded, event.TimeString, comp.ID))

	if comp.CurrentLap >= ep.Config.Laps {
		comp.Status = models.Finished
		comp.TotalTime = event.Time.Sub(comp.PlannedStart)
		ep.WriteLog(fmt.Sprintf(messages.Finished, event.TimeString, comp.ID))
	} else {
		comp.CurrentLap++
		comp.LapStartTime = event.Time
	}
}

func (ep *EventProcessor) handleCannotContinue(event models.Event, comp *models.Competitor) {
	comp.Status = models.NotFinished
	ep.WriteLog(fmt.Sprintf(messages.CannotContinue, event.TimeString, comp.ID, event.ExtraParams))
}

// CheckDisqualifications inspects actual start times against allowed delta and disqualifies late/no-shows.
func (ep *EventProcessor) CheckDisqualifications() {
	deltaTimeStr := ep.Config.StartDelta

	deltaParts := strings.Split(deltaTimeStr, ":")
	if len(deltaParts) != 3 {
		fmt.Printf("Not correct delta time: %s\n", deltaTimeStr)
		return
	}

	h, _ := strconv.Atoi(deltaParts[0])
	m, _ := strconv.Atoi(deltaParts[1])
	s, _ := strconv.Atoi(deltaParts[2])

	startDeltaDuration := time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second

	for _, comp := range ep.Competitors {
		endStartInterval := comp.PlannedStart.Add(startDeltaDuration)
		if comp.ActualStart.IsZero() || comp.ActualStart.After(endStartInterval) || comp.ActualStart.Before(comp.PlannedStart) {
			comp.Status = models.NotStarted
			disqualificationTimeStr := utils.FormatTimeString(endStartInterval.Add(time.Millisecond))
			ep.WriteLog(fmt.Sprintf(messages.Disqualified, disqualificationTimeStr, comp.ID))
		}

	}
}

func (ep *EventProcessor) ProcessEvents(events []models.Event) {
	for _, event := range events {
		ep.ProcessEvent(event)
	}
}

// GenerateReport sorts competitors, includes lap and penalty results, and returns formatted report.
func (ep *EventProcessor) GenerateReport() string {
	ep.CheckDisqualifications()

	var sortedCompetitors []*models.Competitor
	for _, comp := range ep.Competitors {
		sortedCompetitors = append(sortedCompetitors, comp)
	}

	sort.Slice(sortedCompetitors, func(i, j int) bool {
		a, b := sortedCompetitors[i], sortedCompetitors[j]

		if a.Status == models.NotStarted && b.Status == models.NotStarted {
			return a.ID < b.ID
		}
		if a.Status == models.NotStarted {
			return false
		}
		if b.Status == models.NotStarted {
			return true
		}

		if a.Status == models.NotFinished && b.Status == models.NotFinished {
			return a.ID < b.ID
		}
		if a.Status == models.NotFinished {
			return false
		}
		if b.Status == models.NotFinished {
			return true
		}

		return a.TotalTime < b.TotalTime
	})

	var report strings.Builder
	for _, comp := range sortedCompetitors {
		switch comp.Status {
		case models.NotStarted:
			report.WriteString(fmt.Sprintf("[NotStarted] %d", comp.ID))
		case models.NotFinished:
			report.WriteString(fmt.Sprintf("[NotFinished] %d", comp.ID))
		default:
			report.WriteString(fmt.Sprintf("[%s] %d", utils.FormatDurationString(comp.TotalTime), comp.ID))
		}

		report.WriteString(" [")
		for i := 0; i < ep.Config.Laps; i++ {
			if i > 0 {
				report.WriteString(", ")
			}
			if i < len(comp.LapsResult) {
				lapResult := comp.LapsResult[i]
				report.WriteString(fmt.Sprintf("{%s, %.3f}", utils.FormatDurationString(lapResult.Time), lapResult.Speed))
			} else {
				report.WriteString("{,}")
			}
		}
		report.WriteString("]")
		if comp.Hits != comp.Shots {
			report.WriteString(fmt.Sprintf(" {%s, %.3f}", utils.FormatDurationString(comp.PenaltyResult.Time), comp.PenaltyResult.Speed))
		} else {
			report.WriteString(" {,}")
		}

		report.WriteString(fmt.Sprintf(" %d/%d", comp.Hits, comp.Shots))

		report.WriteString("\n")
	}

	return report.String()
}

// SaveReport writes the report string to a file by name.
func (ep *EventProcessor) SaveReport(filename string) error {
	report := ep.GenerateReport()
	return os.WriteFile(filename, []byte(report), 0644)
}

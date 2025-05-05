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
	"yadro-test/config"
	"yadro-test/messages"
	"yadro-test/models"
)

type EventProcessor struct {
	Config      config.Configuration
	Competitors map[int]*models.Competitor
	Events      []models.Event
	logFile     *os.File
	logWriter   *bufio.Writer
	mu          sync.Mutex
}

func NewEventProcessor(config config.Configuration) *EventProcessor {
	return &EventProcessor{
		Config:      config,
		Competitors: make(map[int]*models.Competitor),
		Events:      []models.Event{},
	}
}

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

func (ep *EventProcessor) EnableLogFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	ep.logFile = file
	ep.logWriter = bufio.NewWriter(file)
	return nil
}

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

	switch event.ID {
	case 1: //Участник зарегистрирован
		comp.Status = models.Registered
		ep.WriteLog(fmt.Sprintf(messages.Event1, event.TimeString, comp.ID))

	case 2: //Установлено время старта
		startTime, err := time.Parse(config.TimeFormat, event.ExtraParams)
		if err != nil {
			fmt.Printf("Error parsing start time: %v\n", err)
			return
		}
		comp.PlannedStart = startTime
		comp.LapStartTime = startTime
		comp.Status = models.Registered
		ep.WriteLog(fmt.Sprintf(messages.Event2, event.TimeString, comp.ID, event.ExtraParams))

	case 3: //Участник на стартовой линии
		comp.Status = models.OnStartLine
		ep.WriteLog(fmt.Sprintf(messages.Event3, event.TimeString, comp.ID))
	case 4: //Участник стартовал
		comp.ActualStart = event.Time
		comp.CurrentLap = 1
		comp.Status = models.Started
		ep.WriteLog(fmt.Sprintf(messages.Event4, event.TimeString, comp.ID))
	case 5: //Участник на огневом рубеже
		comp.Status = models.OnFiringRange
		ep.WriteLog(fmt.Sprintf(messages.Event5, event.TimeString, comp.ID, event.ExtraParams))
	case 6: //Мишень была поражена
		comp.Hits++
		ep.WriteLog(fmt.Sprintf(messages.Event6, event.TimeString, event.ExtraParams, comp.ID))
	case 7: //Участник покинул огневой рубеж
		comp.Shots += 5
		comp.Status = models.LeftFiringRange
		ep.WriteLog(fmt.Sprintf(messages.Event7, event.TimeString, comp.ID))
	case 8: //Участник заехал на штрафные круги
		comp.PenaltyStartTime = event.Time
		comp.Status = models.OnPenaltyLaps
		ep.WriteLog(fmt.Sprintf(messages.Event8, event.TimeString, comp.ID))
	case 9: //Участик покинул штрафные круги
		penaltyTime := event.Time.Sub(comp.PenaltyStartTime)
		misses := comp.Shots - comp.Hits
		penaltyDistance := misses * ep.Config.PenaltyLen

		comp.FullPenaltyTime = comp.FullPenaltyTime + penaltyTime

		var speed float64 = 0
		if penaltyTime.Seconds() > 0 {
			speed = float64(penaltyDistance) / comp.FullPenaltyTime.Seconds()
		}

		comp.PenaltyResult = models.PenaltyResult{
			Time:  comp.FullPenaltyTime,
			Speed: speed,
		}
		comp.Status = models.LeftPenaltyLaps
		ep.WriteLog(fmt.Sprintf(messages.Event9, event.TimeString, comp.ID))
	case 10: //Участник закончил основной круг
		lapTime := event.Time.Sub(comp.LapStartTime)
		speed := float64(ep.Config.LapLen) / lapTime.Seconds()
		comp.LapsResult = append(comp.LapsResult, models.LapResult{
			Time:  lapTime,
			Speed: speed,
		})
		comp.Status = models.FinishedLap
		ep.WriteLog(fmt.Sprintf(messages.Event10, event.TimeString, comp.ID))

		if comp.CurrentLap >= ep.Config.Laps {
			comp.Status = models.Finished
			comp.TotalTime = event.Time.Sub(comp.PlannedStart)
			ep.WriteLog(fmt.Sprintf(messages.Event33, event.TimeString, comp.ID))
		} else {
			comp.CurrentLap++
			comp.LapStartTime = event.Time
		}
	case 11: //Участник не может продолжить
		comp.Status = models.NotFinished
		ep.WriteLog(fmt.Sprintf(messages.Event11, event.TimeString, comp.ID, event.ExtraParams))
	}

	ep.Events = append(ep.Events, event)
}

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
			disqualificationTimeStr := models.FormatTimeString(endStartInterval.Add(time.Millisecond))
			ep.WriteLog(fmt.Sprintf(messages.Event32, disqualificationTimeStr, comp.ID))
		}

	}
}

func (ep *EventProcessor) ProcessEvents(events []models.Event) {
	for _, event := range events {
		ep.ProcessEvent(event)
	}
}

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
			report.WriteString(fmt.Sprintf("[%s] %d", models.FormatDurationString(comp.TotalTime), comp.ID))
		}

		report.WriteString(" [")
		for i := 0; i < ep.Config.Laps; i++ {
			if i > 0 {
				report.WriteString("}, {")
			} else {
				report.WriteString("{")
			}

			if i < len(comp.LapsResult) {
				lapResult := comp.LapsResult[i]
				report.WriteString(fmt.Sprintf("%s, %.3f", models.FormatDurationString(lapResult.Time), lapResult.Speed))
			} else {
				report.WriteString(",")
			}
		}
		report.WriteString("}]")

		report.WriteString(fmt.Sprintf(" {%s, %.3f}", models.FormatDurationString(comp.PenaltyResult.Time), comp.PenaltyResult.Speed))

		report.WriteString(fmt.Sprintf(" %d/%d", comp.Hits, comp.Shots))

		report.WriteString("\n")
	}

	return report.String()
}

func (ep *EventProcessor) SaveReport(filename string) error {
	report := ep.GenerateReport()
	return os.WriteFile(filename, []byte(report), 0644)
}

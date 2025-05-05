package models

import (
	"fmt"
	"time"
	"yadro-test/config"
)

const ()

type Event struct {
	Time         time.Time
	TimeString   string
	ID           int
	CompetitorID int
	ExtraParams  string
}

type CompetitorStatus int

const (
	Registered CompetitorStatus = iota
	OnStartLine
	Started
	OnFiringRange
	LeftFiringRange
	OnPenaltyLaps
	LeftPenaltyLaps
	FinishedLap
	Finished
	NotFinished
	NotStarted
)

type LapResult struct {
	Time  time.Duration
	Speed float64
}

type PenaltyResult struct {
	Time  time.Duration
	Speed float64
}

type Competitor struct {
	ID               int
	Status           CompetitorStatus
	PlannedStart     time.Time
	ActualStart      time.Time
	CurrentLap       int
	LapsResult       []LapResult
	PenaltyResult    PenaltyResult
	LapStartTime     time.Time
	PenaltyStartTime time.Time
	FullPenaltyTime  time.Duration
	Hits             int
	Shots            int
	TotalTime        time.Duration
	Comment          string
}

func FormatTimeString(t time.Time) string {
	return "[" + t.Format(config.TimeFormat) + "]"
}

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

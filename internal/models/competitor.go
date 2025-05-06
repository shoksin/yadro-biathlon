package models

import "time"

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
	LastFiringHits   int
	Hits             int
	Shots            int
	TotalTime        time.Duration
	Comment          string
}

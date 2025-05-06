package models

import "time"

type Action int

const (
	ActionRegistered      Action = iota + 1 // участник зарегистрирован
	ActionStartTimeSet                      // время старта установлено
	ActionOnStartLine                       // участник на стартовой линии
	ActionStarted                           // участник стартовал
	ActionOnFiringRange                     // участник на огневом рубеже
	ActionHit                               // мишень поражена
	ActionLeftFiringRange                   // участник покинул огневой рубеж
	ActionOnPenaltyLaps                     // участник на штрафных кругах
	ActionLeftPenaltyLaps                   // участник покинул штрафные круги
	ActionFinishedLap                       // участник закончил круг
	ActionCannotContinue                    // участник не может продолжить

)

type Event struct {
	Time         time.Time
	TimeString   string
	Action       Action
	CompetitorID int
	ExtraParams  string
}

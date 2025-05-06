package events

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"yadro-test/internal/config"
	"yadro-test/internal/models"
)

// ParseTime removes square brackets and parses a time string using the configured format.
func ParseTime(timeString string) (time.Time, error) {
	cleanTimeString := strings.Trim(timeString, "[]")
	return time.Parse(config.TimeFormat, cleanTimeString)
}

// ParseEvent converts a log line into a models.Event, validating action codes and parameters.
func ParseEvent(line string) (models.Event, error) {
	event := models.Event{}
	timeEndIndex := strings.Index(line, "]")
	if timeEndIndex == -1 {
		return models.Event{}, errors.New("invalid event")
	}

	timeString := line[:timeEndIndex+1]
	event.TimeString = timeString

	t, err := ParseTime(timeString)
	if err != nil {
		return event, fmt.Errorf("invalid time format: %v", err)
	}
	event.Time = t

	remainder := strings.TrimSpace(line[timeEndIndex+1:])
	parts := strings.Fields(remainder)
	if len(parts) < 2 {
		return event, errors.New("invalid event format: not enough parts")
	}

	actionInt, err := strconv.Atoi(parts[0])
	if err != nil {
		return event, fmt.Errorf("invalid event ID: %v", err)
	}
	action := models.Action(actionInt)

	//проверим, что это известное действие:
	switch action {
	case models.ActionRegistered,
		models.ActionStartTimeSet,
		models.ActionOnStartLine,
		models.ActionStarted,
		models.ActionOnFiringRange,
		models.ActionHit,
		models.ActionLeftFiringRange,
		models.ActionOnPenaltyLaps,
		models.ActionLeftPenaltyLaps,
		models.ActionFinishedLap,
		models.ActionCannotContinue:
		// всё ок
	default:
		return event, fmt.Errorf("unknown action ID: %d", actionInt)
	}
	event.Action = action

	competitorID, err := strconv.Atoi(parts[1])
	if err != nil {
		return event, fmt.Errorf("invalid competitor ID: %v", err)
	}
	event.CompetitorID = competitorID

	if len(parts) > 2 {
		event.ExtraParams = strings.Join(parts[2:], " ")
	}

	return event, nil
}

// LoadEvents opens a file, reads non-empty lines, and parses them into an event slice.
func LoadEvents(filename string) ([]models.Event, error) {
	var events []models.Event

	file, err := os.Open(filename)
	if err != nil {
		return events, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		event, err := ParseEvent(line)
		if err != nil {
			return events, fmt.Errorf("error parsing event '%s': %v", line, err)
		}
		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return events, err
	}

	return events, nil
}

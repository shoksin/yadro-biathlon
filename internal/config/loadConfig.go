package config

import (
	"encoding/json"
	"os"
)

// TimeFormat defines the layout for parsing and formatting timestamps in events.
const (
	TimeFormat = "15:04:05.000"
)

// Configuration holds race parameters, loaded from a JSON file.
type Configuration struct {
	Laps        int    `json:"laps"`
	LapLen      int    `json:"lapLen"`
	PenaltyLen  int    `json:"penaltyLen"`
	FiringLines int    `json:"firingLines"`
	Start       string `json:"start"`
	StartDelta  string `json:"startDelta"`
}

// LoadConfig reads and parses a JSON configuration file into Configuration.
// The JSON must match the struct tags, otherwise Decode will return an error.
func LoadConfig(filename string) (Configuration, error) {
	var config Configuration
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}

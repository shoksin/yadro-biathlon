package main

import (
	"flag"
	"fmt"
	"yadro-biathlon/internal/config"
	event "yadro-biathlon/internal/events"
	process "yadro-biathlon/internal/processor"
)

func main() {
	//Define command-line flags
	saveLogs := flag.String("save_logs", "", "save logs to file")
	eventsFile := flag.String("events_file", "./internal/config/events", "file with events")
	configFile := flag.String("config_file", "./internal/config/config.json", "file with config")
	resultFile := flag.String("result_file", "resultingTable", "file with results")
	flag.Parse()

	// 'conf' holds race parameters (laps, lap length, penalty length, etc.) and timing settings.
	conf, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("Error loading configuration(%s): %v\n", *configFile, err)
		return
	}

	//'processor' manages state, logs events, and generates the race report.
	processor := process.NewEventProcessor(conf)

	if *saveLogs != "" {
		err = processor.EnableLogFile(*saveLogs)
		if err != nil {
			fmt.Printf("Error opening log file(%s): %v\n", *saveLogs, err)
		}
	}

	events, err := event.LoadEvents(*eventsFile)
	if err != nil {
		fmt.Printf("Error loading events: %v\n", err)
		return
	}

	// Process all events
	processor.ProcessEvents(events)

	// Generate and save the report to the result file
	err = processor.SaveReport(*resultFile)
	if err != nil {
		fmt.Printf("Error saving report: %v\n", err)
	}

	fmt.Println("\nProcessing completed successfully")
	if *saveLogs != "" {
		fmt.Printf("Logs saved to: %s\n", *saveLogs)
	}
	fmt.Printf("Report saved to: %s\n", *resultFile)
}

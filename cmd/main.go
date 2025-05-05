package main

import (
	"flag"
	"fmt"
	"yadro-test/config"
	event "yadro-test/events"
	process "yadro-test/processor"
)

func main() {
	saveLogs := flag.String("save_logs", "", "save logs to file")
	eventsFile := flag.String("events_file", "./config/events", "file with events")
	configFile := flag.String("config_file", "./config/config.json", "file with config")
	resultFile := flag.String("result_file", "resultingTable.json", "file with results")
	flag.Parse()

	conf, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("Error loading configuration(%s): %v\n", *configFile, err)
		return
	}

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

	processor.ProcessEvents(events)

	err = processor.SaveReport(*resultFile)
	if err != nil {
		fmt.Printf("Error saving report: %v\n", err)
	}

	fmt.Println("\nProcessing completed successfully")
	if *saveLogs != "" {
		fmt.Printf("Logs saved to: %s\n", *saveLogs)
	}
	fmt.Printf("Report saved to:%s\n", *resultFile)
}

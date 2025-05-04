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
	eventsFile := flag.String("events_file", "./config/events.json", "file with events")
	flag.Parse()

	conf, err := config.LoadConfig("./config/config.json")
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
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

	err = processor.SaveLog("outputLog.txt")
	if err != nil {
		fmt.Printf("Error saving log: %v\n", err)
	}

	err = processor.SaveReport("resultingTable.txt")
	if err != nil {
		fmt.Printf("Error saving report: %v\n", err)
	}

	fmt.Println("Processing completed successfully")
}

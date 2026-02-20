package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/huckleberry-1881/ohgmas-watch/pkg/task"
)

func main() {
	// Define command line flags
	summaryFlag := flag.Bool("summary", false, "Generate a summary of work completed by tagset")
	tasksFlag := flag.Bool("tasks", false, "Include individual task details in summary (requires --summary)")
	startFlag := flag.String("start", "",
		"Filter segments to only include those closed after this datetime (RFC3339 format: 2006-01-02T15:04:05Z)")
	finishFlag := flag.String("finish", "",
		"Filter segments to only include those closed before this datetime (RFC3339 format: 2006-01-02T15:04:05Z)")
	fileFlag := flag.String("file", "",
		"Path to a custom YAML file for task storage (default: ~/.ohgmas-tasks.yaml)")

	flag.Parse()

	// Check if summary flag was provided
	if *summaryFlag {
		start, finish, err := parseTimeFlags(*startFlag, *finishFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		err = generateSummary(*tasksFlag, start, finish, *fileFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		return
	}

	// Check if tasks flag was provided without summary
	if *tasksFlag {
		fmt.Fprintf(os.Stderr, "Error: --tasks flag requires --summary flag\n")
		os.Exit(1)
	}

	// Determine which file to use
	tasksFilePath := *fileFlag
	if tasksFilePath == "" {
		tasksFilePath = task.GetTasksFilePath()
	}

	// Start TUI application
	app := NewApp(tasksFilePath)

	err := app.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}

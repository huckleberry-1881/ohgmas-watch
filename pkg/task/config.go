package task

import (
	"os"
	"path/filepath"
	"sync"
)

// Config holds configuration for the task package.
type Config struct {
	TasksFilePath string
	AutoSave      bool
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		TasksFilePath: GetDefaultTasksFilePath(),
		AutoSave:      false,
	}
}

// GetDefaultTasksFilePath returns the default path for tasks file.
func GetDefaultTasksFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return defaultTasksFileName
	}

	return filepath.Join(homeDir, defaultTasksFileName)
}

// NewWatchWithConfig creates a new Watch with configuration.
func NewWatchWithConfig(_ *Config) *Watch {
	return &Watch{
		Tasks: []*Task{},
		mu:    sync.RWMutex{},
	}
}
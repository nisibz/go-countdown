package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var saveFile string

func init() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	saveFile = filepath.Join(configDir, "go-countdown", "timers.json")
}

type saveData struct {
	Timers []Timer `json:"timers"`
}

func makeSaveData(m model) saveData {
	return saveData{
		Timers: m.timers,
	}
}

func applySaveData(m *model, s saveData) {
	m.timers = s.Timers
}

func saveToFile(m model) error {
	return saveTimers(m.timers)
}

func loadFromFile() (saveData, error) {
	timers, err := loadTimers()
	return saveData{Timers: timers}, err
}

// saveTimers saves timers directly (for CLI use)
func saveTimers(timers []Timer) error {
	data := saveData{Timers: timers}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(saveFile), 0o755); err != nil {
		return err
	}

	return os.WriteFile(saveFile, b, 0o644)
}

// loadTimers loads timers directly (for CLI use)
func loadTimers() ([]Timer, error) {
	var s saveData

	b, err := os.ReadFile(saveFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &s)
	if err != nil {
		return nil, err
	}

	return s.Timers, nil
}

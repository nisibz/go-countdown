package main

import (
	"encoding/json"
	"os"
)

const saveFile = "countdown.json"

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
	data := makeSaveData(m)

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(saveFile, b, 0o644)
}

func loadFromFile() (saveData, error) {
	var s saveData

	b, err := os.ReadFile(saveFile)
	if err != nil {
		return s, err
	}

	err = json.Unmarshal(b, &s)
	return s, err
}

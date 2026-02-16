package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DurationUnit string

const (
	UnitSmart   DurationUnit = "smart"   // Auto-detect based on current value
	UnitSeconds DurationUnit = "seconds"
	UnitMinutes DurationUnit = "minutes"
	UnitHours   DurationUnit = "hours"
)

type DurationAdjustConfig struct {
	Unit               DurationUnit `json:"unit"`
	IncrementStep      int          `json:"incrementStep"`      // e.g., 1, 5, 10
	ShiftIncrementStep int          `json:"shiftIncrementStep"` // for larger jumps
}

var configFile string

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		configFile = "config.json"
	} else {
		configFile = filepath.Join(homeDir, ".config", "go-countdown", "config.json")
	}
}

func defaultConfig() DurationAdjustConfig {
	return DurationAdjustConfig{
		Unit:               UnitSmart,
		IncrementStep:      1,
		ShiftIncrementStep: 5,
	}
}

func getConfigPath() string {
	return configFile
}

func loadConfig() (DurationAdjustConfig, error) {
	b, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Config doesn't exist, create default
			cfg := defaultConfig()
			if err := saveConfig(cfg); err != nil {
				log.Printf("warning: could not create default config: %v", err)
			}
			return cfg, nil
		}
		return DurationAdjustConfig{}, err
	}

	var cfg DurationAdjustConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		log.Printf("warning: malformed config file, using defaults: %v", err)
		return defaultConfig(), nil
	}

	// Validate and fill in missing values
	if cfg.Unit == "" {
		cfg.Unit = UnitSmart
	}
	if cfg.IncrementStep <= 0 {
		cfg.IncrementStep = 1
	}
	if cfg.ShiftIncrementStep <= 0 {
		cfg.ShiftIncrementStep = 5
	}

	return cfg, nil
}

func saveConfig(cfg DurationAdjustConfig) error {
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(configFile), 0o755); err != nil {
		return err
	}

	return os.WriteFile(configFile, b, 0o644)
}

// getUnitMultiplier returns the time duration multiplier based on the configured unit
// If unit is "smart", it detects the appropriate unit from the current input
func getUnitMultiplier(unit DurationUnit, currentInput string) time.Duration {
	switch unit {
	case UnitSeconds:
		return time.Second
	case UnitMinutes:
		return time.Minute
	case UnitHours:
		return time.Hour
	case UnitSmart:
		// Detect the largest unit in current input
		largestUnit := detectLargestUnit(currentInput)
		switch largestUnit {
		case "y":
			return 365 * 24 * time.Hour
		case "d":
			return 24 * time.Hour
		case "h":
			return time.Hour
		case "m":
			return time.Minute
		default:
			return time.Minute // Default to minutes for empty/small values
		}
	default:
		return time.Minute
	}
}

// detectLargestUnit finds the largest time unit suffix present in the input string
func detectLargestUnit(input string) string {
	// Check for units in order from largest to smallest
	unitPriority := []struct {
		suffix  string
		unitKey string
	}{
		{"y", "y"},
		{"d", "d"},
		{"h", "h"},
		{"m", "m"},
		{"s", "s"},
	}

	for _, u := range unitPriority {
		if containsUnit(input, u.suffix) {
			return u.unitKey
		}
	}
	return "" // No unit detected
}

// containsUnit checks if the input contains the given unit suffix
func containsUnit(input, suffix string) bool {
	inputLower := input
	for i := 0; i < len(inputLower); i++ {
		// Look for the suffix preceded by a digit
		if inputLower[i] == suffix[0] && i > 0 {
			// Check if preceded by a digit
			prev := inputLower[i-1]
			if prev >= '0' && prev <= '9' {
				return true
			}
		}
	}
	return false
}

// formatForInput formats a duration into a compact string suitable for the duration input
// This is similar to formatDuration but more compact for form editing
func formatForInput(d time.Duration) string {
	d = d.Round(time.Second)
	totalSeconds := int(d.Seconds())

	if totalSeconds <= 0 {
		return ""
	}

	days := totalSeconds / 86400
	hours := (totalSeconds % 86400) / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	var parts []string

	// For long durations, use top 2 units
	if days > 0 {
		years := days / 365
		remainingDays := days % 365
		months := remainingDays / 30

		if years > 0 {
			parts = append(parts, fmt.Sprintf("%dy", years))
		}
		if months > 0 {
			parts = append(parts, fmt.Sprintf("%dmo", months))
		}
		if remainingDays%30 > 0 && len(parts) < 2 {
			parts = append(parts, fmt.Sprintf("%dd", remainingDays%30))
		}
	} else if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
		if minutes > 0 {
			parts = append(parts, fmt.Sprintf("%dm", minutes))
		}
	} else if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
		if seconds > 0 && len(parts) < 2 {
			parts = append(parts, fmt.Sprintf("%ds", seconds))
		}
	} else {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	return strings.Join(parts, " ")
}

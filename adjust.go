package main

import (
	"time"
)

// adjustDuration modifies a duration string by adding/subtracting time
// Returns the new duration string formatted for display
func adjustDuration(currentInput string, delta time.Duration, config DurationAdjustConfig) string {
	// Parse current duration (treat empty as 0)
	currentDur, err := parseDuration(currentInput)
	if err != nil {
		currentDur = 0
	}

	// Apply delta
	newDur := currentDur + delta

	// Clamp minimum at 1 second (no negative/zero durations)
	if newDur < time.Second {
		newDur = time.Second
	}

	// Format back to string
	return formatForInput(newDur)
}

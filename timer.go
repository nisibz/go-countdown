package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Timer struct {
	Name      string        `json:"name"`
	End       time.Time     `json:"end"`
	Paused    bool          `json:"paused"`
	Remaining time.Duration `json:"remaining"`
	Duration  time.Duration `json:"duration"`
}

func parseDuration(input string) (time.Duration, error) {
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return 0, fmt.Errorf("empty input")
	}

	var total time.Duration

	// Parse multiple number-suffix pairs (e.g., "30d30m", "1h30m", "2d")
	i := 0
	for i < len(input) {
		// Skip spaces between components
		if input[i] == ' ' {
			i++
			continue
		}

		// Parse number
		numStart := i
		for i < len(input) && input[i] >= '0' && input[i] <= '9' {
			i++
		}
		if i == numStart {
			return 0, fmt.Errorf("expected number at position %d", numStart)
		}
		numStr := input[numStart:i]

		num, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, fmt.Errorf("invalid number %s: %w", numStr, err)
		}

		if num <= 0 {
			return 0, fmt.Errorf("duration must be positive")
		}

		// Parse suffix (default to seconds if at end of input)
		var suffix string
		if i < len(input) {
			suffix = input[i : i+1]
			i++
		} else {
			// No suffix means seconds (e.g., "30" = 30s)
			suffix = "s"
		}

		var d time.Duration
		switch suffix {
		case "s":
			d = time.Duration(num) * time.Second
		case "m":
			d = time.Duration(num) * time.Minute
		case "h":
			d = time.Duration(num) * time.Hour
		case "d":
			d = time.Duration(num) * 24 * time.Hour
		case "y":
			d = time.Duration(num) * 365 * 24 * time.Hour
		default:
			return 0, fmt.Errorf("invalid suffix: %s (use s, m, h, d, y)", suffix)
		}
		total += d
	}

	if total <= 0 {
		return 0, fmt.Errorf("duration must be positive")
	}

	return total, nil
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	totalSeconds := int(d.Seconds())

	days := totalSeconds / 86400
	hours := (totalSeconds % 86400) / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	// Calculate months and years for larger durations
	years := days / 365
	remainingDaysAfterYears := days % 365
	months := remainingDaysAfterYears / 30
	remainingDays := remainingDaysAfterYears % 30

	var parts []string

	// For large durations (> 60 days), show only top 2 units (years, months, or days)
	// This keeps the display compact and prevents truncation
	if days > 60 {
		if years > 0 {
			parts = append(parts, fmt.Sprintf("%dy", years))
		}
		if months > 0 {
			parts = append(parts, fmt.Sprintf("%dmo", months))
		}
		if remainingDays > 0 && len(parts) < 2 {
			parts = append(parts, fmt.Sprintf("%dd", remainingDays))
		}
		// Show at most 2 parts for long durations
		if len(parts) > 2 {
			parts = parts[:2]
		}
		return strings.Join(parts, " ")
	}

	// For shorter durations, show more detail
	if years > 0 {
		parts = append(parts, fmt.Sprintf("%dy", years))
	}
	if months > 0 {
		parts = append(parts, fmt.Sprintf("%dmo", months))
	}
	if remainingDays > 0 {
		parts = append(parts, fmt.Sprintf("%dd", remainingDays))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	if len(parts) > 2 {
		return strings.Join(parts[:2], " ") + " " + strings.Join(parts[2:], " ")
	}
	return strings.Join(parts, " ")
}

func (t Timer) StatusEmoji(now time.Time) string {
	if t.Paused {
		return "⏸️"
	}
	remaining := time.Until(t.End)
	if remaining <= 0 {
		return "✅"
	}
	return "⏳️"
}

func (t Timer) StatusText(now time.Time) string {
	if t.Paused {
		return formatDuration(t.Remaining)
	}
	remaining := time.Until(t.End)
	if remaining <= 0 {
		return "Done"
	}
	return formatDuration(remaining)
}

func (t Timer) EndTimeText(now time.Time) string {
	if t.Paused {
		return "(paused)"
	}
	remaining := time.Until(t.End)
	if remaining <= 0 {
		elapsed := t.Duration - remaining
		return fmt.Sprintf("+%s", formatDuration(elapsed))
	}
	return formatEndTime(t.End, now)
}

func formatEndTime(end, now time.Time) string {
	if end.Day() == now.Day() && end.Month() == now.Month() && end.Year() == now.Year() {
		return end.Format("15:04:05")
	} else if end.Month() == now.Month() && end.Year() == now.Year() {
		return end.Format("2 15:04")
	} else if end.Year() == now.Year() {
		return end.Format("2/01 15:04")
	} else {
		return end.Format("2/01/06 15:04")
	}
}

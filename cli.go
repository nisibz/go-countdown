package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func printUsage() {
	fmt.Println("go-countdown - Terminal countdown timer")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  go-countdown              # Launch TUI interface")
	fmt.Println("  go-countdown <command>    # Run CLI command")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  add <name> <duration>           Add a new timer")
	fmt.Println("  list [--filter]                 List timers (filter: --active, --paused, --done)")
	fmt.Println("  pause [--all] [filter] <index>  Pause timer(s) by 1-based index, or --all")
	fmt.Println("  resume [--all] [filter] <index> Resume timer(s) by 1-based index, or --all")
	fmt.Println("  delete [--done|--all] [filter] <index>  Delete timer(s) by index, --done, or --all")
	fmt.Println("  restart [--all|--active|--paused] [filter] <index>  Restart timer(s)")
	fmt.Println("  edit [filter] <index> <name> <duration>  Edit timer")
	fmt.Println("  help                            Show this help")
	fmt.Println()
	fmt.Println("COMMAND SHORTCUTS:")
	fmt.Println("  a         add")
	fmt.Println("  l         list")
	fmt.Println("  p         pause")
	fmt.Println("  r         resume")
	fmt.Println("  rs        restart")
	fmt.Println("  d         delete")
	fmt.Println("  e         edit")
	fmt.Println("  h         help")
	fmt.Println()
	fmt.Println("BULK OPERATIONS:")
	fmt.Println("  pause --all              Pause all active timers")
	fmt.Println("  resume --all             Resume all paused timers")
	fmt.Println("  delete --done            Delete all completed timers")
	fmt.Println("  delete --all             Delete ALL timers (requires confirmation)")
	fmt.Println("  restart --all            Restart all timers")
	fmt.Println("  restart --active         Restart all active timers")
	fmt.Println("  restart --paused         Restart all paused timers")
	fmt.Println()
	fmt.Println("DURATION FORMAT:")
	fmt.Println("  30s    30 seconds")
	fmt.Println("  5m     5 minutes")
	fmt.Println("  1h     1 hour")
	fmt.Println("  2d     2 days")
	fmt.Println("  1y     1 year")
	fmt.Println("  30d30m Compound: 30 days 30 minutes")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  go-countdown a \"Meeting\" 30m")
	fmt.Println("  go-countdown l                    # List all timers")
	fmt.Println("  go-countdown l --active           # List only active timers")
	fmt.Println("  go-countdown p 1                  # Pause first timer")
	fmt.Println("  go-countdown p --all              # Pause all active timers")
	fmt.Println("  go-countdown r --paused 2         # Resume second paused timer")
	fmt.Println("  go-countdown d --done             # Delete all completed timers")
	fmt.Println("  go-countdown rs --all             # Restart all timers")
}

func parseFilterAndIndex(args []string) (filter, indexStr string, idx int) {
	if len(args) == 0 {
		return "", "", 0
	}
	if strings.HasPrefix(args[0], "--") {
		filter = args[0]
		if len(args) < 2 {
			return filter, "", 0
		}
		indexStr = args[1]
	} else {
		indexStr = args[0]
	}
	var err error
	idx, err = strconv.Atoi(indexStr)
	if err != nil || idx < 1 {
		return filter, indexStr, 0
	}
	return filter, indexStr, idx
}

func getFilteredTimers(timers []Timer, filter string) []Timer {
	now := time.Now()
	var result []Timer
	for _, t := range timers {
		switch filter {
		case "--active":
			if !t.Paused && t.End.After(now) {
				result = append(result, t)
			}
		case "--paused":
			if t.Paused {
				result = append(result, t)
			}
		case "--done":
			if !t.Paused && !t.End.After(now) {
				result = append(result, t)
			}
		default:
			result = append(result, t)
		}
	}
	return result
}

func resolveIndex(timers []Timer, filter string, idx int) (int, error) {
	if idx < 1 {
		return -1, fmt.Errorf("index must be >= 1")
	}
	filtered := getFilteredTimers(timers, filter)
	if idx > len(filtered) {
		return -1, fmt.Errorf("index %d out of range (filter shows %d timer(s))", idx, len(filtered))
	}
	targetTimer := filtered[idx-1]
	for i, t := range timers {
		if t.Name == targetTimer.Name && t.End.Equal(targetTimer.End) {
			return i, nil
		}
	}
	return -1, fmt.Errorf("timer not found")
}

func formatEndTimeCLI(end, now time.Time) string {
	if end.Day() == now.Day() && end.Month() == now.Month() && end.Year() == now.Year() {
		return end.Format("15:04:05")
	} else if end.Month() == now.Month() && end.Year() == now.Year() {
		return end.Format("Jan 2 15:04")
	} else if end.Year() == now.Year() {
		return end.Format("Jan 2")
	} else {
		return end.Format("2006-01-02")
	}
}

func listTimers(timers []Timer, filter string) {
	now := time.Now()
	filtered := getFilteredTimers(timers, filter)

	fmt.Println("Countdown Timers")
	fmt.Println("================")
	fmt.Println()

	if len(filtered) == 0 {
		fmt.Println("No timers found.")
		return
	}

	for i, t := range filtered {
		var statusEmoji, remainingText, endTimeText string

		if t.Paused {
			statusEmoji = "[paused]"
			remainingText = formatDuration(t.Remaining)
			endTimeText = ""
		} else {
			remaining := time.Until(t.End)
			if remaining <= 0 {
				statusEmoji = "[done]"
				remainingText = "Done"
				elapsed := t.Duration - remaining
				endTimeText = fmt.Sprintf("(+%s elapsed)", formatDuration(elapsed))
			} else {
				statusEmoji = "[active]"
				remainingText = formatDuration(remaining)
				endTimeText = fmt.Sprintf("(ends %s)", formatEndTimeCLI(t.End, now))
			}
		}

		fmt.Printf("[%d] %s %-30s %-13s", i+1, statusEmoji, t.Name, remainingText)
		if endTimeText != "" {
			fmt.Printf(" %s", endTimeText)
		}
		fmt.Println()
	}

	fmt.Printf("\nShowing %d timer(s)\n", len(filtered))
}

func executeCLICommand(cmd string, args []string) error {
	// Load timers for CLI commands
	timers, err := loadTimers()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error loading timers: %w", err)
	}
	dirty := false

	switch cmd {
	case "add":
		if len(args) < 2 {
			fmt.Println("Usage: go-countdown add <name> <duration>")
			fmt.Println("\nDuration examples: 30s, 5m, 1h, 2d, 1y, 30d30m, 1h30m")
			return nil
		}
		name := args[0]
		duration := args[1]
		d, err := parseDuration(duration)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
		newTimer := Timer{
			Name:     name,
			End:      time.Now().Add(d),
			Duration: d,
		}
		timers = append(timers, newTimer)
		dirty = true
		fmt.Printf("Added timer \"%s\" (%s)\n", name, formatDuration(d))

	case "list":
		filter := ""
		if len(args) > 0 && strings.HasPrefix(args[0], "--") {
			filter = args[0]
		}
		listTimers(timers, filter)

	case "pause":
		// Check for --all flag
		if len(args) > 0 && args[0] == "--all" {
			count := 0
			now := time.Now()
			for i := range timers {
				if !timers[i].Paused && timers[i].End.After(now) {
					timers[i].Remaining = time.Until(timers[i].End)
					timers[i].Paused = true
					count++
				}
			}
			if count > 0 {
				dirty = true
			}
			fmt.Printf("Paused %d timer(s)\n", count)
		} else {
			filter, _, idx := parseFilterAndIndex(args)
			actualIdx, err := resolveIndex(timers, filter, idx)
			if err != nil {
				return err
			}
			if actualIdx >= 0 && len(timers) > 0 {
				t := &timers[actualIdx]
				if !t.Paused {
					if t.End.After(time.Now()) {
						t.Remaining = time.Until(t.End)
						t.Paused = true
						dirty = true
						fmt.Printf("Paused timer \"%s\"\n", t.Name)
					} else {
						return fmt.Errorf("cannot pause: timer already done")
					}
				} else {
					fmt.Printf("Timer \"%s\" is already paused\n", t.Name)
				}
			}
		}

	case "resume":
		// Check for --all flag
		if len(args) > 0 && args[0] == "--all" {
			count := 0
			for i := range timers {
				if timers[i].Paused && timers[i].Remaining > 0 {
					timers[i].End = time.Now().Add(timers[i].Remaining)
					timers[i].Paused = false
					count++
				}
			}
			if count > 0 {
				dirty = true
			}
			fmt.Printf("Resumed %d timer(s)\n", count)
		} else {
			filter, _, idx := parseFilterAndIndex(args)
			actualIdx, err := resolveIndex(timers, filter, idx)
			if err != nil {
				return err
			}
			if actualIdx >= 0 && len(timers) > 0 {
				t := &timers[actualIdx]
				if t.Paused {
					if t.Remaining > 0 {
						t.End = time.Now().Add(t.Remaining)
						t.Paused = false
						dirty = true
						fmt.Printf("Resumed timer \"%s\"\n", t.Name)
					} else {
						return fmt.Errorf("cannot resume: no remaining time")
					}
				} else {
					fmt.Printf("Timer \"%s\" is already active\n", t.Name)
				}
			}
		}

	case "delete":
		// Check for --done or --all flags
		if len(args) > 0 && args[0] == "--done" {
			now := time.Now()
			newTimers := make([]Timer, 0, len(timers))
			count := 0
			for _, t := range timers {
				if !t.Paused && !t.End.After(now) {
					count++
				} else {
					newTimers = append(newTimers, t)
				}
			}
			if count > 0 {
				dirty = true
			}
			timers = newTimers
			fmt.Printf("Deleted %d completed timer(s)\n", count)
		} else if len(args) > 0 && args[0] == "--all" {
			// Require confirmation for delete --all
			fmt.Print("Delete all timers? [y/N]: ")
			var response string
			_, _ = fmt.Scanln(&response)
			if strings.ToLower(response) == "y" {
				count := len(timers)
				timers = []Timer{}
				dirty = true
				fmt.Printf("Deleted %d timer(s)\n", count)
			} else {
				fmt.Println("Cancelled")
			}
		} else {
			filter, _, idx := parseFilterAndIndex(args)
			actualIdx, err := resolveIndex(timers, filter, idx)
			if err != nil {
				return err
			}
			if actualIdx >= 0 && len(timers) > 0 {
				deletedName := timers[actualIdx].Name
				timers = append(timers[:actualIdx], timers[actualIdx+1:]...)
				dirty = true
				fmt.Printf("Deleted timer \"%s\"\n", deletedName)
			}
		}

	case "restart":
		// Check for --all, --active, or --paused flags
		if len(args) > 0 && args[0] == "--all" {
			count := 0
			for i := range timers {
				if timers[i].Duration > 0 {
					timers[i].End = time.Now().Add(timers[i].Duration)
					timers[i].Paused = false
					timers[i].Remaining = 0
					count++
				}
			}
			if count > 0 {
				dirty = true
			}
			fmt.Printf("Restarted %d timer(s)\n", count)
		} else if len(args) > 0 && args[0] == "--active" {
			now := time.Now()
			count := 0
			for i := range timers {
				if !timers[i].Paused && timers[i].End.After(now) && timers[i].Duration > 0 {
					timers[i].End = time.Now().Add(timers[i].Duration)
					timers[i].Paused = false
					timers[i].Remaining = 0
					count++
				}
			}
			if count > 0 {
				dirty = true
			}
			fmt.Printf("Restarted %d active timer(s)\n", count)
		} else if len(args) > 0 && args[0] == "--paused" {
			count := 0
			for i := range timers {
				if timers[i].Paused && timers[i].Duration > 0 {
					timers[i].End = time.Now().Add(timers[i].Duration)
					timers[i].Paused = false
					timers[i].Remaining = 0
					count++
				}
			}
			if count > 0 {
				dirty = true
			}
			fmt.Printf("Restarted %d paused timer(s)\n", count)
		} else {
			filter, _, idx := parseFilterAndIndex(args)
			actualIdx, err := resolveIndex(timers, filter, idx)
			if err != nil {
				return err
			}
			if actualIdx >= 0 && len(timers) > 0 && timers[actualIdx].Duration > 0 {
				t := &timers[actualIdx]
				t.End = time.Now().Add(t.Duration)
				t.Paused = false
				t.Remaining = 0
				dirty = true
				fmt.Printf("Restarted timer \"%s\"\n", t.Name)
			}
		}

	case "edit":
		if len(args) < 3 {
			fmt.Println("Usage: go-countdown edit [--filter] <index> <name> <duration>")
			fmt.Println("\nExamples:")
			fmt.Println("  go-countdown edit 1 \"New Name\" 10m")
			fmt.Println("  go-countdown edit --active 1 \"New Name\" 10m")
			return nil
		}

		var filter, indexStr, name, durationStr string
		if strings.HasPrefix(args[0], "--") {
			filter = args[0]
			indexStr = args[1]
			name = args[2]
			if len(args) > 3 {
				durationStr = args[3]
			}
		} else {
			indexStr = args[0]
			name = args[1]
			if len(args) > 2 {
				durationStr = args[2]
			}
		}

		idx, err := strconv.Atoi(indexStr)
		if err != nil || idx < 1 {
			return fmt.Errorf("invalid index: %s", indexStr)
		}

		actualIdx, err := resolveIndex(timers, filter, idx)
		if err != nil {
			return err
		}

		if actualIdx >= 0 && len(timers) > 0 {
			t := &timers[actualIdx]
			oldName := t.Name

			// Update name if provided
			if name != "" {
				t.Name = name
			}

			// Update duration if provided
			if durationStr != "" {
				d, err := parseDuration(durationStr)
				if err != nil {
					return fmt.Errorf("invalid duration: %w", err)
				}
				t.Duration = d
				t.End = time.Now().Add(d)
				t.Paused = false
				t.Remaining = 0
			}

			dirty = true
			fmt.Printf("Edited timer: \"%s\" -> \"%s\"\n", oldName, t.Name)
		}

	case "help", "-h", "--help":
		printUsage()

	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}

	// Save if any changes were made
	if dirty {
		if err := saveTimers(timers); err != nil {
			return fmt.Errorf("error saving timers: %w", err)
		}
	}

	return nil
}

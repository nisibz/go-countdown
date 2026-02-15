package main

func (m model) getVisibleTimers() []Timer {
	var result []Timer
	for _, t := range m.timers {
		switch m.filter {
		case filterAll:
			result = append(result, t)
		case filterActive:
			if !t.Paused && t.End.After(m.now) {
				result = append(result, t)
			}
		case filterPaused:
			if t.Paused {
				result = append(result, t)
			}
		case filterDone:
			if !t.Paused && !t.End.After(m.now) {
				result = append(result, t)
			}
		}
	}
	return result
}

func (m model) getActualTimerIndex(visibleIndex int) int {
	visibleTimers := m.getVisibleTimers()
	if visibleIndex < 0 || visibleIndex >= len(visibleTimers) {
		return -1
	}
	targetTimer := visibleTimers[visibleIndex]
	for i, t := range m.timers {
		if t.Name == targetTimer.Name && t.End.Equal(targetTimer.End) {
			return i
		}
	}
	return -1
}

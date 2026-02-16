# go-countdown

A terminal countdown timer application written in Go using the Charm Tea framework for TUI and a separate CLI interface.

![Build](https://github.com/nisibz/go-countdown/workflows/Go/badge.svg)

## Features

- **TUI Interface**: Interactive terminal UI with keyboard navigation
- **CLI Commands**: Quick command-line operations for managing timers
- **Multiple Timers**: Track multiple countdown timers simultaneously
- **Timer States**: Active, paused, and completed timers
- **Filtering**: View all, active, paused, or completed timers
- **Bulk Operations**: Pause, resume, restart, or delete all timers at once
- **Keyboard Shortcuts**: +/- keys to quickly adjust duration when adding/editing timers
- **Persistence**: Timers are saved automatically and restored on restart

## Installation

```bash
go install github.com/nisibz/go-countdown@latest
```

Or build from source:

```bash
git clone https://github.com/nisibz/go-countdown.git
cd go-countdown
go build -o countdown .
```

## Usage

### TUI Mode (Default)

Run the application to launch the interactive terminal UI:

```bash
go run .
# or
./countdown
```

#### TUI Keybindings

| Key | Action |
|-----|--------|
| `a` | Add a new timer |
| `e` | Edit selected timer |
| `d` | Delete selected timer (with confirmation) |
| `p` | Pause/resume selected timer |
| `r` | Restart selected timer (with confirmation) |
| `P` | Pause all active timers |
| `R` | Restart all timers (with confirmation) |
| `Shift+R` | Resume all paused timers |
| `D` | Delete all completed timers |
| `↑/k` | Move cursor up |
| `↓/j` | Move cursor down |
| `ctrl+↑/k` | Reorder timer up |
| `ctrl+↓/j` | Reorder timer down |
| `tab` | Cycle filter mode |
| `1-4` | Filter: All/Active/Paused/Done |
| `?` | Toggle help |
| `q` | Quit |

#### Duration Adjustment (+/-)

When adding or editing a timer, use the `+` and `-` keys to quickly adjust the duration:

- `+` or `=`: Increase duration
- `-` or `_`: Decrease duration
- Minimum duration is 1 second

**Smart Unit Detection**: The adjustment automatically detects which unit to use:
- If duration contains "h" (e.g., "1h30m"), adjustment adds hours
- If duration contains "m" (e.g., "30m"), adjustment adds minutes
- Empty input defaults to minutes

### CLI Mode

```bash
# Add a timer
./countdown add "My Timer" 30m
./countdown add "Meeting" 1h30m

# List all timers
./countdown list

# Pause a timer (by index)
./countdown pause 0

# Resume a timer
./countdown resume 0

# Delete a timer
./countdown delete 0

# Restart a timer
./countdown restart 0

# Show help
./countdown help
```

## Configuration

### Duration Adjustment

The +/- key behavior can be customized via a configuration file.

**Config Location**:
- **All platforms**: `~/.config/go-countdown/config.json`

The config file is automatically created with defaults on first run:

```json
{
  "unit": "smart",
  "incrementStep": 1,
  "shiftIncrementStep": 5
}
```

#### Configuration Options

| Option | Type | Description |
|--------|------|-------------|
| `unit` | string | Time unit for adjustments: `"smart"`, `"seconds"`, `"minutes"`, `"hours"` |
| `incrementStep` | number | Amount to add/subtract when pressing +/- (default: 1) |
| `shiftIncrementStep` | number | Larger step size for Shift+/- (default: 5, future use) |

#### Unit Modes

- **`"smart"`** (default): Auto-detects unit based on current duration value
- **`"seconds"`**: Always adjust by seconds
- **`"minutes"`**: Always adjust by minutes
- **`"hours"`**: Always adjust by hours

#### Example Configurations

**5-minute increments:**
```json
{
  "unit": "minutes",
  "incrementStep": 5,
  "shiftIncrementStep": 15
}
```

**Always adjust by hours:**
```json
{
  "unit": "hours",
  "incrementStep": 1,
  "shiftIncrementStep": 2
}
```

**Smart mode with custom steps:**
```json
{
  "unit": "smart",
  "incrementStep": 10,
  "shiftIncrementStep": 30
}
```

### Duration Format

When adding or editing timers, use these formats:

- `30s` - 30 seconds
- `5m` - 5 minutes
- `1h` - 1 hour
- `2d` - 2 days
- `1y` - 1 year
- `1h30m` - 1 hour 30 minutes
- `30d12h` - 30 days 12 hours
- `30` - 30 seconds (default when no suffix)

## Development

### Project Structure

| File | Purpose |
|------|---------|
| `main.go` | Entry point, Update logic, CLI command routing |
| `tui.go` | Model struct, initialization (`initialModel`) |
| `view.go` | View rendering, popup overlays, table styles |
| `keys.go` | Keybinding definitions (3 keymaps for different states) |
| `timer.go` | Domain logic (Timer struct, duration parsing/formatting) |
| `storage.go` | Persistence layer (load/save to JSON) |
| `cli.go` | CLI command execution |
| `config.go` | Configuration system for duration adjustment |
| `adjust.go` | Duration adjustment logic (+/- keys) |

### Build & Run

```bash
# Run TUI (default)
go run .

# Run CLI commands
go run . add "My Timer" 30m
go run . list

# Build binary
go build -o countdown .

# Run tests
go test ./...
```

## License

MIT License

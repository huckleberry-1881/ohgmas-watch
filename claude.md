# ohgmas-watch

## Project Overview

ohgmas-watch is a terminal-based time tracking utility written in Go that helps users monitor and analyze how they spend their work time. It provides an interactive TUI (Terminal User Interface) for managing tasks and tracking time segments, along with powerful reporting capabilities to generate weekly summaries grouped by task categories and tags.

## Purpose

The application serves as a time management aid to:
- Track time spent on different work tasks
- Organize tasks with tags and categories
- Monitor active and completed work
- Generate reports showing time allocation across weeks and task categories
- Help identify what's consuming the most time in a workday

## Architecture

### Project Structure

```
ohgmas-watch/
├── cmd/ow/               # Main application entry point
│   ├── main.go          # CLI entry point and flag parsing
│   ├── ui.go            # TUI implementation (App struct and components)
│   ├── summary.go       # CLI summary generation
│   ├── helpers.go       # Utility functions for formatting and time parsing
│   ├── helpers_test.go  # Tests for helper functions
│   └── summary_test.go  # Tests for summary generation
├── pkg/task/            # Core business logic package
│   ├── types.go         # Data structure definitions
│   ├── task.go          # Task management operations
│   ├── segment.go       # Time segment operations
│   ├── summary.go       # Reporting and summary generation
│   ├── interfaces.go    # Interface definitions
│   ├── task_test.go     # Tests for task operations
│   ├── segment_test.go  # Tests for segment operations
│   ├── summary_test.go  # Tests for summary generation
│   └── persistence_test.go # Tests for file I/O
├── .github/workflows/   # CI/CD configuration
│   ├── go.yml           # Build and test workflow
│   └── golangci.yml     # Linting workflow
├── .gitignore           # Git ignore rules
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
└── README.md            # Project description
```

### Core Components

#### 1. Watch (`pkg/task/types.go`)
The `Watch` struct is the main container for all tasks being tracked:
- Holds a collection of tasks
- Thread-safe operations using `sync.RWMutex`
- Persists to YAML file (`~/.ohgmas-tasks.yaml` by default)

#### 2. Task (`pkg/task/types.go`)
Represents an individual work item:
- **Fields:**
  - `Name`: Task identifier
  - `Description`: Detailed description
  - `Tags`: Array of tags for categorization
  - `Category`: Status category (work, completed, backlog)
  - `Segments`: Array of time tracking segments
- **Thread-safe:** Uses `sync.RWMutex` for concurrent access
- **Key Methods:**
  - `AddSegment()`: Start tracking time
  - `CloseSegment()`: Stop tracking time
  - `GetClosedSegmentsDuration()`: Calculate total tracked time
  - `SetCategory()`: Change task status
  - `GetLastActivity()`: Get most recent activity timestamp

#### 3. Segment (`pkg/task/types.go`)
Represents a time tracking period:
- **Fields:**
  - `Create`: Start timestamp
  - `Finish`: End timestamp (zero value = open/active)
  - `Note`: Optional note about the work performed
- **Lifecycle:** Created open, then closed when work stops

## Technologies & Dependencies

### Core Technologies
- **Language:** Go 1.24.5
- **TUI Framework:** [tview](https://github.com/rivo/tview) - Rich terminal UI framework
- **Terminal Library:** [tcell](https://github.com/gdamore/tcell) - Low-level terminal handling
- **Serialization:** [go-yaml](https://github.com/goccy/go-yaml) - YAML marshaling/unmarshaling

### Development Tools
- **Linting:** golangci-lint (configured in `.golangci.yml`)
- **CI/CD:** GitHub Actions
  - Automated testing on push to main and PRs
  - Go build verification
  - Linter checks

## Data Model

### Task Categories
Tasks can be organized into three categories:
1. **work** - Active tasks currently being worked on (default)
2. **completed** - Finished tasks
3. **backlog** - Tasks planned for future work

### Tags
Tasks can have multiple tags for flexible categorization. Tags are used for:
- Grouping related tasks in summaries
- Filtering and organization
- Reporting by tag combinations (tagsets)

### Time Tracking
- Time is tracked via **Segments** - discrete periods of work
- Each task can have multiple segments
- Only one segment can be open at a time per task
- Segments can have optional notes describing the work done

### Data Persistence
- Tasks are stored in YAML format
- Default location: `~/.ohgmas-tasks.yaml`
- Custom location can be specified with `--file` flag
- Thread-safe read/write operations
- Automatic saving after each modification

## Usage

### Interactive TUI Mode (Default)

Launch the application:
```bash
./ow
```

Or with a custom data file:
```bash
./ow --file /path/to/tasks.yaml
```

#### TUI Controls

**Navigation:**
- `↑/↓` - Navigate between tasks
- `Enter` - View detailed segment history for selected task

**Task Management:**
- `t` - Create new task
- `m` - Modify selected task (name, description, tags)
- `d` - Delete selected task (with confirmation)

**Time Tracking:**
- `s` - Start new segment (without note)
- `n` - Start new segment with note
- `e` - End active segment

**Category Management:**
- `c` - Mark task as completed
- `w` - Mark task as work (active)
- `b` - Mark task as backlog

**Filtering:**
- `f` - Cycle through category filters (all → completed → work → backlog)

**Exit:**
- `Ctrl+C` - Exit application

#### TUI Display

The main table shows:
- **Status** - Active (▶) or inactive (●) indicator
- **Task Name** - Name of the task
- **Category** - Current category with color coding
  - Green: completed
  - Yellow: work
  - Gray: backlog
- **Tags** - Associated tags in parentheses
- **Last Activity** - Date of most recent segment
- **This Week** - Time tracked since last Monday
- **Duration** - Total tracked time for task

The description pane shows:
- Current or last segment information
- Segment duration
- Task description

### Summary/Report Mode

Generate weekly summaries grouped by tag combinations:

```bash
./ow --summary
```

Include individual task breakdowns:
```bash
./ow --summary --tasks
```

Filter by time range:
```bash
./ow --summary --start 2024-01-01T00:00:00Z --finish 2024-12-31T23:59:59Z
```

Use custom data file:
```bash
./ow --summary --file /path/to/tasks.yaml
```

#### Summary Output Format

```
Week starting 12/04/2024
- client-work, feature-dev [15h30m]
-- Project Alpha [8h45m]
-- Project Beta [6h45m]
- internal, meetings [3h15m]
-- Team Standup [2h00m]
-- Planning Session [1h15m]

Week starting 12/11/2024
- client-work, bug-fixes [12h00m]
-- Critical Bug Fix [9h30m]
-- Code Review [2h30m]
```

## Building & Development

### Build

```bash
go build -o ow ./cmd/ow
```

### Run Tests

```bash
go test -v ./...
```

Run tests with coverage:
```bash
go test -cover ./pkg/task/...
```

Generate detailed coverage report:
```bash
go test -coverprofile=coverage.out ./pkg/task/... && go tool cover -func=coverage.out
```

### Test Coverage

| Package | Coverage |
|---------|----------|
| `pkg/task` | **97.5%** |
| `cmd/ow` | **18.8%** |

**pkg/task** has comprehensive tests for all business logic:

| Test File | Description |
|-----------|-------------|
| [task_test.go](pkg/task/task_test.go) | Core task operations (AddTask, AddSegment, CloseSegment, categories) |
| [segment_test.go](pkg/task/segment_test.go) | Segment filtering, duration calculations, activity tracking |
| [summary_test.go](pkg/task/summary_test.go) | Tagset grouping, weekly summaries, sorting |
| [persistence_test.go](pkg/task/persistence_test.go) | File I/O, YAML serialization, error handling |

**cmd/ow** has tests for utility functions and CLI summary:

| Test File | Description |
|-----------|-------------|
| [helpers_test.go](cmd/ow/helpers_test.go) | Duration formatting, time parsing, week calculations |
| [summary_test.go](cmd/ow/summary_test.go) | Summary generation, time filtering, output formatting |

Note: `cmd/ow` coverage is lower because `ui.go` contains TUI code (tview components, key handlers, form dialogs) that requires interactive testing. The testable utility functions in `helpers.go` and `summary.go` have full coverage.

Tests include:
- Table-driven tests for comprehensive edge case coverage
- Concurrency safety tests for thread-safe operations
- File permission and special character handling
- Large data set handling (100 tasks, 50 segments each)
- stdout capture for testing print functions

### Linting

```bash
golangci-lint run
```

### CI/CD

GitHub Actions runs two workflows:

**[go.yml](.github/workflows/go.yml)** - Build and test:
- Runs on `ubuntu-latest` with Go `1.24.x`
- Builds and tests on every push to main and PR

**[golangci.yml](.github/workflows/golangci.yml)** - Linting:
- Runs on `ubuntu-latest` and `macos-latest` matrix
- Uses golangci-lint v2.3
- Triggered on push to main and all PRs

## Code Organization

### Package: `pkg/task`

This package contains all core business logic and is designed to be thread-safe and reusable.

#### Key Files

**[types.go](pkg/task/types.go)**
- Data structure definitions (Watch, Task, Segment)
- Thread-safety via mutexes

**[task.go](pkg/task/task.go)**
- Task and segment CRUD operations
- File I/O operations
- Category management
- Sorting and filtering logic

**[segment.go](pkg/task/segment.go)**
- Segment time range filtering
- Duration calculations
- Activity tracking

**[summary.go](pkg/task/summary.go)**
- Weekly summary generation
- Tagset grouping
- Report data structures

**[interfaces.go](pkg/task/interfaces.go)**
- Interface definitions for task management
- `Manager` interface for Watch operations
- `TimeTracker` interface for Task operations
- `Persister` interface for file I/O
- Compile-time interface compliance checks

### Package: `cmd/ow`

The main application package containing the TUI and CLI interface.

**[main.go](cmd/ow/main.go)**
- CLI entry point
- Command-line flag parsing
- Mode dispatch (TUI vs summary)

**[ui.go](cmd/ow/ui.go)**
- `App` struct holding all TUI state
- Component initialization (`initTable`, `initDescriptionView`, etc.)
- Key handlers (`handleKeyEvent`, `handleRuneKey`)
- Form dialogs (new task, modify task, segment with note)
- Error dialog for save failures
- Task deletion with confirmation
- Segment details view

**[summary.go](cmd/ow/summary.go)**
- CLI summary generation (`generateSummary`)
- Weekly report formatting
- Task filtering and display

**[helpers.go](cmd/ow/helpers.go)**
- Duration formatting (`formatDuration`)
- Time parsing utilities (`parseTimeFlags`)
- Week calculation functions (`getMondayOfWeek`, `getLastMonday`, `getWeekStarts`)

## Key Concepts

### Thread Safety

All public methods in the `task` package are thread-safe:
- Uses `sync.RWMutex` for concurrent read/write protection
- Safe for use in goroutines
- The TUI uses a background goroutine to update ongoing segment durations

### Week Calculation

Weeks are defined as Monday-to-Monday:
- Week starts: Monday at 00:00:00
- Week ends: Following Monday at 00:00:00 (exclusive)
- Helpers in [helpers.go](cmd/ow/helpers.go:50-88) calculate week boundaries

### Segment Filtering

Time filtering for segments follows these rules:
- Only closed segments (non-zero Finish time) are included in reports
- Exclusive lower bound: `segment.Finish > start`
- Inclusive upper bound: `segment.Finish <= finish`
- This matches the "This Week" column logic in the TUI
- See [segment.go](pkg/task/segment.go:7-25) for implementation

### Tagset Summaries

Tasks are grouped by unique tag combinations:
- Tags are sorted alphabetically to create a consistent key
- Tasks with identical tag sets are grouped together
- Summaries show total duration for each tagset
- Useful for understanding time allocation across different types of work

## Development Guidelines

### Adding New Features

1. **Business logic** should go in `pkg/task/`
2. **TUI components and handlers** should go in `cmd/ow/ui.go`
3. **CLI summary logic** should go in `cmd/ow/summary.go`
4. **Time/formatting utilities** should go in `cmd/ow/helpers.go`
5. **Maintain thread safety** - use mutexes for shared data
6. **Follow existing patterns** for consistency

### UI Architecture

The TUI uses an `App` struct pattern:
- All shared state is held in the `App` struct
- UI components are initialized via `init*` methods
- Key handlers use a map-based dispatch (`handleRuneKey`)
- Forms use shared styling via `styleForm()`
- Errors are shown via modal dialogs, not panics

### Writing Tests

Tests for `pkg/task` should follow these patterns:
- Use table-driven tests for comprehensive coverage
- Use `t.Parallel()` for concurrent test execution
- Test both success and error cases
- Include concurrency tests for thread-safe operations
- Use `t.TempDir()` for file I/O tests

Test files are excluded from strict linting rules (see `.golangci.yml`):
- `gocyclo`, `gosmopolitan`, `gosec` relaxed for test readability

### Code Quality

The project uses golangci-lint with strict settings:
- All linters enabled by default
- Exceptions in `.golangci.yml`
- Deprecated packages (like `gopkg.in/yaml`) are forbidden
- Functions kept small to satisfy cyclomatic complexity limits
- Test files have relaxed linting for practical test writing

### Commit Conventions

Based on recent commit history:
- `feat:` for new features
- `fix:` for bug fixes
- `refactor:` for code refactoring
- Use descriptive commit messages
- Reference PR numbers when applicable

## Recent Changes

Based on git history:

1. **Summary formatting** - Updated summary output to be more useful
2. **Code optimization** - Light optimizations in duplicated code and minor tweaks
3. **Weekly statistics and categories** - Implemented weekly reporting and category system
4. **Package refactoring** - Moved items out into packages for better organization
5. **UI refactoring** - Separated TUI code into App struct pattern with `ui.go`, `summary.go`, and `helpers.go`
6. **Test suites** - Comprehensive test coverage for `pkg/task` (97.5%) and `cmd/ow` helpers/summary
7. **Removed unused code** - Deleted `config.go`, `doc.go`, `integration_test.go`, `yaml_test.go`

## File Locations

- **Tasks data:** `~/.ohgmas-tasks.yaml` (default)
- **Binary:** `./ow` (after build)
- **CLI source:** `./cmd/ow/main.go`
- **TUI source:** `./cmd/ow/ui.go`
- **Summary source:** `./cmd/ow/summary.go`
- **Business logic:** `./pkg/task/`
- **pkg/task tests:** `./pkg/task/*_test.go`
- **cmd/ow tests:** `./cmd/ow/*_test.go`
- **Lint config:** `./.golangci.yml`
- **Git ignore:** `./.gitignore`

## API Reference (for developers)

### Watch Methods

```go
// Task Management
AddTask(name, description string, tags []string, category string)
GetTasksByCategory(category string) []*Task
GetTasksSortedByActivity() []*Task
GetTasksSortedByActivityWithFilter(categoryFilter string) []*Task
GetTaskIndex(task *Task) int

// Persistence
SaveTasks() error
LoadTasks() error
SaveTasksToFile(filePath string) error
LoadTasksFromFile(filePath string) error

// Reporting
GetSummaryByTagset(start, finish *time.Time) []TagsetSummary
GetWeeklySummaryByTagset(weekStarts []time.Time) []WeeklySummary
GetWeeklySummaryByTagsetWithTasks(weekStarts []time.Time) []WeeklySummary
GetEarliestAndLatestSegmentTimes() (time.Time, time.Time)
```

### Task Methods

```go
// Segment Management
AddSegment(note string)
CloseSegment()
HasUnclosedSegment() bool
GetLastSegment() *Segment

// Duration Calculations
GetClosedSegmentsDuration() time.Duration
GetCurrentSegmentDuration() time.Duration
GetFilteredClosedSegmentsDuration(start, finish *time.Time) time.Duration
GetThisWeekDuration(weekStart time.Time) time.Duration

// Metadata
SetCategory(category string)
GetCategory() string
GetLastActivity() time.Time
IsActive() bool
HasSegmentsInRange(start, finish *time.Time) bool
```


# ohgmas-watch

A utility to aid in time management to track what is consuming the most time in a work day.

## Usage

### Interactive Mode

Run `./ow` to launch the TUI. Tasks are stored per fiscal quarter in `~/ohgmas/YYYY-FD.yaml` (fiscal year starts in October; e.g., `~/ohgmas/2026-F3.yaml` for Apr-Jun 2026). Use `--file /path/to/tasks.yaml` for a custom data file.

#### Key Bindings

| Key | Action |
|-----|--------|
| `t` | Create new task |
| `m` | Modify selected task |
| `d` | Delete selected task |
| `s` | Start new segment |
| `n` | Start new segment with note |
| `e` | End active segment |
| `c` / `w` / `b` | Set category to completed / work / backlog |
| `f` | Cycle category filter |
| `Enter` | View segment history |
| `Ctrl+C` | Exit |

### Summary Mode

```bash
./ow --summary              # weekly summaries by tagset (current quarter)
./ow --summary --all        # aggregate across all quarter files in ~/ohgmas/
./ow --summary --tasks      # include individual task breakdowns
./ow --summary --start 2024-01-01T00:00:00Z --finish 2024-12-31T23:59:59Z
```

`--all` only applies to `--summary` and is mutually exclusive with `--file`.

## Build

```bash
go build -o ow ./cmd/ow
```

## Test

```bash
go test ./...
```
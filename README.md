# kaizen

Local-first CLI tool for tracking daily habits.

## Current Features

- Create a habit: `kaizen new "Read daily"`
- Mark habits done: `kaizen done read gym`
- Mark a deliberate rest day that keeps the streak alive: `kaizen skip read`
- Remove a check-in: `kaizen undo read`
- Backfill any past day: `kaizen done read -d yesterday`, `-d -3`, `-d 2026-08-30`
- Initialize the data files: `kaizen init`

## Installation

### From source with make

Builds and copies `kaizen` to `~/.local/bin`:

```bash
make build
```

### From source with Go

```bash
go install github.com/amiraminb/kaizen@latest
```

## Quick Start

```bash
kaizen init
kaizen new "Read daily"
kaizen done read
```

## Data

Data lives in `~/Documents/.kaizen` by default. Set `KAIZEN_DATA_DIR` to an absolute
path to keep it somewhere else, such as a synced or version-controlled folder. If
`KAIZEN_DATA_DIR` is set but unusable, kaizen fails rather than falling back, so your
history never lands somewhere you would not look for it.

| File | Contents |
| --- | --- |
| `config.json` | `day_start_hour`, the cutoff that decides which day a check-in belongs to |
| `habits.json` | one record per habit: id, slug, name, schedule, start date, archived-at |
| `entries.json` | one record per habit per day: status (`done` or `skipped`) and an optional note |

Writes are atomic and durable. Each save marshals to a temp file in the same
directory, fsyncs it, renames it over the target, then fsyncs the directory.

## Model

A day resolves to exactly one of five states for a given habit:

```
  ✓  done        an entry with status done
  ~  skipped     an entry with status skipped, the streak survives
  ◦  pending     scheduled, no entry yet, and the day is today
  ✗  miss        scheduled, no entry, and the day is over
  ·  n/a         before the start date or after the habit was archived
```

`day_start_hour` defaults to 4, so a check-in typed at 01:30 counts for the previous
day rather than the new one. Dates use the machine's local timezone.

Habits are archived, never deleted, so retiring one keeps its history and stops it
generating misses. A habit is addressed by its slug, and any unambiguous prefix
works: `kaizen done med` resolves `meditate`.

## Roadmap

- `kaizen today` and `kaizen streak`
- interactive checklist on bare `kaizen`, plus `edit` and `archive`
- `kaizen log` and `kaizen report`

# kaizen

Local-first CLI tool for tracking daily habits.

## Current Features

- Check off today interactively: `kaizen` with no arguments
- Create a habit: `kaizen new "Read daily"`, or `kaizen new` to be prompted
- Rename a habit or change its slug: `kaizen edit`, or `kaizen edit read --slug reading`
- Mark habits done: `kaizen done read gym`
- Mark a deliberate rest day that keeps the streak alive: `kaizen skip read`
- Fix a day you got wrong: `kaizen -d -2` opens that day, and `c` clears a check-in
- Backfill any past day: `kaizen done read -d yesterday`, `-d -3`, `-d 2026-08-30`
- Per-habit completion over a range: `kaizen report`, `kaizen report 30d`, `kaizen report lastmonth`
- Every check-in with its notes: `kaizen report entries`, `kaizen report entries read 7d`
- Initialize the data files: `kaizen init`

```
$ kaizen
Today

> [✓]  read      Read daily    changed
  [~]  gym       Gym session   changed
  [ ]  meditate  Meditate 10m  streak 4

(space cycle, d done, s skip, c clear, enter save, esc cancel)

$ kaizen report 20d
2026-08-17 .. 2026-09-05

habit  days                  done  rate  cur  best
read   ✗✓✓✓✓✗✓✓✓✓✓✓✓~✓✓✓✓✓✓    17   89%   13    13
gym    ✗✗✓✗✓✗✓✗✓✗✓✗✓✗✓✗✓~✓◦     9   50%    2     2

$ kaizen report entries read 7d
date        habit  status   note
2026-09-05  read   done
2026-09-04  read   done
2026-08-30  read   skipped  flu
```

### Ranges

`report` and `report entries` accept `today`, `yesterday`, `week`, `lastweek`,
`month`, `lastmonth`, `year`, `lastyear`, `Nd` for the last N days, a single
`YYYY-MM-DD`, or a `YYYY-MM-DD..YYYY-MM-DD` span. The default is the current month.
Open-ended ranges stop at today rather than running into the future.

A range never ends in the future. An end date past today is clamped to today, and a
range that starts in the future is rejected.

Use `Nd` for positional ranges: a leading dash is consumed by the flag parser before
the command sees it. The two relative forms are not the same window: `7d` is the last
7 days, while `-7` means "7 days ago through today", which is 8 days. `-N` still works
for the `--date` flag, where it is a flag value, as in `kaizen done read -d -3`.

A habit whose start date has not arrived shows as `·` and cannot be checked off until
it starts.

Output drops all colour when it is not going to a terminal, and when `NO_COLOR` is
set, so `kaizen | grep` and `kaizen report | awk` stay clean. The glyphs carry
every state on their own.

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

The stats engine already honours an `archived_at` timestamp, which stops a habit
generating misses while keeping its history, but no command sets it yet. A habit is
addressed by its slug, and any unambiguous prefix works: `kaizen done med` resolves
`meditate`.

Streaks are always measured over a habit's whole history, never over the window on
screen, so a 30-day report cannot cap a 40-day run at 30. Streak numbers appear in
`kaizen report` as `cur` and `best`, and next to each habit in the checklist.

Bare `kaizen` opens today by default and any past day with `-d`, which is the only way
to turn a day you recorded by mistake back into a miss. It needs a terminal: piped or
scripted it prints that day's status instead of failing, so `kaizen | less` and
`kaizen -d -2 > status.txt` both do something useful.

## Roadmap

- retiring a habit (the `archived_at` field exists but nothing sets it yet)
- measured habits, such as pages read or minutes meditated

# kaizen

Local-first CLI tool for tracking daily habits.

## Current Features

- Check off today interactively: `kaizen` with no arguments
- Create a habit: `kaizen new "Read daily"`, or `kaizen new` to be prompted
- Rename a habit or change its slug: `kaizen edit`, or `kaizen edit read --slug reading`
- Mark habits done: `kaizen done read gym`
- Mark a deliberate rest day that keeps the streak alive: `kaizen skip read`
- Mark a missed day that breaks the streak: `kaizen miss read`, or `kaizen miss read -d yesterday`
- Fix a day you got wrong: `kaizen -d -2` opens that day, and `c` clears a check-in
- Backfill any past day: `kaizen done read -d yesterday`, `-d -3`, `-d 2026-08-30`
- Read a habit note in the terminal: `kaizen notes read`
- Edit one persistent note document for a habit: `kaizen notes read -e`
- Pause before giving in to an urge and reflect: `kaizen urge`; save the answers with
  `kaizen notes <habit> -e`
- Per-habit completion over a range: `kaizen report`, `kaizen report 30d`, `kaizen report lastmonth`
- Focus on one habit: `kaizen report read month`, `kaizen report read 2026-10-31`,
  `kaizen report read year`
- Initialize the data files: `kaizen init`

```
$ kaizen
Today

> [✓]  read      Read daily    changed
  [~]  gym       Gym session   changed
  [ ]  meditate  Meditate 10m  streak 4

(space cycle, d done, s skip, m miss, c clear, enter save, esc cancel)

$ kaizen report 20d
2026-08-17 .. 2026-09-05
Completion = done / (done + missed). Skipped and pending days are excluded.
Streaks as of 2026-09-05

read
  completion: 89% (17/19 days)  17 done  2 missed  1 skipped
  streak today: 13  best lifetime: 13
  2026-08-17 .. 2026-09-05
  Aug 17–22 (Mon–Sat)  ✗✓✓✓✓✗
  Aug 23–29 (Sun–Sat)  ✓✓✓✓✓✓✓
  Aug 30–Sep 05 (Sun–Sat)  ~✓✓✓✓✓✓

gym
  completion: 50% (9/18 days)  9 done  9 missed  1 skipped
  streak today: 2  best lifetime: 2
  2026-08-17 .. 2026-09-05
  Aug 17–22 (Mon–Sat)  ✗✗✓✗✓✗
  Aug 23–29 (Sun–Sat)  ✓✗✓✗✓✗✓
  Aug 30–Sep 05 (Sun–Sat)  ✗✓✗✓~✓◦
```

### Ranges

Use `kaizen report [range]` for all active habits, or `kaizen report <habit> <range>`
for one habit. Habits match by slug or an unambiguous prefix. A single argument is
always a date or range; with no arguments, the report shows the current month.
Add `--all` to include retired habits, including when selecting one habit.

`report` accepts `today`, `yesterday`, `week`, `lastweek`,
`month`, `lastmonth`, `year`, `lastyear`, `Nd` for the last N days, a single
`YYYY-MM-DD`, or a `YYYY-MM-DD..YYYY-MM-DD` span. The default is the current month.
Open-ended ranges stop at today rather than running into the future.
Each habit's timeline starts at its start date or the beginning of the requested
range, whichever is later; days before the habit started are omitted.

Reports show each habit's name on its own line, followed by an indented row with
its completion rate and done, missed, and skipped totals before its timeline.
Completion is `done / (done + missed)`, excluding skipped and
pending days. For example, `completion: 80% (8/10 days)` means 8 done days out of
8 done + 2 missed days. The overall percentage covers the selected range; each
month's percentage in a year report covers only that month's days in the range.

Ranges up to 31 days show one row per Sunday–Saturday calendar week, labeled with
the start/end dates and weekdays. Full weeks have seven marks; partial weeks at
the habit's start or the range boundaries have fewer. `year`, `lastyear`, and
longer ranges show one row per month with its completion rate, including the year
and the completed-day fraction. When selecting a habit, each month also expands
into its weekly day-by-day rows, with weeks clipped to that month and the selected
range. Each habit has its own displayed date range; months before it started are
omitted.

`week` runs from Sunday through today; `lastweek` covers the previous Sunday
through Saturday.

Output wraps to the terminal width, with an 80-column fallback for redirected
output. A rate of `—` means there are no done or missed days to score: pending and
skipped days are excluded. Habits that have not started say `starts YYYY-MM-DD`.

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
| `entries.json` | one record per habit per day: status (`done`, `skipped`, or `missed`) and an optional note |
| `notes.json` | one editable note document per habit |

Writes are atomic and durable. Each save marshals to a temp file in the same
directory, fsyncs it, renames it over the target, then fsyncs the directory.

## Model

A day resolves to exactly one of five states for a given habit:

```
  ✓  done        an entry with status done
  ~  skipped     an entry with status skipped, the streak survives
  ◦  pending     scheduled, no entry yet, and the day is today
  ✗  miss        explicitly marked missed, or scheduled with no entry after the day is over
  ·  n/a         before the start date or after the habit was archived
```

`day_start_hour` defaults to 4, so a check-in typed at 01:30 counts for the previous
day rather than the new one. Dates use the machine's local timezone.

`kaizen miss read` records a miss immediately, even today. Like `done` and `skip`, it
accepts multiple habits, `-d` for a past date, and `-n` for a note. Misses count against
completion and break streaks; use `done` or `skip` to correct one, or `c` in the
checklist to clear it. The checklist also supports `m` to mark a miss.

The stats engine already honours an `archived_at` timestamp, which stops a habit
generating misses while keeping its history, but no command sets it yet. A habit is
addressed by its slug, and any unambiguous prefix works: `kaizen done med` resolves
`meditate`.

Streaks are always measured over a habit's whole history, never over the window on
screen, so a 30-day report cannot cap a 40-day run at 30. Streak numbers appear in
`kaizen report` as `streak today` and `best lifetime`, and next to each habit in the
checklist. The report heading states the streaks' as-of date, even when the selected
range is historical; the counts and rates cover only the selected range.

Bare `kaizen` opens today by default and any past day with `-d`, and lets you clear a
check-in recorded by mistake. It needs a terminal: piped or
scripted it prints that day's status instead of failing, so `kaizen | less` and
`kaizen -d -2 > status.txt` both do something useful.

## Roadmap

- retiring a habit (the `archived_at` field exists but nothing sets it yet)
- measured habits, such as pages read or minutes meditated

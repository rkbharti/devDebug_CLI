# LogSensei

<p align="center">
  <img src="docs/logsensei.gif" alt="LogSensei demo" />
</p>

A fast, smart CLI tool that analyzes log files and surfaces real errors instantly.
Supports plain text and JSON logs, custom patterns, real-time watching, and multi-format export.

## Install

Download the latest binary from [Releases](https://github.com/rkbharti/LogSensei_CLI/releases):

| Platform | Binary                        |
| -------- | ----------------------------- |
| Windows  | `logsensei-windows-amd64.exe` |
| Linux    | `logsensei-linux-amd64`       |
| macOS    | `logsensei-darwin-amd64`      |

Or build from source (requires Go 1.22+):

    git clone https://github.com/rkbharti/LogSensei_CLI.git
    cd LogSensei_CLI
    go build -o logsensei .

## Commands

    logsensei analyze <file|folder>   Scan log file or folder for errors
    logsensei compare <old> <new>     Diff two log files
    logsensei init                    Generate starter logsensei.yaml
    logsensei version                 Print version info

## Analyze Flags

    --type    Filter by error type: panic, timeout, error
    --format  Export report: json or md
    --follow  Watch file in real-time (tail mode)
    --quiet   No output, only exit code (CI use)
    --since   Show errors after time  e.g. 2026-04-19T10:00:00
    --until   Show errors before time e.g. 2026-04-19T18:00:00

## Custom Patterns (logsensei.yaml)

Run `logsensei init` to generate a starter config, then edit it:

    patterns:
      - name: "Auth Failure"
        keyword: "unauthorized"

      - name: "5xx HTTP Error"
        regex: "HTTP [5][0-9]{2}"

      - name: "Retry Exhausted"
        regex: "(?i)failed after [0-9]+ retr"

Place `logsensei.yaml` in the same directory where you run the command.

## Examples

    # analyze a log file
    logsensei analyze app.log

    # filter only panic errors
    logsensei analyze app.log --type panic

    # export as JSON report
    logsensei analyze app.log --format json

    # watch a live log file
    logsensei analyze app.log --follow

    # compare two log files
    logsensei compare old.log new.log

    # use in CI pipelines (exit 1 if errors found)
    logsensei analyze app.log --quiet

## CI Usage

    - name: Check logs for errors
      run: logsensei analyze app.log --quiet

Exit code `0` = no errors. Exit code `1` = errors found.

## Tech Stack

- Language : Go 1.22
- CLI : cobra
- Styling : charmbracelet/lipgloss
- Testing : Go standard testing — 102 tests
- CI/CD : GitHub Actions

## License

MIT

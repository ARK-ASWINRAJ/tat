# Terminal Activity Tracker (TAT)

TAT is a local, privacy-first developer tool that automatically records terminal activity so it’s searchable later. It captures each command with timestamp, working directory, exit code, and duration, and stores it in a local SQLite database. Quickly recall “that one command I used to fix Wi‑Fi” or audit what ran in a project directory last week.

- Runs locally, keeps data in ~/.tat/tat.db
- Works with Zsh and Bash
- Simple CLI to initialize, enable/disable, and search
- Extensible architecture for future features (full output capture, web UI, AI/MCP)

***

## Table of Contents

- What TAT Does
- Installation
- Quick Start (Normal Mode)
- Development Mode (No Install)
- Usage
  - Status
  - Enable/Disable
  - Search
- Privacy Controls
- Uninstall
- Troubleshooting
- How It Works (Under the Hood)
- Roadmap
- Project Structure
- License

***

## What TAT Does

- Records for each command:
  - Command line (cmd)
  - Working directory (cwd)
  - Start/end timestamps
  - Exit code
  - Duration in milliseconds
- Stores everything locally in SQLite
- Lets you search previous commands by substring:
  - Example: tat search "wifi" or tat search "docker build"

Note: In this MVP, search is over the typed command string (cmd). Full stdout/stderr capture and richer search are planned.

***

## Installation

Option A: Build locally (recommended)

- git clone <your-repo-url> && cd <repo>
- go build -o ~/.local/bin/tat ./cmd/tat
- Add to PATH if needed:
  - Zsh (macOS):
    - echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
    - source ~/.zshrc
  - Bash (Linux):
    - echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
    - source ~/.bashrc

Option B: Temporary dev alias (current terminal only)

- From the repo root:
  - alias tat='go run '"$PWD"'/cmd/tat'
- Note: Use the built binary for reliability; this alias is for quick local testing.

***

## Quick Start (Normal Mode)

1) Initialize and enable
- tat init
- tat status
  - If it shows Enabled: false, run tat enable

2) Install shell hooks
- tat install-shell
  - Adds source lines to ~/.bashrc and/or ~/.zshrc and copies hook scripts to ~/.tat/

3) Restart the terminal (or source rc)
- Zsh: source ~/.zshrc
- Bash: source ~/.bashrc

4) Verify the hook loaded
- Zsh: echo $TAT_ZSH_SOURCED → should print 1
- Bash: echo $TAT_BASH_SOURCED → should print 1

5) Use the terminal normally
- Run some commands:
  - ls
  - pwd
  - echo "wifi test"
  - false

6) Search your history
- tat search "wifi"
- tat search echo

Expected output format:
[<cwd>] <command> (exit:<code or ->)

***

## Development Mode (No Install)

Use this if you want to test without modifying shell configs or PATH. You’ll manually simulate the shell hook events.

1) Initialize
- go run ./cmd/tat init
- go run ./cmd/tat enable
- go run ./cmd/tat status

2) Simulate a command run:
- Preexec (command start):
  - printf '{"event":"preexec","cmd":"echo wifi test","cwd":"%s","ts":"%s"}\n' "$PWD" "$(date -Iseconds)" | go run ./cmd/tat record
- Postexec (command end):
  - printf '{"event":"postexec","cmd":"echo wifi test","cwd":"%s","ts":"%s","exit":%d,"duration_ms":%d}\n' "$PWD" "$(date -Iseconds)" 0 120 | go run ./cmd/tat record

3) Search
- go run ./cmd/tat search wifi

Tip: Ensure you use straight quotes " and ' in JSON (no “smart quotes”).

***

## Usage

Status
- tat status
  - Shows whether recording is enabled and the database path.

Enable/Disable
- tat enable
- tat disable

Search
- tat search "<query>"
  - Returns up to 50 matches by substring on the command text.
  - Example: tat search "git commit" or tat search wifi

***

## Privacy Controls

- Temporarily pause recording in the current shell session:
  - export TAT_DISABLE=1
  - To resume: unset TAT_DISABLE (or open a new terminal)
- Disable globally:
  - tat disable
  - Re-enable: tat enable
- Exclude directories (advanced):
  - Edit ~/.tat/config.yaml and add paths under exclude_dirs
  - Restart the shell or open a new terminal

***

## Uninstall

- Remove hook source lines from shell configs:
  - Bash: edit ~/.bashrc and remove: source ~/.tat/tat.bash
  - Zsh: edit ~/.zshrc and remove: source ~/.tat/tat.zsh
- Remove TAT files:
  - rm -rf ~/.tat
- Remove the binary (if installed):
  - rm ~/.local/bin/tat

***

## Troubleshooting

“tat: command not found”
- Ensure ~/.local/bin is in PATH and restart your terminal (or source rc).
- Echo PATH to verify: echo $PATH

Hook not loaded (TAT_ZSH_SOURCED/TAT_BASH_SOURCED is empty)
- Confirm the source lines exist:
  - grep "source ~/.tat/tat.zsh" ~/.zshrc
  - grep "source ~/.tat/tat.bash" ~/.bashrc
- Ensure the hook files exist:
  - ls ~/.tat/tat.zsh ~/.tat/tat.bash
- Re-run:
  - tat install-shell
- Re-source:
  - source ~/.zshrc or source ~/.bashrc

Search returns nothing
- Confirm Enabled: tat status (enable if needed)
- Ensure you ran commands after the hook loaded (open a new terminal)
- Try a common query: tat search echo
- Check DB exists: ls ~/.tat/tat.db

JSON errors during dev mode testing
- Use straight quotes in JSON. Curly “smart quotes” cause decoding errors.
- Prefer printf examples given above.

macOS PATH quirks (Zsh)
- Add to ~/.zprofile if Terminal sessions don’t pick up PATH:
  - echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zprofile && source ~/.zprofile

***

## How It Works (Under the Hood)

- Shell hooks send a small JSON event before and after each command:
  - preexec: about to run a specific command
  - postexec: command finished, includes exit code and duration
- tat record reads one event from stdin and writes it to the local SQLite DB at ~/.tat/tat.db
- tat search queries the commands table using a substring match on cmd and prints results

MVP Data Model
- sessions: tracks a day’s active session per machine/shell metadata
- commands: one row per command run, with timestamps, cwd, exit code, duration
- outputs: reserved for future stdout/stderr capture

***

## Roadmap

- Phase 1.1:
  - Optional interactive “tat start” subshell and wire stdout/stderr chunks to DB
  - Better correlation using shell PID
  - Replace LIKE with SQLite FTS5 for more powerful search
- Phase 2:
  - Local HTTP API and Next.js dashboard for browsing and search
- Phase 3:
  - AI/MCP integration for natural language search, suggestions, and pattern recognition

***

## Project Structure

- cmd/tat
  - Main CLI entrypoint and subcommands (init, enable, disable, status, record, search)
- internal/config
  - Config loading and defaults (reads ~/.tat/config.yaml)
- internal/storage
  - Models and SQLite open/migrate via GORM
- internal/recorder
  - Event types and (future) async ingester
- scripts
  - tat.bash and tat.zsh shell hook scripts (installed to ~/.tat/)
- docs
  - Additional documentation (architecture, privacy, etc.)

***

## License

	•	MIT © 2025 ARK-ASWINRAJ. See LICENSE for details.

***


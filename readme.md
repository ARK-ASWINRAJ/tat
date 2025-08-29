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
  - Version
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
  - Example:
    ```bash
    tat search "wifi"
    tat search "docker build"
    ```
- Note: In this MVP, search is over the typed command string (cmd). Full stdout/stderr capture and richer search are planned.

***

## Installation

### Option A: One‑liner install (recommended)

- This installs the latest binary to ~/.local/bin, adds PATH if needed, initializes TAT, and installs shell hooks once.
- Works on macOS (Intel/Apple Silicon) and Linux x86_64/arm64.

```bash
# macOS/Linux
curl -fsSL https://raw.githubusercontent.com/ARK-ASWINRAJ/tat/main/scripts/install.sh | bash
```

What the installer does:
	•	Detects OS/arch and downloads the latest release asset (for example, tat-v0.1.0-darwin-arm64). 
	•	Installs to ~/.local/bin/tat and makes it executable. 
	•	Ensures PATH includes ~/.local/bin by updating ~/.bashrc and/or ~/.zshrc, creating the file if it’s missing. 
	•	Runs tat init; if disabled, runs tat enable; installs shell hooks once. 
	•	Prints a reminder to source the shell rc or open a new terminal. 

### Option B: Manual download from Releases

- Download the appropriate prebuilt binary from the latest GitHub Release and place it in PATH.

```bash
# Linux (x86_64)
curl -L -o ~/.local/bin/tat https://github.com/ARK-ASWINRAJ/tat/releases/latest/download/tat-linux-amd64 && chmod +x ~/.local/bin/tat

# macOS Intel
curl -L -o ~/.local/bin/tat https://github.com/ARK-ASWINRAJ/tat/releases/latest/download/tat-darwin-amd64 && chmod +x ~/.local/bin/tat

# macOS Apple Silicon
curl -L -o ~/.local/bin/tat https://github.com/ARK-ASWINRAJ/tat/releases/latest/download/tat-darwin-arm64 && chmod +x ~/.local/bin/tat
```

```bash
# Ensure PATH includes ~/.local/bin (run once)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc
# or for Zsh:
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc
```

### Option C: Build locally

```bash
git clone https://github.com/ARK-ASWINRAJ/tat.git && cd tat
go build -o ~/.local/bin/tat ./cmd/tat

# Add to PATH if needed (Zsh macOS)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc

# Add to PATH if needed (Bash Linux)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc
```

### Option D: Temporary dev alias (current terminal only)

```bash
# From the repo root
alias tat='go run '"$PWD"'/cmd/tat'
# Note: Use the built binary for reliability; this alias is only for quick local testing.
```

***

## Quick Start (Normal Mode)

```bash
# 1) Initialize and enable
tat init
tat status   # if "Enabled: false" then:
tat enable
```

```bash
# 2) Shell hooks installation
# No separate step needed. `tat init` detects missing hooks and prompts to install them automatically once.
# After installation, hooks load on every new shell session via your rc file.

# Reload the shell to activate in the current session:
# Zsh:
source ~/.zshrc
# Bash:
source ~/.bashrc
```

```bash
# 3) Verify hooks loaded
# Zsh:
echo $TAT_ZSH_SOURCED   # -> 1
# Bash:
echo $TAT_BASH_SOURCED  # -> 1
```

```bash
# 4) Use normally and search
ls
echo "wifi test"
false

tat search "wifi"
```

```bash
# Note for development mode:
# Skip installing hooks. Use the manual event simulation instead to test recording:
printf '{"event":"preexec","cmd":"echo dev","cwd":"%s","ts":"%s"}\n' "$PWD" "$(date -Iseconds)" | tat record
printf '{"event":"postexec","cmd":"echo dev","cwd":"%s","ts":"%s","exit":%d,"duration_ms":%d}\n' "$PWD" "$(date -Iseconds)" 0 100 | tat record
# This avoids changing shell configs during development; rely on `tat search` to verify data is recorded.
```

***

## Development Mode (No Install)

Use this if you want to test without modifying shell configs or PATH. You’ll manually simulate the shell hook events.

```bash
# 1) Initialize
go run ./cmd/tat init
go run ./cmd/tat enable
go run ./cmd/tat status
```

```bash
# 2) Simulate a command run:
# Preexec (command start)
printf '{"event":"preexec","cmd":"echo wifi test","cwd":"%s","ts":"%s"}\n' "$PWD" "$(date -Iseconds)" | go run ./cmd/tat record

# Postexec (command end)
printf '{"event":"postexec","cmd":"echo wifi test","cwd":"%s","ts":"%s","exit":%d,"duration_ms":%d}\n' "$PWD" "$(date -Iseconds)" 0 120 | go run ./cmd/tat record
```

```bash
# 3) Search
go run ./cmd/tat search wifi

# Tip: Ensure you use straight quotes " and ' in JSON (no “smart quotes”).
```

***

## Usage

```bash
# Status
tat status
# Shows whether recording is enabled and the database path.
```

```bash
# Enable/Disable
tat enable
tat disable
```

```bash
# Search
tat search "<query>"
# Returns up to 50 matches by substring on the command text.
# Example:
tat search "git commit"
tat search wifi
```

```bash
# Version
tat version
# Prints the binary version (e.g., v0.1.0). The binary is built with: -ldflags="-X 'main.version=$VERSION'"
```

***

## Privacy Controls

```bash
# Temporarily pause recording in the current shell session:
export TAT_DISABLE=1

# To resume:
unset TAT_DISABLE
# or open a new terminal
```

```bash
# Disable globally:
tat disable

# Re-enable:
tat enable
```

```bash
# Exclude directories (advanced):
# Edit ~/.tat/config.yaml and add paths under exclude_dirs
# Restart the shell or open a new terminal
```

***

## Uninstall

```bash
# Remove hook source lines from shell configs:
# Bash:
#   edit ~/.bashrc and remove: source ~/.tat/tat.bash
# Zsh:
#   edit ~/.zshrc and remove: source ~/.tat/tat.zsh
```

```bash
# Remove TAT files (data):
rm -rf ~/.tat   # Only if you want to remove all recorded data
```

```bash
# Remove the binary (if installed):
rm ~/.local/bin/tat
```

***

## Troubleshooting

```bash
# “tat: command not found”
# Ensure ~/.local/bin is in PATH and restart your terminal (or source rc).
echo $PATH
```

```bash
# Hook not loaded (TAT_ZSH_SOURCED/TAT_BASH_SOURCED is empty)
# Confirm the source lines exist:
grep "source ~/.tat/tat.zsh" ~/.zshrc
grep "source ~/.tat/tat.bash" ~/.bashrc

# Ensure the hook files exist:
ls ~/.tat/tat.zsh ~/.tat/tat.bash

# Re-run:
tat init   # or `tat install-shell` if you want to reinstall hooks explicitly

# Re-source:
source ~/.zshrc   # or
source ~/.bashrc
```

```bash
# Search returns nothing
tat status           # confirm Enabled (enable if needed)
# Ensure you ran commands after the hook loaded (open a new terminal)
tat search echo
ls ~/.tat/tat.db
```

```bash
# JSON errors during dev mode testing
# Use straight quotes in JSON. Curly “smart quotes” cause decoding errors.
# Prefer the printf examples provided above.
```

```bash
# macOS PATH quirks (Zsh)
# If Terminal sessions don’t pick up PATH:
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zprofile && source ~/.zprofile
```

***

## How It Works (Under the Hood)

- Shell hooks send a small JSON event before and after each command:
  - preexec: about to run a specific command
  - postexec: command finished, includes exit code and duration
- `tat record` reads one event from stdin and writes it to the local SQLite DB at `~/.tat/tat.db`.
- `tat search` queries the commands table using a substring match on `cmd` and prints results.

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

```
cmd/tat
  # Main CLI entrypoint and subcommands (init, enable, disable, status, record, search)
internal/config
  # Config loading and defaults (reads ~/.tat/config.yaml)
internal/storage
  # Models and SQLite open/migrate via GORM
internal/recorder
  # Event types and (future) async ingester
scripts
  # tat.bash and tat.zsh shell hook scripts (installed to ~/.tat/)
docs
  # Additional documentation (architecture, privacy, etc.)
```

***

## License

```text
MIT © 2025 ARK-ASWINRAJ. See LICENSE for details.
```

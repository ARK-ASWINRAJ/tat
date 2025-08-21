package recorder

import "time"

type CommandStart struct {
	SessionID string
	Cmd       string
	CWD       string
	StartedAt time.Time
	Shell     string
	Hostname  string
}

type CommandEnd struct {
	CommandID  string
	ExitCode   int
	EndedAt    time.Time
	DurationMs int64
}

type CommandOutput struct {
	CommandID string
	Stream    string // stdout|stderr
	Chunk     string
	At        time.Time
}

package storage

import "time"

type Session struct {
	ID        string `gorm:"primaryKey"`
	StartedAt time.Time
	EndedAt   *time.Time
	CWD       string
	Shell     string
	Hostname  string
	Status    string // active/completed/aborted
	Tags      string // JSON, optional
}

type Command struct {
	ID         string `gorm:"primaryKey"`
	SessionID  string `gorm:"index"`
	LineNo     int
	Cmd        string
	ExitCode   *int
	DurationMs *int64
	CWD        string
	StartedAt  time.Time
	EndedAt    *time.Time
}

type Output struct {
	ID        string `gorm:"primaryKey"`
	CommandID string `gorm:"index"`
	Stream    string // stdout|stderr
	Chunk     string
	At        time.Time
}

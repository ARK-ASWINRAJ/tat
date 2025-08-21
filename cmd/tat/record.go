package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ARK-ASWINRAJ/tat/internal/config"
	"github.com/ARK-ASWINRAJ/tat/internal/storage"
	"github.com/spf13/cobra"
)

func recordCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "record",
		Short: "Internal: read a single JSON event from stdin and persist",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if !cfg.Enabled {
				return nil
			}
			st, err := storage.Open(context.Background(), cfg.DatabasePath)
			if err != nil {
				return err
			}
			// For MVP: create or reuse a single default session per day
			sessID, err := ensureSession(st)
			if err != nil {
				return err
			}

			type payload struct {
				Event      string `json:"event"` // preexec|postexec
				Cmd        string `json:"cmd"`
				CWD        string `json:"cwd"`
				TS         string `json:"ts"`
				Exit       int    `json:"exit"`
				DurationMs int64  `json:"duration_ms"`
			}

			rd := bufio.NewReader(os.Stdin)
			line, _, err := rd.ReadLine()
			if err != nil {
				return err
			}
			var p payload
			if err := json.Unmarshal(line, &p); err != nil {
				return err
			}

			t, _ := time.Parse(time.RFC3339, p.TS)

			switch p.Event {
			case "preexec":
				// Insert command start (no CommandID handling yet)
				cmdRow := storage.Command{
					ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
					SessionID: sessID,
					Cmd:       p.Cmd,
					CWD:       p.CWD,
					StartedAt: t,
				}
				return st.DB.Create(&cmdRow).Error
			case "postexec":
				// Naive: update last command by started_at desc
				var last storage.Command
				if err := st.DB.
					Where("session_id = ?", sessID).
					Order("started_at DESC").Limit(1).
					Find(&last).Error; err != nil {
					return err
				}
				return st.DB.Model(&last).Updates(map[string]any{
					"exit_code":   p.Exit,
					"ended_at":    t,
					"duration_ms": p.DurationMs,
				}).Error
			default:
				return nil
			}
		},
	}
}

func ensureSession(st *storage.Store) (string, error) {
	// For MVP: one active session per day
	var s storage.Session
	today := time.Now().Format("2006-01-02")
	if err := st.DB.Where("status = ? AND date(started_at) = date(?)", "active", today).
		Order("started_at desc").First(&s).Error; err == nil {
		return s.ID, nil
	}
	s = storage.Session{
		ID:        fmt.Sprintf("sess-%d", time.Now().UnixNano()),
		StartedAt: time.Now(),
		CWD:       "",
		Shell:     os.Getenv("SHELL"),
		Hostname:  func() string { h, _ := os.Hostname(); return h }(),
		Status:    "active",
	}
	if err := st.DB.Create(&s).Error; err != nil {
		return "", err
	}
	return s.ID, nil
}

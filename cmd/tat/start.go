package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/ARK-ASWINRAJ/tat/internal/config"
	"github.com/ARK-ASWINRAJ/tat/internal/storage"
	"github.com/creack/pty"
	"github.com/spf13/cobra"
)

func startCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start an interactive shell session with output capture (MVP)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			st, err := storage.Open(context.Background(), cfg.DatabasePath)
			if err != nil {
				return err
			}

			shell := os.Getenv("SHELL")
			if shell == "" {
				shell = "/bin/bash"
			}

			c := exec.Command(shell, "-l")
			ptmx, err := pty.Start(c)
			if err != nil {
				return err
			}
			defer func() { _ = ptmx.Close() }()

			// naive: just mirror to stdout/stderr
			go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
			_, _ = io.Copy(os.Stdout, ptmx)

			// on exit, mark a session completed if needed
			_ = st // future: write outputs to DB
			_ = time.Now()
			return nil
		},
	}
}

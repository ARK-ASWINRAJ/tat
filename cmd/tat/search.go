package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/ARK-ASWINRAJ/tat/internal/config"
	"github.com/ARK-ASWINRAJ/tat/internal/storage"
	"github.com/spf13/cobra"
)

func searchCmd() *cobra.Command {
	var withOutput bool
	var limit int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search past commands (simple LIKE MVP)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			q := "%" + args[0] + "%"
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			st, err := storage.Open(context.Background(), cfg.DatabasePath)
			if err != nil {
				return err
			}

			type row struct {
				Cmd      string
				CWD      string
				ExitCode *int
				Stdout   string
				Stderr   string
			}
			var rows []row
			if err := st.DB.
				Table("commands").
				Select("cmd, cwd, exit_code, stdout, stderr").
				Where("cmd LIKE ?", q).
				Order("started_at DESC").
				Limit(limit).
				Scan(&rows).Error; err != nil {
				return err
			}

			for _, r := range rows {
				ec := "-"
				if r.ExitCode != nil {
					ec = fmt.Sprint(*r.ExitCode)
				}
				fmt.Printf("[%s] %s (exit:%s)\n", r.CWD, r.Cmd, ec)

				if withOutput {
					so := previewOneLine(r.Stdout, 200)
					se := previewOneLine(r.Stderr, 200)
					if so != "" {
						fmt.Printf("  stdout: %s\n", so)
					}
					if se != "" {
						fmt.Printf("  stderr: %s\n", se)
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&withOutput, "with-output", false, "show a short preview of stdout/stderr")
	cmd.Flags().IntVar(&limit, "limit", 50, "max results to return")

	return cmd
}

// previewOneLine trims to max chars and flattens newlines for single-line display
func previewOneLine(s string, max int) string {
	if s == "" {
		return ""
	}
	// Normalize newlines and make them visible
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", " ⏎ ")
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

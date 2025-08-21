package main

import (
	"context"
	"fmt"

	"github.com/ARK-ASWINRAJ/tat/internal/config"
	"github.com/ARK-ASWINRAJ/tat/internal/storage"
	"github.com/spf13/cobra"
)

func searchCmd() *cobra.Command {
	return &cobra.Command{
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
			}
			var rows []row
			if err := st.DB.
				Table("commands").
				Select("cmd, cwd, exit_code").
				Where("cmd LIKE ?", q).
				Order("started_at DESC").Limit(50).
				Scan(&rows).Error; err != nil {
				return err
			}
			for _, r := range rows {
				ec := "-"
				if r.ExitCode != nil {
					ec = fmt.Sprint(*r.ExitCode)
				}
				fmt.Printf("[%s] %s (exit:%s)\n", r.CWD, r.Cmd, ec)
			}
			return nil
		},
	}
}

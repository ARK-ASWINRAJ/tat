package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ARK-ASWINRAJ/tat/internal/config"
	"github.com/ARK-ASWINRAJ/tat/internal/storage"
)

func main() {
	root := &cobra.Command{Use: "tat", Short: "Terminal Activity Tracker"}

	root.AddCommand(initCmd(), enableCmd(), disableCmd(), statusCmd(), startCmd(), searchCmd(), recordCmd())
	_ = root.Execute()
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize TAT (config, database)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			// Ensure dir exists
			dir := filepath.Dir(cfg.DatabasePath)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			// Open DB to run migrations
			_, err = storage.Open(context.Background(), cfg.DatabasePath)
			if err != nil {
				return err
			}
			fmt.Println("Initialized:", cfg.DatabasePath)
			return nil
		},
	}
}

func enableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Enable tracking",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setEnabled(true)
		},
	}
}

func disableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable tracking",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setEnabled(false)
		},
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			fmt.Printf("Enabled: %v\nDB: %s\n", cfg.Enabled, cfg.DatabasePath)
			return nil
		},
	}
}

func setEnabled(v bool) error {
	// naive: write config file if missing, then rewrite enabled flag
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	// write file
	home, _ := os.UserHomeDir()
	cfgDir := filepath.Join(home, ".tat")
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		return err
	}
	content := fmt.Sprintf(
		"enabled: %v\ndatabase_path: %q\nexclude_dirs: [\"%s\"]\ninclude_dirs: []\nmax_output_kb_per_command: %d\nredact_patterns:\n  - %q\n  - %q\n  - %q\n",
		v, cfg.DatabasePath, filepath.Join(home, ".ssh"), cfg.MaxOutputKB,
		`(?i)password=[^ ]+`,
		`(?i)authorization:\s*bearer\s+[A-Za-z0-9\-_\.]+`,
		`(?i)aws_secret_access_key=[A-Za-z0-9/+=]+`,
	)
	return os.WriteFile(cfgPath, []byte(content), 0o644)
}

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		Short: "Initialize TAT (config, database) and optionally install shell hooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			// Ensure ~/.tat directory exists
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

			hooksInstalled, err := checkShellHooks()
			if err != nil {
				// Non-fatal - just warn
				fmt.Printf("Warning: failed to check shell hooks: %v\n", err)
			}
			if !hooksInstalled {
				prompt := "Shell hooks for TAT are not installed. Install them now? (Y/n) ['n' for development mode]: "
				fmt.Print(prompt)
				var response string
				_, err := fmt.Scanln(&response)
				if err != nil || (strings.ToLower(response) != "n" && strings.ToLower(response) != "no") {
					// User said yes or pressed enter - install hooks
					fmt.Println("Installing shell hooks...")
					// Use your existing install-shell command code or call a function here
					if err := installShellHooks(); err != nil {
						return fmt.Errorf("failed to install shell hooks: %w", err)
					}
					fmt.Println("Shell hooks installed. Please restart your shell or run 'source ~/.bashrc' or 'source ~/.zshrc' as applicable.")
				} else {
					fmt.Println("Skipping shell hook installation. You can install them later with 'tat install-shell'.")
				}
			} else {
				fmt.Println("Shell hooks already installed.")
			}
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
func isHookInstalled(shellRcPath, sourceLine string) (bool, error) {
	f, err := os.Open(shellRcPath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == sourceLine {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, err
	}
	return false, nil
}

func checkShellHooks() (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}

	bashRc := filepath.Join(home, ".bashrc")
	zshRc := filepath.Join(home, ".zshrc")

	bashSource := "source ~/.tat/tat.bash"
	zshSource := "source ~/.tat/tat.zsh"

	bashInstalled, err1 := isHookInstalled(bashRc, bashSource)
	zshInstalled, err2 := isHookInstalled(zshRc, zshSource)

	// If any errors reading rc files, treat as false but log error downstream
	if err1 != nil && !errors.Is(err1, os.ErrNotExist) {
		return false, err1
	}
	if err2 != nil && !errors.Is(err2, os.ErrNotExist) {
		return false, err2
	}

	return bashInstalled || zshInstalled, nil
}

func installShellHooks() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	tatDir := filepath.Join(home, ".tat")
	if err := os.MkdirAll(tatDir, 0o755); err != nil {
		return err
	}

	// Copy hook scripts from your repo 'scripts' folder to ~/.tat/
	if err := copyFile("scripts/tat.bash", filepath.Join(tatDir, "tat.bash")); err != nil {
		return err
	}
	if err := copyFile("scripts/tat.zsh", filepath.Join(tatDir, "tat.zsh")); err != nil {
		return err
	}

	// Append source lines to rc files if missing
	bashRc := filepath.Join(home, ".bashrc")
	zshRc := filepath.Join(home, ".zshrc")

	if err := appendLineIfMissing(bashRc, "source ~/.tat/tat.bash"); err != nil {
		return err
	}
	if err := appendLineIfMissing(zshRc, "source ~/.tat/tat.zsh"); err != nil {
		return err
	}
	return nil
}

// Helper copies file contents
func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0o644)
}

func appendLineIfMissing(filename, line string) error {
	b, err := os.ReadFile(filename)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if strings.Contains(string(b), line) {
		return nil
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("\n" + line + "\n")
	return err
}

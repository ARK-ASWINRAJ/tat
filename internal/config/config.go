package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Enabled        bool     `mapstructure:"enabled"`
	DatabasePath   string   `mapstructure:"database_path"`
	ExcludeDirs    []string `mapstructure:"exclude_dirs"`
	IncludeDirs    []string `mapstructure:"include_dirs"`
	MaxOutputKB    int      `mapstructure:"max_output_kb_per_command"`
	RedactPatterns []string `mapstructure:"redact_patterns"`
}

func Default() Config {
	home, _ := os.UserHomeDir()
	return Config{
		Enabled:      true,
		DatabasePath: filepath.Join(home, ".tat", "tat.db"),
		ExcludeDirs:  []string{filepath.Join(home, ".ssh")},
		IncludeDirs:  []string{},
		MaxOutputKB:  1024,
		RedactPatterns: []string{
			`(?i)password=[^ ]+`,
			`(?i)authorization:\s*bearer\s+[A-Za-z0-9\-_\.]+`,
			`(?i)aws_secret_access_key=[A-Za-z0-9/+=]+`,
		},
	}
}

func Load() (Config, error) {
	home, _ := os.UserHomeDir()
	cfgDir := filepath.Join(home, ".tat")
	_ = os.MkdirAll(cfgDir, 0o755)

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(cfgDir)

	def := Default()
	v.SetDefault("enabled", def.Enabled)
	v.SetDefault("database_path", def.DatabasePath)
	v.SetDefault("exclude_dirs", def.ExcludeDirs)
	v.SetDefault("include_dirs", def.IncludeDirs)
	v.SetDefault("max_output_kb_per_command", def.MaxOutputKB)
	v.SetDefault("redact_patterns", def.RedactPatterns)

	_ = v.ReadInConfig() // ignore if not present

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return def, err
	}
	return c, nil
}

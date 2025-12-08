package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// LoadConfig は設定ファイル（config.toml）を読み込み、設定構造体を返します。
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みエラー: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("設定のパースエラー: %w", err)
	}

	// SanitizeRules
	cfg.SanitizeRules.Any = viper.GetStringMapString("setting.sanitize_rules.any")
	cfg.SanitizeRules.End = viper.GetStringMapString("setting.sanitize_rules.end")

	return &cfg, nil
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Setting    Setting    `mapstructure:"setting"`
	DirSetting DirSetting `mapstructure:"dir_setting"`
}

// Validate は設定値の妥当性をチェック
func (c *Config) Validate() error {
	// ディレクトリの存在確認
	dirs := []struct {
		path string
		name string
	}{
		{c.DirSetting.SourceDir, "source_dir"},
		{c.DirSetting.HtmlDir, "html_dir"},
		{c.DirSetting.LogDir, "log_dir"},
		{c.DirSetting.OutputDir, "output_dir"},
	}

	for _, dir := range dirs {
		if dir.path == "" {
			return fmt.Errorf("%sが設定されていません", dir.name)
		}

		// 相対パスを絶対パスに変換
		absPath, err := filepath.Abs(dir.path)
		if err != nil {
			return fmt.Errorf("%sの絶対パス変換に失敗: %w", dir.name, err)
		}

		// ディレクトリの存在確認（存在しない場合は作成）
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			if err := os.MkdirAll(absPath, 0755); err != nil {
				return fmt.Errorf("%sディレクトリの作成に失敗: %w", dir.name, err)
			}
		}
	}

	return nil
}

type Setting struct {
	SetMainImage   bool `mapstructure:"set_main_image"`
	SaveParsedData bool `mapstructure:"save_parsed_data"`
	Convert        bool `mapstructure:"convert"`
	Debug          bool `mapstructure:"debug"`
}

type DirSetting struct {
	SourceDir        string `mapstructure:"source_dir"`
	HtmlDir          string `mapstructure:"html_dir"`
	OutputDir        string `mapstructure:"output_dir"`
	LogDir           string `mapstructure:"log_dir"`
	Mp3OutputDirName string `mapstructure:"mp3_output_dir_name"`
}

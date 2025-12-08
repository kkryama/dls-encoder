package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// テスト用の設定ファイルを一時的に作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config")
	err := os.MkdirAll(configPath, 0755)
	if err != nil {
		t.Fatalf("テスト用ディレクトリの作成に失敗: %v", err)
	}

	configContent := `
[setting]
set_main_image = true
save_parsed_data = true
convert = true
debug = false

[setting.sanitize_rules.any]
"/" = "／"

[setting.sanitize_rules.end]
"." = "．"

[dir_setting]
source_dir = "./data/source"
html_dir = "./data/html"
output_dir = "./data/output"
log_dir = "./data/log"
mp3_output_dir_name = "mp3-output"
`
	err = os.WriteFile(filepath.Join(configPath, "config.toml"), []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("テスト用設定ファイルの作成に失敗: %v", err)
	}

	// カレントディレクトリを一時的に変更
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("カレントディレクトリの取得に失敗: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("元のディレクトリへの復帰に失敗: %v", err)
		}
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("カレントディレクトリの変更に失敗: %v", err)
	}

	// 設定の読み込みをテスト
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}

	// 期待される値のテスト
	if !cfg.Setting.SetMainImage {
		t.Error("SetMainImageの値が期待される値と異なります")
	}
	if !cfg.Setting.SaveParsedData {
		t.Error("SaveParsedDataの値が期待される値と異なります")
	}
	if !cfg.Setting.Convert {
		t.Error("Convertの値が期待される値と異なります")
	}
	if cfg.Setting.Debug {
		t.Error("Debugの値が期待される値と異なります")
	}

	if cfg.DirSetting.SourceDir != "./data/source" {
		t.Errorf("SourceDirの値が期待される値と異なります: got %v, want %v", cfg.DirSetting.SourceDir, "./data/source")
	}
	if cfg.DirSetting.HtmlDir != "./data/html" {
		t.Errorf("HtmlDirの値が期待される値と異なります: got %v, want %v", cfg.DirSetting.HtmlDir, "./data/html")
	}
	if cfg.DirSetting.OutputDir != "./data/output" {
		t.Errorf("OutputDirの値が期待される値と異なります: got %v, want %v", cfg.DirSetting.OutputDir, "./data/output")
	}
	if cfg.DirSetting.LogDir != "./data/log" {
		t.Errorf("LogDirの値が期待される値と異なります: got %v, want %v", cfg.DirSetting.LogDir, "./data/log")
	}
	if cfg.DirSetting.Mp3OutputDirName != "mp3-output" {
		t.Errorf("Mp3OutputDirNameの値が期待される値と異なります: got %v, want %v", cfg.DirSetting.Mp3OutputDirName, "mp3-output")
	}
}

func TestLoadConfig_Error(t *testing.T) {
	// 設定ファイルが存在しない状態でのテスト
	tempDir := t.TempDir()

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("カレントディレクトリの取得に失敗: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("元のディレクトリへの復帰に失敗: %v", err)
		}
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("カレントディレクトリの変更に失敗: %v", err)
	}

	_, err = LoadConfig()
	if err == nil {
		t.Error("設定ファイルが存在しない場合にエラーが発生すべき")
	}
}

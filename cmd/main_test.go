package main

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kkryama/dls-encoder/internal/config"
)

func TestProcessDirectoriesBuildsHtmlPathWithJoin(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	htmlDir := filepath.Join(tmpDir, "html")
	if err := os.MkdirAll(htmlDir, 0755); err != nil {
		t.Fatalf("HTMLディレクトリの作成に失敗: %v", err)
	}

	key := "d_123456"
	htmlPath := filepath.Join(htmlDir, key+".html")
	htmlContent := `<!DOCTYPE html><html><head><title>【作品】テスト作品(テストサークル)</title></head><body></body></html>`
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		t.Fatalf("HTMLファイルの作成に失敗: %v", err)
	}

	cfg := &config.Config{
		Setting: config.Setting{
			SetMainImage:   false,
			SaveParsedData: false,
		},
		DirSetting: config.DirSetting{
			SourceDir:        filepath.Join(tmpDir, "source"),
			HtmlDir:          htmlDir,
			OutputDir:        filepath.Join(tmpDir, "output"),
			LogDir:           filepath.Join(tmpDir, "log"),
			Mp3OutputDirName: "mp3-output",
			ImageDir:         filepath.Join(tmpDir, "image"),
		},
	}

	data, notApplicable, missingImage, err := processDirectories(ctx, cfg, []string{key})
	if err != nil {
		t.Fatalf("processDirectoriesの実行に失敗: %v", err)
	}

	if len(notApplicable) != 0 {
		t.Fatalf("処理対象外データが想定外に検出されました: %v", notApplicable)
	}

	if len(missingImage) != 0 {
		t.Fatalf("画像不足データが想定外に検出されました: %v", missingImage)
	}

	if _, ok := data[key]; !ok {
		t.Fatalf("データにキー%qが含まれていません", key)
	}
}

func TestSplitActorNames(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  []string
	}{
		{"commaAndMiddleDot", "Alice, Bob・Carol", []string{"Alice", "Bob", "Carol"}},
		{"slashes", "Alice／Bob / Carol", []string{"Alice", "Bob", "Carol"}},
		{"japaneseComma", "Alice、Bob，Carol", []string{"Alice", "Bob", "Carol"}},
		{"trimsEmpty", " Alice ,, Bob ", []string{"Alice", "Bob"}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := splitActorNames(tc.input)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("splitActorNames(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}

	if got := splitActorNames(""); got != nil {
		t.Fatalf("splitActorNames of empty string should return nil, got %v", got)
	}
}

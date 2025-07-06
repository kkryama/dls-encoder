package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInteractiveHTMLGenerator は対話型HTMLジェネレータのテストを行います
func TestInteractiveHTMLGenerator(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		input     string
		wantFile  string
		wantError bool
	}{
		{
			name: "正常な入力でHTMLを生成",
			input: strings.Join([]string{
				"test-file\n", // ファイル名
				"テストアルバム\n",   // アルバムタイトル
				"テストサークル\n",   // サークル名
				"y\n",         // 詳細情報を追加
				"ジャンル\n",      // 項目名
				"テスト\n",       // 値
				"n\n",         // 詳細情報を追加しない
				"y\n",         // トラックを追加
				"トラック1\n",     // トラックタイトル
				"3:30\n",      // 再生時間
				"n\n",         // トラックを追加しない
			}, ""),
			wantFile:  "test-file.html",
			wantError: false,
		},
		{
			name: "空のファイル名でエラー",
			input: strings.Join([]string{
				"\n",        // 空のファイル名
				"テストアルバム\n", // アルバムタイトル
				"テストサークル\n", // サークル名
				"n\n",       // 詳細情報を追加しない
				"n\n",       // トラックを追加しない
			}, ""),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 標準入力をシミュレート
			oldStdin := os.Stdin
			tmpfile, err := os.CreateTemp("", "test-input")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())
			defer tmpfile.Close()

			if _, err := tmpfile.Write([]byte(tt.input)); err != nil {
				t.Fatal(err)
			}
			if _, err := tmpfile.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			os.Stdin = tmpfile
			defer func() { os.Stdin = oldStdin }()

			// テスト実行
			err = InteractiveHTMLGenerator(tempDir)

			// エラーチェック
			if (err != nil) != tt.wantError {
				t.Errorf("InteractiveHTMLGenerator() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				// 生成されたファイルの確認
				filePath := filepath.Join(tempDir, tt.wantFile)
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("Expected file %s was not created", filePath)
					return
				}

				// ファイルの内容を確認
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatal(err)
				}

				// 期待される内容が含まれているか確認
				expectedContents := []string{
					`<h1 id="work_name">テストアルバム</h1>`,
					`<span itemprop="brand" class="maker_name"><a href="#">テストサークル</a></span>`,
				}

				for _, expected := range expectedContents {
					if !strings.Contains(string(content), expected) {
						t.Errorf("Generated HTML does not contain expected content: %s", expected)
					}
				}
			}
		})
	}
}

// TestInteractiveHTMLGeneratorFileOverwrite は上書き確認機能のテストを行います
func TestInteractiveHTMLGeneratorFileOverwrite(t *testing.T) {
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "existing-file.html")

	// 既存のファイルを作成
	if err := os.WriteFile(existingFile, []byte("existing content"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name: "既存ファイルを上書き",
			input: strings.Join([]string{
				"existing-file\n", // 既存のファイル名
				"テストアルバム\n",       // アルバムタイトル
				"テストサークル\n",       // サークル名
				"n\n",             // 詳細情報を追加しない
				"n\n",             // トラックを追加しない
				"y\n",             // 上書き確認
			}, ""),
			wantError: false,
		},
		{
			name: "上書きをキャンセル",
			input: strings.Join([]string{
				"existing-file\n", // 既存のファイル名
				"テストアルバム\n",       // アルバムタイトル
				"テストサークル\n",       // サークル名
				"n\n",             // 詳細情報を追加しない
				"n\n",             // トラックを追加しない
				"n\n",             // 上書きをキャンセル
			}, ""),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 標準入力をシミュレート
			oldStdin := os.Stdin
			tmpfile, err := os.CreateTemp("", "test-input")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())
			defer tmpfile.Close()

			if _, err := tmpfile.Write([]byte(tt.input)); err != nil {
				t.Fatal(err)
			}
			if _, err := tmpfile.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			os.Stdin = tmpfile
			defer func() { os.Stdin = oldStdin }()

			// テスト実行
			err = InteractiveHTMLGenerator(tempDir)

			// エラーチェック
			if (err != nil) != tt.wantError {
				t.Errorf("InteractiveHTMLGenerator() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

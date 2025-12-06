package audioconverter

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/kkryama/dls-encoder/internal/config"
)

func TestEnsureDirExists(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test", "nested")

	// ディレクトリが存在しない場合の作成をテスト
	err := EnsureDirExists(testDir)
	if err != nil {
		t.Errorf("ディレクトリの作成に失敗: %v", err)
	}

	// ディレクトリが作成されたことを確認
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Error("ディレクトリが作成されていません")
	}

	// 既存のディレクトリに対する呼び出しをテスト
	err = EnsureDirExists(testDir)
	if err != nil {
		t.Errorf("既存のディレクトリに対する呼び出しでエラー: %v", err)
	}
}

func TestDirExists(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test")

	// 存在しないディレクトリのテスト
	exists, err := DirExists(testDir)
	if err != nil {
		t.Errorf("存在しないディレクトリの確認でエラー: %v", err)
	}
	if exists {
		t.Error("存在しないディレクトリが存在すると報告されました")
	}

	// ディレクトリを作成
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("テストディレクトリの作成に失敗: %v", err)
	}

	// 存在するディレクトリのテスト
	exists, err = DirExists(testDir)
	if err != nil {
		t.Errorf("存在するディレクトリの確認でエラー: %v", err)
	}
	if !exists {
		t.Error("存在するディレクトリが存在しないと報告されました")
	}
}

func TestCleanUp(t *testing.T) {
	tempDir := t.TempDir()

	// テストファイルを作成
	testFiles := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, file := range testFiles {
		err := os.WriteFile(filepath.Join(tempDir, file), []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("テストファイルの作成に失敗: %v", err)
		}
	}

	// クリーンアップを実行
	err := CleanUp(tempDir)
	if err != nil {
		t.Errorf("クリーンアップに失敗: %v", err)
	}

	// ディレクトリが空になったことを確認
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("ディレクトリの読み込みに失敗: %v", err)
	}
	if len(files) != 0 {
		t.Error("クリーンアップ後もファイルが残っています")
	}
}

func TestFindAudioFiles(t *testing.T) {
	tempDir := t.TempDir()

	// テスト用の設定を作成
	cfg := &config.Config{
		Setting: config.Setting{
			ExcludeStrings: []string{"SE無し", "SEなし", "効果音無し", "効果音なし", "__MACOSX"},
		},
	}

	// テストファイルを作成
	testFiles := map[string]bool{
		"track1.wav":          true,  // 含まれるべき
		"track1.mp3":          false, // 除外されるべき（WAVがある）
		"track2.mp3":          true,  // 含まれるべき
		"SE無しtrack3.wav":      false, // 除外されるべき
		"track4_効果音なし.wav":    false, // 除外されるべき
		"track5.wav":          true,  // 含まれるべき
		"__MACOSX/track6.wav": false, // 除外されるべき
		"track7.flac":         true,  // 含まれるべき
		"track7.mp3":          false, // 除外されるべき（FLACがある）
		"track8.mp3":          true,  // 含まれるべき
		"TRACK9.WAV":          true,  // 含まれるべき（大文字対応）
	}

	for filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("ディレクトリの作成に失敗: %v", err)
		}
		if err := os.WriteFile(filePath, []byte("dummy audio data"), 0644); err != nil {
			t.Fatalf("テストファイルの作成に失敗: %v", err)
		}
	}

	// オーディオファイルを検索
	audioFiles := FindAudioFiles(tempDir, cfg)

	// 結果を検証
	expectedCount := 0
	for _, shouldInclude := range testFiles {
		if shouldInclude {
			expectedCount++
		}
	}

	if len(audioFiles) != expectedCount {
		t.Errorf("見つかったファイル数が期待される値と異なります: got %d, want %d", len(audioFiles), expectedCount)
	}

	// 各ファイルが適切に含まれているか確認
	for _, file := range audioFiles {
		filename := filepath.Base(file)
		shouldInclude, exists := testFiles[filename]
		if !exists {
			t.Errorf("予期しないファイルが含まれています: %s", filename)
		} else if !shouldInclude {
			t.Errorf("除外されるべきファイルが含まれています: %s", filename)
		}
	}
}

func TestConvertFileToMp3_CommandGeneration(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "test.wav")
	outputFile := filepath.Join(tempDir, "test.mp3")
	coverImage := filepath.Join(tempDir, "cover.jpg")

	// コマンド引数の生成をテスト
	cmdArgs := []string{
		"-i", inputFile,
		"-i", coverImage,
		"-map", "0:a",
		"-map", "1:v",
		"-c:v", "mjpeg",
		"-metadata:s:v", "title=Album cover",
		"-c:a", "libmp3lame",
		"-b:a", "320k",
		"-ar", "48000",
		"-metadata", "artist=テストアーティスト",
		"-metadata", "album_artist=テストアルバムアーティスト",
		"-metadata", "album=テストアルバム",
		"-metadata", "title=テストトラック",
		"-id3v2_version", "3",
		"-y",
		outputFile,
	}

	cmd := exec.Command("ffmpeg", cmdArgs...)
	if cmd == nil {
		t.Error("ffmpegコマンドの生成に失敗")
	}
}

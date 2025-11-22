package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kkryama/dls-encoder/internal/model"
)

func TestFindMainImage(t *testing.T) {
	// テスト用のディレクトリ構造を作成
	tempDir := t.TempDir()
	imageDir := filepath.Join(tempDir, "image")

	err := os.MkdirAll(imageDir, 0755)
	if err != nil {
		t.Fatalf("テストディレクトリの作成に失敗: %v", err)
	}

	// テスト用の画像ファイルを作成
	testFiles := map[string]bool{
		"RJ12345678.webp": true,
		"RJ12345678.jpg":  true,
		"other_image.jpg": false,
	}

	for filename := range testFiles {
		path := filepath.Join(imageDir, filename)
		err := os.WriteFile(path, []byte("dummy image data"), 0644)
		if err != nil {
			t.Fatalf("テストファイルの作成に失敗 %s: %v", filename, err)
		}
	}

	// メインイメージを検索
	mainImage, err := FindMainImage(imageDir, "RJ12345678")
	if err != nil {
		t.Fatalf("FindMainImageの実行に失敗: %v", err)
	}

	// webpファイルが優先して見つかることを確認
	expectedPath := filepath.Join(imageDir, "RJ12345678.webp")
	if mainImage != expectedPath {
		t.Errorf("FindMainImage: got %q, want %q", mainImage, expectedPath)
	}

	// webpファイルを削除して、jpgが見つかることを確認
	err = os.Remove(expectedPath)
	if err != nil {
		t.Fatalf("テストファイルの削除に失敗: %v", err)
	}

	mainImage, err = FindMainImage(imageDir, "RJ12345678")
	if err != nil {
		t.Fatalf("FindMainImageの2回目の実行に失敗: %v", err)
	}

	expectedPath = filepath.Join(imageDir, "RJ12345678.jpg")
	if mainImage != expectedPath {
		t.Errorf("FindMainImage (jpg): got %q, want %q", mainImage, expectedPath)
	}
}

func TestLoadTargets(t *testing.T) {
	// テスト用のディレクトリ構造を作成
	tempDir := t.TempDir()

	// ディレクトリとファイルを作成
	testDirs := []string{"dir1", "dir2", "dir3"}
	for _, dir := range testDirs {
		err := os.Mkdir(filepath.Join(tempDir, dir), 0755)
		if err != nil {
			t.Fatalf("テストディレクトリの作成に失敗 %s: %v", dir, err)
		}
	}

	// ファイルも作成（これらは無視されるべき）
	err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("test"), 0644)
	if err != nil {
		t.Fatalf("テストファイルの作成に失敗: %v", err)
	}

	// LoadTargetsを実行
	targets, err := LoadTargets(tempDir)
	if err != nil {
		t.Fatalf("LoadTargetsの実行に失敗: %v", err)
	}

	// 結果を検証
	if len(targets) != len(testDirs) {
		t.Errorf("ターゲット数が不一致: got %d, want %d", len(targets), len(testDirs))
	}

	// 各ディレクトリが含まれているか確認
	for _, dir := range testDirs {
		found := false
		for _, target := range targets {
			if target == dir {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ディレクトリ %q が結果に含まれていません", dir)
		}
	}
}

func TestSaveJSON(t *testing.T) {
	// テストデータを作成
	testData := map[string]model.IndividualData{
		"RJ12345678": {
			AlbumTitle: "テストアルバム",
			Actor:      "テスト声優",
			Brand:      "テストブランド",
			TrackList: []model.Track{
				{TrackTitle: "トラック1", TrackDuration: "3:45"},
			},
			Additional: map[string]string{
				"ジャンル": "テスト",
			},
		},
	}

	// テスト用の出力ディレクトリを作成
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output")

	// JSONを保存
	err := SaveJSON(outputPath, testData)
	if err != nil {
		t.Fatalf("SaveJSONの実行に失敗: %v", err)
	}

	// 保存されたJSONファイルを読み込んで検証
	jsonPath := filepath.Join(filepath.Dir(outputPath), "RJ12345678.json")
	jsonFile, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("保存されたJSONファイルの読み込みに失敗: %v", err)
	}

	var savedData model.IndividualData
	err = json.Unmarshal(jsonFile, &savedData)
	if err != nil {
		t.Fatalf("JSONのアンマーシャルに失敗: %v", err)
	}

	// データの検証
	expected := testData["RJ12345678"]
	if savedData.AlbumTitle != expected.AlbumTitle {
		t.Errorf("AlbumTitle: got %q, want %q", savedData.AlbumTitle, expected.AlbumTitle)
	}
	if savedData.Actor != expected.Actor {
		t.Errorf("Actor: got %q, want %q", savedData.Actor, expected.Actor)
	}
	if savedData.Brand != expected.Brand {
		t.Errorf("Brand: got %q, want %q", savedData.Brand, expected.Brand)
	}
	if len(savedData.TrackList) != len(expected.TrackList) {
		t.Errorf("TrackList length: got %d, want %d", len(savedData.TrackList), len(expected.TrackList))
	}
	if genre := savedData.Additional["ジャンル"]; genre != expected.Additional["ジャンル"] {
		t.Errorf("Additional[ジャンル]: got %q, want %q", genre, expected.Additional["ジャンル"])
	}
}

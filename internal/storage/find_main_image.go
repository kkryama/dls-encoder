package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindMainImage は指定されたキーに対応するメイン画像ファイルを検索します。
// imageDir 直下から、メイン画像（webp または jpg）を探します。
func FindMainImage(imageDir, key string) (string, error) {
	var mainImagePath string

	// imageDir 直下のファイルを調べる
	entries, err := os.ReadDir(imageDir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", imageDir, err)
	}

	imageFileNameWebp := fmt.Sprintf("%s.webp", key)
	imageFileNameJpg := fmt.Sprintf("%s.jpg", key)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := filepath.Base(entry.Name())
		lowerFileName := strings.ToLower(fileName)

		// .webp または .jpg ファイルのみ対象、かつファイル名が一致するか確認
		if fileName == imageFileNameWebp && strings.HasSuffix(lowerFileName, ".webp") {
			mainImagePath = filepath.Join(imageDir, fileName)
			// webp を見つけたら即終了
			break
		} else if fileName == imageFileNameJpg && strings.HasSuffix(lowerFileName, ".jpg") && mainImagePath == "" {
			// jpg は webp が見つかっていない場合のみセット
			mainImagePath = filepath.Join(imageDir, fileName)
		}
	}

	return mainImagePath, nil
}

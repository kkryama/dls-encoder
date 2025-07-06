package storage

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FindMainImage は指定されたキーに対応するメイン画像ファイルを検索します。
// HTMLディレクトリ内でキーを含むディレクトリから、メイン画像（webp または jpg）を探します。
func FindMainImage(htmlDir, key string) (string, error) {
	var mainImagePath string

	// htmlDir 直下のディレクトリのうち key を持ったディレクトリを取得する
	var searchDir string
	entries, err := os.ReadDir(htmlDir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", htmlDir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.Contains(entry.Name(), key) {
			searchDir = filepath.Join(htmlDir, entry.Name())
		}
	}

	// 対象ディレクトリから名前に [key]_img_main を含むファイルが存在するか調べる
	imageFileName := fmt.Sprintf("%s_img_main", key)
	err = filepath.WalkDir(searchDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fileName := filepath.Base(path)
		lowerFileName := strings.ToLower(fileName)

		// .webp または .jpg ファイルのみ対象、かつファイル名に imageFileName を含むか確認
		if strings.Contains(fileName, imageFileName) {
			if strings.HasSuffix(lowerFileName, ".webp") {
				mainImagePath = path
				// webp を見つけたら即終了
				return fs.SkipDir
			} else if strings.HasSuffix(lowerFileName, ".jpg") && mainImagePath == "" {
				// jpg は webp が見つかっていない場合のみセット
				mainImagePath = path
				// まだ webp が見つかるかもしれないので継続
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error walking directory %s: %w", searchDir, err)
	}

	if mainImagePath != "" {
		return mainImagePath, nil // 見つかったら即返す
	}

	return mainImagePath, nil
}

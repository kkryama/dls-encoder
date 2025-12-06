package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindMainImage は指定されたキーに対応するメイン画像ファイルを検索します。
// imageDir 直下から、メイン画像（webp または jpg）を探します。
func FindMainImage(imageDir, key string) (string, error) {
	candidates := []string{
		filepath.Join(imageDir, fmt.Sprintf("%s.webp", key)),
		filepath.Join(imageDir, fmt.Sprintf("%s.jpg", key)),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to access %s: %w", candidate, err)
		}
	}

	return "", nil
}

package storage

import (
	"fmt"
	"os"
)

// LoadTargets は指定されたベースディレクトリ内のサブディレクトリ名を取得します。
// 処理対象となるディレクトリの一覧を返します。
func LoadTargets(baseDir string) ([]string, error) {
	var targets []string
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("ディレクトリ(%v)の読み込みに失敗しました: %v", baseDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			targets = append(targets, entry.Name())
		}
	}

	return targets, nil
}

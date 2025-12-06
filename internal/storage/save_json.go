package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kkryama/dls-encoder/internal/model" // モデルパッケージのインポート
)

// SaveJSON は個別データをJSON形式でファイルに保存します。
// 各データはキーごとに別々のJSONファイルとして保存されます。
func SaveJSON(fileOutputDir string, data map[string]model.IndividualData) error {
	if err := os.MkdirAll(fileOutputDir, 0755); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}

	for key, value := range data {
		filename := filepath.Join(fileOutputDir, key+".json")
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("JSONファイルの作成に失敗しました: %w", err)
		}

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err = encoder.Encode(value); err != nil {
			file.Close()
			return fmt.Errorf("JSONのエンコードに失敗しました: %w", err)
		}

		if err := file.Close(); err != nil {
			return fmt.Errorf("JSONファイルのクローズに失敗しました: %w", err)
		}
	}

	return nil
}

package audioconverter

import (
	"os"
	"path/filepath"
	"strings"
)

// FindAudioFiles は指定されたディレクトリから音声ファイルを検索し、パスのリストを返します。
// WAVファイルを優先し、WAVが存在しない場合のみMP3ファイルを対象とします。
func FindAudioFiles(directory string) []string {
	var audioFiles []string
	seen := make(map[string]bool)                                            // 拡張子を除いたファイル名を記録するマップ
	excludeStrings := []string{"SE無し", "SEなし", "効果音無し", "効果音なし", "__MACOSX"} // パス中に含まれていたら除外する文字列リスト

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// 除外対象の文字列が含まれている場合はスキップ
			for _, excl := range excludeStrings {
				if strings.Contains(path, excl) {
					return nil
				}
			}

			name := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			ext := strings.ToLower(filepath.Ext(info.Name()))

			if ext == ".wav" {
				// `.wav` を見つけたら記録し、リストに追加
				seen[name] = true
				audioFiles = append(audioFiles, path)
			} else if ext == ".mp3" {
				// `.mp3` の場合、まだ `.wav` が見つかっていないなら追加
				if _, exists := seen[name]; !exists {
					seen[name] = false // `.mp3` を追加したことを記録
					audioFiles = append(audioFiles, path)
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil
	}

	// `.mp3` を追加した後で、`.wav` を追加していた場合は `.mp3` を削除
	filteredFiles := []string{}
	for _, path := range audioFiles {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if seen[name] { // `.wav` がある場合のみ `.mp3` をスキップ
			if strings.ToLower(filepath.Ext(path)) == ".mp3" {
				continue
			}
		}
		filteredFiles = append(filteredFiles, path)
	}

	return filteredFiles
}

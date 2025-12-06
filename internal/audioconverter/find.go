package audioconverter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kkryama/dls-encoder/internal/config"
	"github.com/kkryama/dls-encoder/internal/logger"
)

// FindAudioFiles は指定されたディレクトリから音声ファイルを検索し、パスのリストを返します。
// WAVファイルを優先し、WAVが存在しない場合FLACを、次にMP3ファイルを対象とします。
func FindAudioFiles(directory string, cfg *config.Config) []string {
	audioFiles := make(map[string]string)        // 拡張子を除いたファイル名をキーとして、対応するファイルのフルパスを値に持つマップ
	seen := make(map[string]int)                 // 拡張子を除いたファイル名をキーとして、そのファイルの優先度（WAV:3, FLAC:2, MP3:1）を値に持つマップ
	excludeStrings := cfg.Setting.ExcludeStrings // パス中に含まれていたら除外する文字列リスト

	// 優先度: WAV > FLAC > MP3
	priority := map[string]int{
		".wav":  3,
		".flac": 2,
		".mp3":  1,
	}

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// 除外対象の文字列が含まれている場合はスキップ
			for _, excl := range excludeStrings {
				if strings.Contains(path, excl) {
					logger.Debug(fmt.Sprintf("除外文字列 '%s' がパス '%s' に含まれているため、スキップします。", excl, path))
					return nil
				}
			}

			name := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())) // 拡張子を除いたファイル名を取得
			ext := strings.ToLower(filepath.Ext(info.Name()))                  // 拡張子を小文字で取得（大文字対応）

			// 優先度の高い拡張子の場合にのみマップを更新
			// 例1: "track1.mp3" (p=1) が見つかると、seen["track1"] = 0 < 1 なので、seen["track1"] = 1, audioFiles["track1"] = "/path/to/track1.mp3"。
			// 例2: 次に "track1.wav" (p=3) が見つかると、seen["track1"] = 1 < 3 なので、seen["track1"] = 3, audioFiles["track1"] = "/path/to/track1.wav" に更新。
			// 例3: 次に "track1.flac" (p=2) が見つかると、seen["track1"] = 3 > 2 なので、更新されません（WAV が優先）。
			if p, ok := priority[ext]; ok {
				if seen[name] < p {
					prevPriority := seen[name]
					prevPath := audioFiles[name]
					seen[name] = p          // 優先度を更新
					audioFiles[name] = path // パスを記録
					if prevPriority > 0 {
						logger.Debug(fmt.Sprintf("%s は優先度 %d から %d に更新されたため '%s' を '%s' へ差し替えました。", name, prevPriority, p, prevPath, path))
					} else {
						logger.Debug(fmt.Sprintf("%s を優先度 %d のファイル '%s' として登録しました。", name, p, path))
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		logger.Warn(fmt.Sprintf("音声ファイル探索中にエラー: %v", err))
		return nil
	}

	// マップからリストに変換
	var result []string
	keys := make([]string, 0, len(audioFiles))
	for name := range audioFiles {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		result = append(result, audioFiles[name])
	}

	return result
}

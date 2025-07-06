package audioconverter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// MP3Metadata はMP3ファイルのメタデータを格納する構造体です。
type MP3Metadata struct {
	Artist      string  // アーティスト名
	AlbumArtist string  // アルバムアーティスト名
	AlbumTitle  string  // アルバムタイトル
	TrackName   string  // トラック名
	CoverImage  *string // 画像ファイルのパス（nil の場合は画像なし）
}

// EnsureDirExists はディレクトリが存在しない場合に作成します。
func EnsureDirExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("ディレクトリの作成に失敗: %w", err)
		}
	}
	return nil
}

// DirExists はディレクトリが存在するかどうかを確認します。
func DirExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

// CleanUp はディレクトリ内のすべてのファイルを削除します。
func CleanUp(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		filePath := filepath.Join(dir, file.Name())
		if err := os.Remove(filePath); err != nil {
			return err
		}
	}

	return nil
}

// ConvertFileToMp3 は音声ファイルをMP3形式に変換します。
func ConvertFileToMp3(inputFile, mp3File string, metadata MP3Metadata) error {
	return ConvertFileToMp3WithContext(context.Background(), inputFile, mp3File, metadata)
}

// ConvertFileToMp3WithContext はコンテキスト対応で音声ファイルをMP3形式に変換します。
func ConvertFileToMp3WithContext(ctx context.Context, inputFile, mp3File string, metadata MP3Metadata) error {
	cmdArgs := []string{
		"-i", inputFile, // 入力ファイル
	}

	if metadata.CoverImage != nil && *metadata.CoverImage != "" {
		cmdArgs = append(cmdArgs,
			"-i", *metadata.CoverImage, // 画像ファイルを入力として追加
			"-map", "0:a", // 最初の入力 (wav) のオーディオストリームを使用
			"-map", "1:v", // 2つ目の入力 (画像) のビデオストリームを使用
			"-c:v", "mjpeg", // JPEG 画像として保存
			"-metadata:s:v", "title=Album cover", // 画像のメタデータ
		)
	}

	cmdArgs = append(cmdArgs,
		"-c:a", "libmp3lame", // LAME MP3 エンコーダを使用
		// "-q:a", "2", // MP3 の品質を設定（0が最高品質、9が最低品質）
		"-b:a", "320k", // 320kbps の固定ビットレート
		"-ar", "48000", // サンプリングレートを 48kHz に設定
		"-metadata", "artist="+metadata.Artist,
		"-metadata", "album_artist="+metadata.AlbumArtist,
		"-metadata", "album="+metadata.AlbumTitle,
		"-metadata", "title="+metadata.TrackName,
		"-id3v2_version", "3", // ID3v2.3 を使用
		"-y",    // 出力ファイルを強制的に上書き
		mp3File, // 出力ファイルのパス
	)

	// ffmpeg でエンコードする（コンテキスト対応）
	cmd := exec.CommandContext(ctx, "ffmpeg", cmdArgs...)
	err := cmd.Run()
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("変換処理がキャンセルされました: %w", ctx.Err())
		}
		return fmt.Errorf("ファイル変換に失敗しました %s -> %s (コマンド引数: %v): %w", inputFile, mp3File, cmdArgs, err)
	}
	return nil
}

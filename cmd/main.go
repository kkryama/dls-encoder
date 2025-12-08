// Package main はDLサイトからダウンロードしたコンテンツのエンコーダーを提供します。
// HTMLファイルから必要な情報を抽出し、音声ファイルをMP3形式に変換します。
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/kkryama/dls-encoder/internal/audioconverter"
	"github.com/kkryama/dls-encoder/internal/config"
	"github.com/kkryama/dls-encoder/internal/generator"
	"github.com/kkryama/dls-encoder/internal/logger"
	"github.com/kkryama/dls-encoder/internal/model"
	"github.com/kkryama/dls-encoder/internal/parser"
	"github.com/kkryama/dls-encoder/internal/storage"
)

const (
	logFileFormat   = "results_%s.log"
	timeStampFormat = "20060102_150405"
	mp3Extension    = ".mp3"
	version         = "20251209-073257" // 作業日時で更新
)

// truncateAlbumTitle はアルバムタイトルが長すぎる場合に省略します。
func truncateAlbumTitle(title string) string {
	const maxLen = 20
	runes := []rune(title)
	if len(runes) <= maxLen {
		return title
	}
	return string(runes[:maxLen]) + "(…略)"
}

// sanitizeDirName はディレクトリ名から無効な文字を置換します。
// config で定義された sanitize_rules に基づいて置き換えを行います。
func sanitizeDirName(name string, cfg *config.Config) string {
	result := name
	// any: 常に置き換え
	for from, to := range cfg.SanitizeRules.Any {
		result = strings.ReplaceAll(result, from, to)
	}
	// end: 末尾のみ置き換え
	for from, to := range cfg.SanitizeRules.End {
		trimmed := strings.TrimRight(result, from)
		count := len(result) - len(trimmed)
		result = trimmed + strings.Repeat(to, count)
	}
	return result
}

// main はプログラムのエントリーポイントです。
// エラーが発生した場合は、エラーメッセージを表示して終了します。
func main() {
	// コンテキストとシグナルハンドリングの設定
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// シグナルハンドラーでグレースフルシャットダウン
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.LogMessage("終了シグナルを受信しました。処理を停止します...")
		cancel()
	}()

	// フラグの定義
	createHTML := flag.Bool("create-html", false, "HTMLファイルを対話形式で作成します")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("設定ファイルの読み込みに失敗: %v\n", err)
		os.Exit(1)
	}

	// 設定値のバリデーション
	if err := cfg.Validate(); err != nil {
		fmt.Printf("設定値の検証に失敗: %v\n", err)
		os.Exit(1)
	}

	if *createHTML {
		// HTML生成モード
		if err := generator.InteractiveHTMLGenerator(cfg.DirSetting.HtmlDir, cfg.DirSetting.ImageDir); err != nil {
			fmt.Printf("HTMLファイルの生成に失敗しました: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 通常のエンコード処理
	if err := runWithContext(ctx, cfg); err != nil {
		fmt.Printf("エラー: %v\n", err)
		os.Exit(1)
	}
}

// runWithContext はコンテキストを使用して変換処理の全体フローを制御します。
// 依存関係の確認、ログの初期化、HTMLの解析、MP3変換を実行します。
func runWithContext(ctx context.Context, cfg *config.Config) error {
	logger.LogMessage("dls-encoder version: " + version)

	if err := validateDependencies(); err != nil {
		return fmt.Errorf("依存関係の確認に失敗: %w", err)
	}

	logFile, err := setupLogging(cfg.DirSetting.LogDir, cfg.Setting.Debug)
	if err != nil {
		return fmt.Errorf("ログ設定の初期化に失敗: %w", err)
	}
	defer func() {
		if logFile != nil {
			if err := logFile.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "ログファイルのクローズでエラー: %v\n", err)
			}
		}
	}()

	targetDirs, err := storage.LoadTargets(cfg.DirSetting.SourceDir)
	if err != nil {
		return fmt.Errorf("対象ディレクトリ一覧の読み込みに失敗: %w", err)
	}
	logger.LogMessage(fmt.Sprintf("対象ディレクトリ: %v", targetDirs))
	logger.LogDebugEvent("target_directories_loaded", map[string]interface{}{
		"directories": targetDirs,
		"count":       len(targetDirs),
	})

	data, notApplicableData, missingImageData, err := processDirectories(ctx, cfg, targetDirs)
	if err != nil {
		return fmt.Errorf("ディレクトリの処理に失敗: %w", err)
	}

	if err := handleConversion(ctx, cfg, data, notApplicableData, missingImageData); err != nil {
		return fmt.Errorf("変換処理に失敗: %w", err)
	}

	logger.LogMessage("処理が正常に終了しました。")
	return nil
}

// validateDependencies はffmpegコマンドが利用可能かを確認します。
// ffmpegが見つからない場合はエラーを返します。
func validateDependencies() error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg がインストールされていない、または PATH に見つかりません")
	}
	return nil
}

// setupLogging は指定されたディレクトリにログファイルを作成し、初期化します。
// タイムスタンプを含むログファイル名を生成し、ロガーの設定を行います。
func setupLogging(logDir string, debug bool) (*os.File, error) {
	if len(logDir) == 0 {
		return nil, fmt.Errorf("ログディレクトリが設定されていません")
	}

	timestamp := time.Now().Format(timeStampFormat)
	logFilePath := filepath.Join(logDir, fmt.Sprintf(logFileFormat, timestamp))

	logFile, err := logger.Init(logFilePath, debug)
	if err != nil {
		// logger.Init内でファイルが部分的に開かれている場合の対処
		return nil, fmt.Errorf("ログファイルの初期化に失敗: %w", err)
	}
	return logFile, nil
}

// processDirectories はターゲットディレクトリ内のHTMLファイルを処理します。
// HTML解析、メイン画像の確認、JSONデータの保存を行います。
// 処理結果として、個別データ、処理対象外データ、画像不足データを返します。
func processDirectories(ctx context.Context, cfg *config.Config, targetDirs []string) (map[string]model.IndividualData, []string, []string, error) {
	logger.LogDebugEvent("processDirectories_called", map[string]interface{}{
		"targetDirs": targetDirs,
		"sourceDir":  cfg.DirSetting.SourceDir,
		"htmlDir":    cfg.DirSetting.HtmlDir,
	})
	data := make(map[string]model.IndividualData)
	var notApplicableData []string
	var missingImageData []string

	for _, targetDir := range targetDirs {
		key := filepath.Base(targetDir)
		targetHtml := filepath.Join(cfg.DirSetting.HtmlDir, targetDir+".html")

		if err := processDirectory(cfg, targetHtml, key, data, &notApplicableData, &missingImageData); err != nil {
			logger.LogWarnEvent("directory_processing_error", map[string]interface{}{
				"error":      err.Error(),
				"key":        key,
				"targetHtml": targetHtml,
				"message":    fmt.Sprintf("ディレクトリの処理でエラーが発生: %v", err),
			})
			continue
		}
	}

	if cfg.Setting.SaveParsedData {
		if err := storage.SaveJSON(cfg.DirSetting.LogDir, data); err != nil {
			return nil, nil, nil, fmt.Errorf("変換前のデータ保存に失敗: %w", err)
		}
	}

	return data, notApplicableData, missingImageData, nil
}

// processDirectory は単一のディレクトリのHTMLファイルを処理します。
// HTMLファイルの存在確認、解析、メイン画像の処理を行います。
// エラー発生時は処理対象外リストに追加します。
func processDirectory(cfg *config.Config, targetHtml, key string, data map[string]model.IndividualData, notApplicableData, missingImageData *[]string) error {
	logger.LogDebugEvent("processDirectory_called", map[string]interface{}{
		"targetHtml": targetHtml,
		"key":        key,
	})

	if _, err := os.Stat(targetHtml); err != nil {
		*notApplicableData = append(*notApplicableData, key)
		return fmt.Errorf("HTMLファイルのアクセスに失敗: %w", err)
	}

	individualData, err := parser.ExtractData(targetHtml, key, cfg)
	if err != nil {
		*notApplicableData = append(*notApplicableData, key)
		return fmt.Errorf("データの取得に失敗: %w", err)
	}

	if cfg.Setting.SetMainImage {
		if err := processMainImage(cfg.DirSetting.ImageDir, key, &individualData, missingImageData); err != nil {
			return err
		}
	}

	data[key] = individualData
	return nil
}

// processMainImage はメイン画像を検索し、データにパスを設定します。
// 画像が見つからない、またはエラーが発生した場合は画像不足リストに追加します。
func processMainImage(imageDir string, key string, individualData *model.IndividualData, missingImageData *[]string) error {
	logger.LogDebugEvent("processMainImage_called", map[string]interface{}{
		"key":        key,
		"albumTitle": individualData.AlbumTitle,
		"actor":      individualData.Actor,
	})
	mainImagePath, err := storage.FindMainImage(imageDir, key)
	if err != nil {
		*missingImageData = append(*missingImageData, key)
		return fmt.Errorf("メイン画像の検索に失敗: %w", err)
	}

	if len(mainImagePath) == 0 {
		*missingImageData = append(*missingImageData, key)
		return fmt.Errorf("メイン画像が見つかりません")
	}

	// 相対パスを絶対パスに変換
	absImagePath, err := filepath.Abs(mainImagePath)
	if err != nil {
		*missingImageData = append(*missingImageData, key)
		return fmt.Errorf("メイン画像の絶対パスの取得に失敗: %w", err)
	}

	individualData.MainImage = absImagePath
	return nil
}

// handleConversion は音声ファイルのMP3変換を実行します。
// 設定に基づいて変換処理の実行可否を判断し、処理結果をログ出力します。
func handleConversion(ctx context.Context, cfg *config.Config, data map[string]model.IndividualData, notApplicableData, missingImageData []string) error {
	logger.LogDebugEvent("handleConversion_called", map[string]interface{}{
		"dataCount":          len(data),
		"notApplicableCount": len(notApplicableData),
		"missingImageCount":  len(missingImageData),
		"conversionEnabled":  cfg.Setting.Convert,
	})

	if !cfg.Setting.Convert {
		logger.LogMessage("変換処理がOFFに設定されているため、処理を終了します")
		return nil
	}

	keys := getSortedKeys(data)
	logger.LogMessage(fmt.Sprintf("処理対象: %v", keys))
	logger.LogDebugEvent("processing_targets", map[string]interface{}{
		"keys":  keys,
		"count": len(keys),
	})
	if len(notApplicableData) > 0 || len(missingImageData) > 0 {
		var allNotApplicable []string
		allNotApplicable = append(allNotApplicable, notApplicableData...)
		allNotApplicable = append(allNotApplicable, missingImageData...)
		logger.LogMessage(fmt.Sprintf("処理対象外: %v", allNotApplicable))
		logger.LogDebugEvent("excluded_targets", map[string]interface{}{
			"items": allNotApplicable,
			"count": len(allNotApplicable),
		})
	}

	for _, key := range keys {
		if err := convertFiles(ctx, cfg, key, data[key]); err != nil {
			return fmt.Errorf("[%s]の変換に失敗: %w", key, err)
		}
	}

	printResults(cfg, notApplicableData, missingImageData)
	return nil
}

// getSortedKeys はマップのキーをソートしたスライスを返します。
func getSortedKeys(data map[string]model.IndividualData) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// convertFiles は指定されたディレクトリ内の音声ファイルをMP3に変換します。
// 出力先の準備、メタデータの設定、ファイルの変換を行います。
func convertFiles(ctx context.Context, cfg *config.Config, key string, value model.IndividualData) error {
	logger.LogDebugEvent("convertFiles_called", map[string]interface{}{
		"key":        key,
		"albumTitle": value.AlbumTitle,
		"actor":      value.Actor,
		"brand":      value.Brand,
		"sourceDir":  cfg.DirSetting.SourceDir,
		"outputDir":  cfg.DirSetting.OutputDir,
	})

	targetDir := filepath.Join(cfg.DirSetting.SourceDir, key)
	audioFiles := audioconverter.FindAudioFiles(targetDir, cfg)

	if len(audioFiles) == 0 {
		return fmt.Errorf("音声ファイルが見つかりません: %s", targetDir)
	}

	// Actor が複数の場合、省略してディレクトリ名を短くする
	actors := splitActorNames(value.Actor)
	if len(actors) == 0 {
		trimmed := strings.TrimSpace(value.Actor)
		if trimmed != "" {
			actors = []string{trimmed}
		}
	}
	if len(actors) > 2 {
		actors = append(actors[:2], "他")
	}
	actorDir := sanitizeDirName(strings.Join(actors, "・"), cfg)

	shortAlbumTitle := sanitizeDirName(truncateAlbumTitle(value.AlbumTitle), cfg)

	mp3OutputDir := filepath.Join(cfg.DirSetting.OutputDir, cfg.DirSetting.Mp3OutputDirName, actorDir, sanitizeDirName(value.Brand, cfg), fmt.Sprintf("【%s】%s", key, shortAlbumTitle))

	if err := prepareOutputDirectory(mp3OutputDir); err != nil {
		return err
	}

	var coverImage *string
	if value.MainImage != "" {
		coverImage = &value.MainImage
	}

	baseMetaData := audioconverter.MP3Metadata{
		Artist:      value.Actor,
		AlbumArtist: value.Brand,
		AlbumTitle:  value.AlbumTitle,
		CoverImage:  coverImage,
	}

	// MP3メタデータのデバッグログを出力
	logger.LogDebugEvent("mp3_metadata_prepared", map[string]interface{}{
		"key":        key,
		"coverImage": value.MainImage,
		"artist":     value.Actor,
		"albumTitle": value.AlbumTitle,
	})

	for _, inputFile := range audioFiles {
		if err := convertSingleFile(ctx, inputFile, mp3OutputDir, baseMetaData); err != nil {
			return fmt.Errorf("ファイル変換に失敗: %w", err)
		}
		logger.LogMessage(fmt.Sprintf("[%s] のファイル [%s] のMP3変換が完了", key, path.Base(inputFile)))
		logger.LogDebugEvent("mp3_conversion_completed", map[string]interface{}{
			"key":       key,
			"file":      path.Base(inputFile),
			"inputPath": inputFile,
		})
	}

	return nil
}

// prepareOutputDirectory は出力ディレクトリの準備を行います。
// ディレクトリが存在する場合はクリーンアップを、存在しない場合は作成を行います。
func prepareOutputDirectory(dir string) error {
	exists, err := audioconverter.DirExists(dir)
	if err != nil {
		return fmt.Errorf("ディレクトリの確認に失敗: %w", err)
	}

	if exists {
		if err := audioconverter.CleanUp(dir); err != nil {
			return fmt.Errorf("ディレクトリのクリーンアップに失敗: %w", err)
		}
	} else {
		if err := audioconverter.EnsureDirExists(dir); err != nil {
			return fmt.Errorf("ディレクトリの作成に失敗: %w", err)
		}
	}

	return nil
}

// convertSingleFile は単一の音声ファイルをMP3に変換します。
// 出力パスの生成、メタデータの設定、ファイルの変換を行います。
func convertSingleFile(ctx context.Context, inputFile, outputDir string, baseMetaData audioconverter.MP3Metadata) error {
	logger.LogDebugEvent("convertSingleFile_called", map[string]interface{}{
		"inputFile":  inputFile,
		"outputDir":  outputDir,
		"artist":     baseMetaData.Artist,
		"albumTitle": baseMetaData.AlbumTitle,
		"coverImage": baseMetaData.CoverImage,
	})

	name := path.Base(inputFile)
	nameWithoutExt := strings.TrimSuffix(name, filepath.Ext(name))
	mp3OutputPath := filepath.Join(outputDir, nameWithoutExt+mp3Extension)

	metaData := baseMetaData
	metaData.TrackName = nameWithoutExt

	if err := audioconverter.ConvertFileToMp3WithContext(ctx, inputFile, mp3OutputPath, metaData); err != nil {
		return fmt.Errorf("MP3変換に失敗: %w", err)
	}

	return nil
}

// splitActorNames は声優名をカンマや中黒などの区切り文字で分割します。
func splitActorNames(actor string) []string {
	if actor == "" {
		return nil
	}

	delimiters := func(r rune) bool {
		switch r {
		case ',', '・', '／', '/', '、', '，':
			return true
		default:
			return false
		}
	}

	parts := strings.FieldsFunc(actor, delimiters)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// printResults は変換処理の結果をログに出力します。
// 処理対象外ファイルと画像不足ファイルの情報を表示します。
func printResults(cfg *config.Config, notApplicableData, missingImageData []string) {
	logger.LogDebugEvent("printResults_called", map[string]interface{}{
		"notApplicableData": notApplicableData,
		"missingImageData":  missingImageData,
	})

	if len(notApplicableData) > 0 {
		logger.LogWarnMessage(fmt.Sprintf("下記のファイルはMP3変換できませんでした。[%s]に対応するHTMLファイルが存在するか確認してください", cfg.DirSetting.HtmlDir))
		logger.LogWarnMessage(fmt.Sprintf("  %v", notApplicableData))
		logger.LogDebugEvent("mp3_conversion_not_applicable", map[string]interface{}{
			"files":    notApplicableData,
			"count":    len(notApplicableData),
			"html_dir": cfg.DirSetting.HtmlDir,
		})
	}

	if cfg.Setting.SetMainImage && len(missingImageData) > 0 {
		logger.LogWarnMessage(fmt.Sprintf("下記のファイルはメイン画像が見つからず、MP3変換処理を実行できませんでした。[%s]配下の画像ファイルを確認してください", cfg.DirSetting.ImageDir))
		logger.LogWarnMessage(fmt.Sprintf("  %v", missingImageData))
		logger.LogDebugEvent("main_image_missing", map[string]interface{}{
			"files":     missingImageData,
			"count":     len(missingImageData),
			"image_dir": cfg.DirSetting.ImageDir,
		})
	}
}

package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// LogLevel はログレベルを定義
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger インターフェースでログ機能を統一
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

// StandardLogger は Logger インターフェースの標準実装
type StandardLogger struct {
	consoleLogger *log.Logger
	fileLogger    *log.Logger
	level         LogLevel
	debugEnabled  bool
	mu            sync.RWMutex // スレッドセーフ化のためのmutex
}

// NewStandardLogger は新しい StandardLogger を作成
func NewStandardLogger(level LogLevel, debugEnabled bool) *StandardLogger {
	return &StandardLogger{
		consoleLogger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
		level:         level,
		debugEnabled:  debugEnabled,
	}
}

// パッケージレベルのデフォルトロガー
var (
	defaultLogger Logger       = NewStandardLogger(INFO, false)
	defaultMu     sync.RWMutex // デフォルトロガーアクセス用mutex
)

// SetDefaultLogger はデフォルトロガーを設定
func SetDefaultLogger(logger Logger) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultLogger = logger
}

// Init はログファイルとコンソールの両方にログ出力を設定します。
func Init(logFilePath string, debug bool) (*os.File, error) {
	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, err
	}

	// デフォルトロガーの設定を更新（標準ライブラリのlogは使用しない）
	// デバッグモードが有効な場合はDEBUGレベルを、そうでなければINFOレベルを設定
	level := INFO
	if debug {
		level = DEBUG
	}

	defaultMu.Lock()
	defaultLogger = NewStandardLogger(level, debug)
	if sl, ok := defaultLogger.(*StandardLogger); ok {
		sl.SetFileOutput(logFile)
	}
	defaultMu.Unlock()

	return logFile, nil
}

// SetFileOutput はファイル出力を設定
func (l *StandardLogger) SetFileOutput(logFile *os.File) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fileLogger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
}

// Debug はデバッグレベルのログを出力
func (l *StandardLogger) Debug(args ...interface{}) {
	l.mu.RLock()
	enabled := l.debugEnabled && l.level <= DEBUG
	l.mu.RUnlock()

	if enabled {
		l.output("DEBUG", fmt.Sprint(args...))
	}
}

// Info は情報レベルのログを出力
func (l *StandardLogger) Info(args ...interface{}) {
	l.mu.RLock()
	enabled := l.level <= INFO
	l.mu.RUnlock()

	if enabled {
		l.output("INFO", fmt.Sprint(args...))
	}
}

// Warn は警告レベルのログを出力
func (l *StandardLogger) Warn(args ...interface{}) {
	l.mu.RLock()
	enabled := l.level <= WARN
	l.mu.RUnlock()

	if enabled {
		l.output("WARN", fmt.Sprint(args...))
	}
}

// Error はエラーレベルのログを出力
func (l *StandardLogger) Error(args ...interface{}) {
	l.mu.RLock()
	enabled := l.level <= ERROR
	l.mu.RUnlock()

	if enabled {
		l.output("ERROR", fmt.Sprint(args...))
	}
}

// Fatal は致命的エラーレベルのログを出力し、プログラムを終了
func (l *StandardLogger) Fatal(args ...interface{}) {
	l.output("FATAL", fmt.Sprint(args...))
	os.Exit(1)
}

// output は実際のログ出力を行う内部メソッド
func (l *StandardLogger) output(level, message string) {
	l.outputWithDepth(level, message, 5)
}

// outputWithDepth は指定された深度でログ出力を行う内部メソッド
func (l *StandardLogger) outputWithDepth(level, message string, depth int) {
	logMessage := fmt.Sprintf("[%s] %s", level, message)

	l.mu.RLock()
	consoleLogger := l.consoleLogger
	fileLogger := l.fileLogger
	l.mu.RUnlock()

	// コンソール出力
	if consoleLogger != nil {
		if err := consoleLogger.Output(depth, logMessage); err != nil {
			// エラーが発生してもログ出力は継続する
			fmt.Fprintf(os.Stderr, "コンソールログ出力エラー: %v\n", err)
		}
	}

	// ファイル出力
	if fileLogger != nil {
		if err := fileLogger.Output(depth, logMessage); err != nil {
			// エラーが発生してもログ出力は継続する
			fmt.Fprintf(os.Stderr, "ファイルログ出力エラー: %v\n", err)
		}
	}
}

// ==============================================
// パッケージレベルのAPI関数
// ==============================================

// Debug はデバッグレベルのログを出力
func Debug(args ...interface{}) {
	defaultMu.RLock()
	logger := defaultLogger
	defaultMu.RUnlock()
	logger.Debug(args...)
}

// Info は情報レベルのログを出力
func Info(args ...interface{}) {
	defaultMu.RLock()
	logger := defaultLogger
	defaultMu.RUnlock()
	logger.Info(args...)
}

// Warn は警告レベルのログを出力
func Warn(args ...interface{}) {
	defaultMu.RLock()
	logger := defaultLogger
	defaultMu.RUnlock()
	logger.Warn(args...)
}

// Error はエラーレベルのログを出力
func Error(args ...interface{}) {
	defaultMu.RLock()
	logger := defaultLogger
	defaultMu.RUnlock()
	logger.Error(args...)
}

// Fatal は致命的エラーレベルのログを出力し、プログラムを終了
func Fatal(args ...interface{}) {
	defaultMu.RLock()
	logger := defaultLogger
	defaultMu.RUnlock()
	logger.Fatal(args...)
}

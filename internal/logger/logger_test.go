package logger

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	// デバッグモードを有効にしてロガーを初期化
	logFile, err := Init(logPath, true)
	if err != nil {
		t.Fatalf("ロガーの初期化に失敗: %v", err)
	}
	defer logFile.Close()

	// ファイルが作成されたことを確認
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("ログファイルが作成されていません")
	}

	// デフォルトロガーがStandardLoggerであることを確認
	if _, ok := defaultLogger.(*StandardLogger); !ok {
		t.Error("デフォルトロガーがStandardLoggerに設定されていません")
	}
}

func TestLoggingOutput(t *testing.T) {
	// 新しいテスト用ロガーを作成
	var buf bytes.Buffer
	testLogger := NewStandardLogger(INFO, false)
	testLogger.consoleLogger = log.New(&buf, "", 0)

	// デフォルトロガーを一時的にテスト用に変更
	oldLogger := defaultLogger
	defaultLogger = testLogger
	defer func() { defaultLogger = oldLogger }()

	testCases := []struct {
		name     string
		logFunc  func(...interface{})
		message  string
		expected string
	}{
		{"info", info, "info message", "[INFO] info message"},
		{"errorLog", errorLog, "error message", "[ERROR] error message"},
		{"warn", warn, "warn message", "[WARN] warn message"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			tc.logFunc(tc.message)
			output := buf.String()
			if !strings.Contains(output, tc.expected) {
				t.Errorf("%s: 期待される出力が含まれていません。got: %q, want: %q", tc.name, output, tc.expected)
			}
		})
	}
}

func TestDebugOutput(t *testing.T) {
	var buf bytes.Buffer

	// デバッグモードが無効の場合
	testLoggerNoDebug := NewStandardLogger(INFO, false)
	testLoggerNoDebug.consoleLogger = log.New(&buf, "", 0)
	oldLogger := defaultLogger
	defaultLogger = testLoggerNoDebug

	debug("debug message")
	if buf.String() != "" {
		t.Error("デバッグモードが無効の場合、出力されるべきではありません")
	}

	// デバッグモードが有効の場合
	buf.Reset()
	testLoggerWithDebug := NewStandardLogger(DEBUG, true)
	testLoggerWithDebug.consoleLogger = log.New(&buf, "", 0)
	defaultLogger = testLoggerWithDebug

	debug("debug message")
	if !strings.Contains(buf.String(), "[DEBUG] debug message") {
		t.Errorf("デバッグメッセージが出力されていません。output: %q", buf.String())
	}

	// デフォルトロガーを復元
	defaultLogger = oldLogger
}

func TestFileOutput(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	// ロガーを初期化
	logFile, err := Init(logPath, false)
	if err != nil {
		t.Fatalf("ロガーの初期化に失敗: %v", err)
	}
	defer logFile.Close()

	// テストメッセージを出力
	testMessage := "test file output"
	info(testMessage)

	// ファイルの内容を読み取り
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ログファイルの読み取りに失敗: %v", err)
	}

	// 出力を検証
	if !strings.Contains(string(content), testMessage) {
		t.Error("ファイル出力にメッセージが含まれていません")
	}
}

func TestFatal(t *testing.T) {
	if os.Getenv("TEST_FATAL") == "1" {
		fatal("fatal error")
		return
	}
	// 通常のテストでは、Fatalが呼び出されないことを確認するだけです
	// Fatalは実際にプロセスを終了させるため、単体テストでは呼び出しません
}

func TestStructuredLogging(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewStandardLogger(DEBUG, true)
	testLogger.consoleLogger = log.New(&buf, "", 0)

	oldLogger := defaultLogger
	defaultLogger = testLogger
	defer func() { defaultLogger = oldLogger }()

	testCases := []struct {
		name     string
		logFunc  func(string, map[string]interface{})
		logLevel string
	}{
		{"LogDebugEvent", LogDebugEvent, "DEBUG"},
		{"LogInfoEvent", LogInfoEvent, "INFO"},
		{"LogWarnEvent", LogWarnEvent, "WARN"},
		{"LogErrorEvent", LogErrorEvent, "ERROR"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			tc.logFunc("test_event", map[string]interface{}{
				"key":     "value",
				"numeric": 123,
			})
			output := buf.String()

			// ログレベルの確認
			if !strings.Contains(output, "["+tc.logLevel+"]") {
				t.Errorf("%s: ログレベルが含まれていません。output: %q", tc.name, output)
			}

			// JSONの確認
			if !strings.Contains(output, `"event":"test_event"`) {
				t.Errorf("%s: イベント名が含まれていません。output: %q", tc.name, output)
			}
			if !strings.Contains(output, `"key":"value"`) {
				t.Errorf("%s: キーが含まれていません。output: %q", tc.name, output)
			}
			if !strings.Contains(output, `"numeric":123`) {
				t.Errorf("%s: 数値が含まれていません。output: %q", tc.name, output)
			}
		})
	}
}

func TestInternalLogFunctions(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewStandardLogger(DEBUG, true)
	testLogger.consoleLogger = log.New(&buf, "", 0)

	oldLogger := defaultLogger
	defaultLogger = testLogger
	defer func() { defaultLogger = oldLogger }()

	testCases := []struct {
		name     string
		logFunc  func(...interface{})
		message  string
		expected string
	}{
		{"debug", debug, "debug message", "[DEBUG] debug message"},
		{"info", info, "info message", "[INFO] info message"},
		{"warn", warn, "warn message", "[WARN] warn message"},
		{"errorLog", errorLog, "error message", "[ERROR] error message"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			tc.logFunc(tc.message)
			output := buf.String()
			if !strings.Contains(output, tc.expected) {
				t.Errorf("%s: 期待される出力が含まれていません。got: %q, want: %q", tc.name, output, tc.expected)
			}
		})
	}
}

func TestSimpleMessageLogging(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewStandardLogger(INFO, false)
	testLogger.consoleLogger = log.New(&buf, "", 0)

	oldLogger := defaultLogger
	defaultLogger = testLogger
	defer func() { defaultLogger = oldLogger }()

	testCases := []struct {
		name     string
		logFunc  func(string)
		message  string
		expected string
	}{
		{"LogMessage", LogMessage, "info message", "[INFO] info message"},
		{"LogWarnMessage", LogWarnMessage, "warn message", "[WARN] warn message"},
		{"LogErrorMessage", LogErrorMessage, "error message", "[ERROR] error message"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			tc.logFunc(tc.message)
			output := buf.String()
			if !strings.Contains(output, tc.expected) {
				t.Errorf("%s: 期待される出力が含まれていません。got: %q, want: %q", tc.name, output, tc.expected)
			}

			// JSONが含まれていないことを確認
			if strings.Contains(output, "{") || strings.Contains(output, "}") {
				t.Errorf("%s: シンプルメッセージにJSON形式が含まれています。output: %q", tc.name, output)
			}
		})
	}
}

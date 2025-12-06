package logger

import (
	"encoding/json"
	"fmt"
)

// LogDebugEvent はデバッグ情報をJSON形式でログ出力します。
// eventName: ログイベントの名前
// data: ログに出力するデータのマップ
func LogDebugEvent(eventName string, data map[string]interface{}) {
	data["event"] = eventName
	logJson, err := json.Marshal(data)
	if err != nil {
		debug(fmt.Sprintf("JSON marshal error for event %s: %v", eventName, err))
		return
	}
	debug(string(logJson))
}

// LogInfoEvent は情報レベルのイベントをJSON形式でログ出力します。
// eventName: ログイベントの名前
// data: ログに出力するデータのマップ
func LogInfoEvent(eventName string, data map[string]interface{}) {
	data["event"] = eventName
	logJson, err := json.Marshal(data)
	if err != nil {
		warn(fmt.Sprintf("JSON marshal error for event %s: %v", eventName, err))
		return
	}
	info(string(logJson))
}

// LogWarnEvent は警告レベルのイベントをJSON形式でログ出力します。
// eventName: ログイベントの名前
// data: ログに出力するデータのマップ
func LogWarnEvent(eventName string, data map[string]interface{}) {
	data["event"] = eventName
	logJson, err := json.Marshal(data)
	if err != nil {
		warn(fmt.Sprintf("JSON marshal error for event %s: %v", eventName, err))
		return
	}
	warn(string(logJson))
}

// LogErrorEvent はエラーレベルのイベントをJSON形式でログ出力します。
// eventName: ログイベントの名前
// data: ログに出力するデータのマップ
func LogErrorEvent(eventName string, data map[string]interface{}) {
	data["event"] = eventName
	logJson, err := json.Marshal(data)
	if err != nil {
		errorLog(fmt.Sprintf("JSON marshal error for event %s: %v", eventName, err))
		return
	}
	errorLog(string(logJson))
}

// LogFatalEvent は致命的エラーレベルのイベントをJSON形式でログ出力し、プログラムを終了します。
// eventName: ログイベントの名前
// data: ログに出力するデータのマップ
func LogFatalEvent(eventName string, data map[string]interface{}) {
	data["event"] = eventName
	logJson, err := json.Marshal(data)
	if err != nil {
		fatal(fmt.Sprintf("JSON marshal error for event %s: %v", eventName, err))
		return
	}
	fatal(string(logJson))
}

// ==============================================
// シンプルなメッセージログAPI（ユーザー向け）
// ==============================================

// LogMessage は情報レベルのシンプルなメッセージをログ出力します。
// ユーザーに伝えたい情報を簡潔に出力する場合に使用します。
func LogMessage(message string) {
	info(message)
}

// LogWarnMessage は警告レベルのシンプルなメッセージをログ出力します。
// ユーザーに警告を伝える場合に使用します。
func LogWarnMessage(message string) {
	warn(message)
}

// LogErrorMessage はエラーレベルのシンプルなメッセージをログ出力します。
// ユーザーにエラーを伝える場合に使用します。
func LogErrorMessage(message string) {
	errorLog(message)
}

// LogFatalMessage は致命的エラーのシンプルなメッセージをログ出力し、プログラムを終了します。
// ユーザーに致命的エラーを伝えてプログラムを終了する場合に使用します。
func LogFatalMessage(message string) {
	fatal(message)
}

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
		Debug(fmt.Sprintf("JSON marshal error for event %s: %v", eventName, err))
		return
	}
	Debug(string(logJson))
}

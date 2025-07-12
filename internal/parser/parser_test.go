package parser

import (
	"os"
	"testing"
)

func TestParseHTML(t *testing.T) {
	// テスト用のHTML文字列
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
	<title>テスト作品</title>
</head>
<body>
	<h1 id="work_name">テストアルバム</h1>
	<div>
		<span itemprop="brand" class="maker_name"><a>テストブランド</a></span>
	</div>
	<table id="work_outline">
		<tr>
			<th>声優</th>
			<td><a>テスト声優A</a>, <a>テスト声優B</a></td>
		</tr>
		<tr>
			<th>ジャンル</th>
			<td><div>テストジャンル1</div><div>テストジャンル2</div></td>
		</tr>
	</table>
	<div class="work_parts type_tracklist">
		<div class="work_parts_heading">【トラックリスト】</div>
		<div class="work_tracklist">
			<div class="work_tracklist_item">
				<p class="title">トラック1</p>
				<p class="time">3:45</p>
			</div>
			<div class="work_tracklist_item">
				<p class="title">トラック2</p>
				<p class="time">4:20</p>
			</div>
		</div>
	</div>
</body>
</html>`

	// HTMLを解析
	data, err := parseHTML(htmlContent)
	if err != nil {
		t.Fatalf("HTML解析エラー: %v", err)
	}

	// 期待される結果の検証
	expectedValues := map[string]string{
		"アルバムタイトル": "テストアルバム",
		"サークル名":    "テストブランド",
		"声優":       "テスト声優A, テスト声優B",
		"ジャンル":     "テストジャンル1, テストジャンル2",
		"トラックリスト":  "トラック1 (3:45), トラック2 (4:20)",
	}

	for key, expected := range expectedValues {
		if actual, exists := data[key]; !exists {
			t.Errorf("キー %q が見つかりません", key)
		} else if actual != expected {
			t.Errorf("%s: got %q, want %q", key, actual, expected)
		}
	}
}

func TestExtractHtml(t *testing.T) {
	// テスト用の一時ファイルを作成
	tempDir := t.TempDir()
	htmlFilePath := tempDir + "/test.html"

	// テスト用のHTML内容
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
	<title>テスト作品</title>
</head>
<body>
	<h1 id="work_name">テストアルバム</h1>
	<div>
		<span itemprop="brand" class="maker_name"><a>テストブランド</a></span>
	</div>
	<table id="work_outline">
		<tr>
			<th>声優</th>
			<td><a>テスト声優</a></td>
		</tr>
	</table>

	<div class="work_parts type_tracklist">
		<h3 class="work_parts_heading">収録内容</h3>
		<div class="work_parts_area">
			<ul class="work_tracklist">
			<li class="work_tracklist_item">
				<div class="title">トラック1</div>
				<div class="time">01:00</div>
			</li>
			<li class="work_tracklist_item">
				<div class="title">トラック2</div>
				<div class="time">02:00</div>
			</li>
			<li class="work_tracklist_item">
				<div class="title">トラック3</div>
				<div class="time">03:00</div>
			</li>
			</ul>
		</div>
	</div>

	<div class="work_main_image">
		<img src="test.jpg" alt="メイン画像">
	</div>
</body>
</html>`

	// ファイルに書き込み
	if err := os.WriteFile(htmlFilePath, []byte(htmlContent), 0644); err != nil {
		t.Fatalf("テストファイルの作成に失敗: %v", err)
	}

	// まず、parseHTMLの結果を確認
	parsedData, err := parseHTML(htmlContent)
	if err != nil {
		t.Fatalf("HTML解析エラー: %v", err)
	}

	// トラックリストの形式を確認
	expectedTrackListStr := "トラック1 (01:00), トラック2 (02:00), トラック3 (03:00)"
	if trackList, exists := parsedData["トラックリスト"]; !exists {
		t.Fatalf("トラックリストが見つかりません")
	} else if trackList != expectedTrackListStr {
		t.Errorf("トラックリスト文字列: got %q, want %q", trackList, expectedTrackListStr)
	}

	// ExtractHtmlを実行
	result, err := ExtractHtml(htmlFilePath)
	if err != nil {
		t.Fatalf("ExtractHtmlの実行に失敗: %v", err)
	}

	// 基本情報の検証
	if result.AlbumTitle != "テストアルバム" {
		t.Errorf("AlbumTitle: got %q, want %q", result.AlbumTitle, "テストアルバム")
	}
	if result.Brand != "テストブランド" {
		t.Errorf("Brand: got %q, want %q", result.Brand, "テストブランド")
	}
	if result.Actor != "テスト声優" {
		t.Errorf("Actor: got %q, want %q", result.Actor, "テスト声優")
	}

	// トラックリストの検証
	if len(result.TrackList) != 3 {
		t.Errorf("TrackList length: got %d, want 3", len(result.TrackList))
		return
	}

	expectedTrack := struct {
		title    string
		duration string
	}{
		title:    "トラック1",
		duration: "1分0秒",
	}

	track := result.TrackList[0]
	if track.TrackTitle != expectedTrack.title {
		t.Errorf("Track title: got %q, want %q", track.TrackTitle, expectedTrack.title)
	}
	if track.TrackDuration != expectedTrack.duration {
		t.Errorf("Track duration: got %q, want %q", track.TrackDuration, expectedTrack.duration)
	}
}

func TestExtractHtml_FileNotFound(t *testing.T) {
	_, err := ExtractHtml("non_existent_file.html")
	if err == nil {
		t.Error("存在しないファイルでエラーが発生すべき")
	}
}

func TestParseHTML_InvalidHTML(t *testing.T) {
	invalidHTML := "This is not valid HTML"
	_, err := parseHTML(invalidHTML)
	if err != nil {
		t.Logf("期待通り、無効なHTMLでエラーが発生: %v", err)
	}
}

package model

// Track はトラック情報を格納する構造体です。
type Track struct {
	TrackTitle    string `json:"track_title"`    // トラックタイトル
	TrackDuration string `json:"track_duration"` // 再生時間
}

// IndividualData は個別の作品データを格納する構造体です。
type IndividualData struct {
	AlbumTitle string            `json:"album_title"` // アルバムタイトル
	Actor      string            `json:"actor"`       // 声優名
	Brand      string            `json:"brand"`       // ブランド名
	MainImage  string            `json:"main_image"`  // メイン画像のパス
	TrackList  []Track           `json:"track_list"`  // トラック一覧
	Additional map[string]string `json:"additional"`  // 追加情報
}

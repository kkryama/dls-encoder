package generator

import (
	"bytes"
	"html/template"
)

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.AlbumTitle}}</title>
</head>
<body>
    <h1 id="work_name">{{.AlbumTitle}}</h1>
    
    <div class="maker_name">
        <span itemprop="brand" class="maker_name"><a href="#">{{.BrandName}}</a></span>
    </div>

    <table id="work_outline">
        {{range $key, $value := .Details}}
        <tr>
            <th>{{$key}}</th>
            <td>{{$value}}</td>
        </tr>
        {{end}}
    </table>

    <div class="work_parts type_tracklist">
        <div class="work_parts_heading">【収録内容】</div>
        {{range .Tracks}}
        <div class="work_tracklist_item">
            <span class="title">{{.Title}}</span>
            <span class="time">{{.Duration}}</span>
        </div>
        {{end}}
    </div>
</body>
</html>`

const parseDTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.AlbumTitle}}</title>
</head>
<body>
    <h1 class="productTitle__txt">{{.AlbumTitle}}</h1>
    
    <a class="circleName__txt">{{.BrandName}}</a>

    <table id="work_outline">
        {{range $key, $value := .Details}}
        <tr>
            <th>{{$key}}</th>
            <td>{{$value}}</td>
        </tr>
        {{end}}
    </table>

    <!-- parseDタイプではトラックリストを省略 -->
</body>
</html>`

// Track はトラック情報を格納する構造体です。
type Track struct {
	Title    string // トラックタイトル
	Duration string // 再生時間
}

// TemplateData はHTMLテンプレート生成に使用するデータ構造です。
type TemplateData struct {
	AlbumTitle string            // アルバムタイトル
	BrandName  string            // ブランド名
	Details    map[string]string // 詳細情報のキーバリューペア
	Tracks     []Track           // トラック一覧
	IsParseD   bool              // parseDタイプかどうか
}

// GenerateHTML はテンプレートデータからHTMLを生成します。
func GenerateHTML(data *TemplateData) (string, error) {
	var tmplStr string
	if data.IsParseD {
		tmplStr = parseDTemplate
	} else {
		tmplStr = htmlTemplate
	}

	tmpl, err := template.New("album").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

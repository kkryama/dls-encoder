package generator

import (
	"strings"
	"testing"
)

func TestGenerateHTML(t *testing.T) {
	tests := []struct {
		name    string
		data    *TemplateData
		want    []string
		wantErr bool
	}{
		{
			name: "基本的なデータでHTMLを生成",
			data: &TemplateData{
				AlbumTitle: "テストアルバム",
				BrandName:  "テストサークル",
				Details: map[string]string{
					"ジャンル": "テスト",
					"作者":   "テスト作者",
				},
				Tracks: []Track{
					{Title: "トラック1", Duration: "3:30"},
					{Title: "トラック2", Duration: "4:15"},
				},
			},
			want: []string{
				`<h1 id="work_name">テストアルバム</h1>`,
				`<span itemprop="brand" class="maker_name"><a href="#">テストサークル</a></span>`,
				`<th>ジャンル</th>`,
				`<td>テスト</td>`,
				`<th>作者</th>`,
				`<td>テスト作者</td>`,
				`<span class="title">トラック1</span>`,
				`<span class="time">3:30</span>`,
				`<span class="title">トラック2</span>`,
				`<span class="time">4:15</span>`,
			},
			wantErr: false,
		},
		{
			name: "最小限のデータでHTMLを生成",
			data: &TemplateData{
				AlbumTitle: "最小限アルバム",
				BrandName:  "最小限サークル",
				Details:    map[string]string{},
				Tracks:     []Track{},
			},
			want: []string{
				`<h1 id="work_name">最小限アルバム</h1>`,
				`<span itemprop="brand" class="maker_name"><a href="#">最小限サークル</a></span>`,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateHTML(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("GenerateHTML() = %v, want %v", got, want)
				}
			}
		})
	}
}

package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// parseHTML はHTMLコンテンツを解析し、必要な情報を抽出してマップで返します。
func parseHTML(htmlContent string) (map[string]string, error) {
	data := make(map[string]string)

	// goquery で HTML を解析
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("HTMLの解析に失敗しました: %v", err)
	}

	// タイトルとして h1 要素の work_name を取得して data に追加
	productName := doc.Find("h1#work_name").Text()
	data["アルバムタイトル"] = productName

	// サークル名 を取得する
	brandName := strings.TrimSpace(doc.Find("span[itemprop='brand'].maker_name a").Text())
	data["サークル名"] = brandName

	// `#work_outline` テーブルの `tr` をループし 概要 を取得する
	doc.Find("#work_outline tr").Each(func(i int, s *goquery.Selection) {
		th := strings.TrimSpace(s.Find("th").Text())
		td := s.Find("td")

		if th == "" || td.Length() == 0 {
			return
		}

		var values []string
		td.Find("a, div").Each(func(i int, t *goquery.Selection) {
			text := strings.TrimSpace(t.Text())
			if text != "" {
				values = append(values, text)
			}
		})

		if len(values) == 0 {
			text := strings.TrimSpace(td.Text())
			if text != "" {
				values = append(values, text)
			}
		}

		// 1つの値がある場合、または複数の場合の格納
		if len(values) == 1 {
			data[th] = values[0]
		} else if len(values) > 1 {
			data[th] = strings.Join(values, ", ")
		}
	})

	// 収録内容（work_parts type_tracklist）を取得
	doc.Find(".work_parts.type_tracklist").Each(func(i int, s *goquery.Selection) {
		// 見出し（【収録内容】）を取得
		heading := strings.TrimSpace(s.Find(".work_parts_heading").Text())
		if heading != "" {
			data["収録内容"] = heading
		}

		// 各トラックの情報を取得
		var trackList []string
		s.Find(".work_tracklist_item").Each(func(i int, item *goquery.Selection) {
			title := strings.TrimSpace(item.Find(".title").Text())
			time := strings.TrimSpace(item.Find(".time").Text())
			if title != "" && time != "" {
				trackList = append(trackList, fmt.Sprintf("%s (%s)", title, time))
			}
		})

		// 収録内容を1つの文字列として格納
		if len(trackList) > 0 {
			data["トラックリスト"] = strings.Join(trackList, ", ")
		}
	})

	return data, nil
}

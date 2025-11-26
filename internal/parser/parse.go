package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// parseRJ は RJxxxxxxxx のHTMLを解析します。
func parseRJ(htmlContent string) (map[string]string, error) {
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
			separator := ", "
			if th == "声優" {
				separator = "・"
			}
			data[th] = strings.Join(values, separator)
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

// parseD は d_xxxxxx のHTMLを解析します。
func parseD(htmlContent string) (map[string]string, error) {
	data := make(map[string]string)

	// goquery で HTML を解析
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("HTMLの解析に失敗しました: %v", err)
	}

	// タイトルを取得（優先: CSSセレクタ）
	titleSelection := doc.Find("h1.productTitle__txt")
	if titleSelection.Length() > 0 {
		// キャンペーン情報を除去
		titleSelection.Find("span.productTitle__txt--campaign").Remove()
		productTitle := strings.TrimSpace(titleSelection.Text())
		if productTitle != "" {
			data["アルバムタイトル"] = productTitle
		}
	}
	if data["アルバムタイトル"] == "" {
		// フォールバック: タイトルタグから
		title := strings.TrimSpace(doc.Find("title").First().Text())
		if title == "" {
			// og:title を確認
			title = strings.TrimSpace(doc.Find("meta[property='og:title']").AttrOr("content", ""))
		}
		if title != "" && strings.Contains(title, "同人") {
			titlePart := strings.Split(title, "｜")[0]
			re1 := regexp.MustCompile(`【([^】]+)】([^【]+)【([^】]+)】\(([^)]+)\)`)
			if matches := re1.FindStringSubmatch(titlePart); len(matches) >= 5 {
				data["アルバムタイトル"] = strings.TrimSpace(matches[1] + matches[2])
				data["声優"] = strings.TrimSpace(matches[3])
				data["サークル名"] = strings.TrimSpace(matches[4])
			} else {
				re2 := regexp.MustCompile(`(.+)\(([^)]+)\)`)
				if matches := re2.FindStringSubmatch(titlePart); len(matches) >= 3 {
					data["アルバムタイトル"] = strings.TrimSpace(matches[1])
					data["サークル名"] = strings.TrimSpace(matches[2])
				} else {
					data["アルバムタイトル"] = titlePart
				}
			}
		}
	}

	// サークル名を取得
	if data["サークル名"] == "" {
		brandName := strings.TrimSpace(doc.Find("a.circleName__txt").Text())
		if brandName != "" {
			data["サークル名"] = brandName
		}
	}

	// 声優を取得
	if data["声優"] == "" {
		doc.Find("div.productInformation__item dl.informationList").Each(func(i int, s *goquery.Selection) {
			dt := strings.TrimSpace(s.Find("dt.informationList__ttl").Text())
			if dt == "声優" {
				var actors []string
				s.Find("dd.informationList__txt a").Each(func(j int, t *goquery.Selection) {
					actor := strings.TrimSpace(t.Text())
					if actor != "" {
						actors = append(actors, actor)
					}
				})
				if len(actors) > 0 {
					if len(actors) == 1 {
						data["声優"] = actors[0]
					} else {
						data["声優"] = strings.Join(actors, "・")
					}
				}
			}
		})
	}

	// もし声優が取得できなかった場合、説明文からCVを抽出
	if data["声優"] == "" {
		doc.Find(".m-productSummary .summary").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			if strings.Contains(text, "CV") || strings.Contains(text, "声優") {
				lines := strings.Split(text, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.Contains(line, "CV") || strings.Contains(line, "声優") {
						// "シナリオ＆CV；:柚木つばめ" のような形式から抽出
						if strings.Contains(line, ":") {
							parts := strings.Split(line, ":")
							if len(parts) > 1 {
								actor := strings.TrimSpace(parts[1])
								// 余分な文字を除去
								actor = strings.TrimSpace(strings.Split(actor, " ")[0])
								data["声優"] = actor
								break
							}
						}
					}
				}
			}
		})
	}

	// メイン画像を取得
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists && strings.Contains(src, "main") {
			if strings.HasPrefix(src, "//") {
				src = "https:" + src
			}
			data["メイン画像"] = src
			return
		}
	})

	// og:image も確認
	if data["メイン画像"] == "" {
		ogImage := doc.Find("meta[property='og:image']").AttrOr("content", "")
		if ogImage != "" {
			data["メイン画像"] = ogImage
		}
	}

	// トラックリストを取得
	// NOTE: 現状ではトラックリストを活用できていないため省略

	// #work_outline テーブルの `tr` をループし 概要 を取得する
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
			separator := ", "
			if th == "声優" {
				separator = "・"
			}
			data[th] = strings.Join(values, separator)
		}
	})

	return data, nil
}

// parseHTML はローカルHTMLコンテンツを解析し、必要な情報を抽出してマップで返します。
func parseHTML(htmlContent string, dirName string) (map[string]string, error) {
	if strings.HasPrefix(dirName, "d_") {
		return parseD(htmlContent)
	}
	return parseRJ(htmlContent)
}

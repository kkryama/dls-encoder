package parser

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/kkryama/dls-encoder/internal/config"
	"github.com/kkryama/dls-encoder/internal/model"
)

// ExtractData はHTMLファイルからデータを抽出し、構造化されたデータを返します。
func ExtractData(targetHtmlFilePath, dirName string, cfg *config.Config) (model.IndividualData, error) {
	var result model.IndividualData

	var parsedHtml map[string]string
	var err error

	// ローカルHTMLを読む
	htmlContentBytes, err := os.ReadFile(targetHtmlFilePath)
	if err != nil {
		return result, fmt.Errorf("ファイルの読み込みに失敗しました: %v", err)
	}
	htmlContent := string(htmlContentBytes)
	parsedHtml, err = parseHTML(htmlContent, dirName)

	if err != nil {
		return result, fmt.Errorf("データの取得に失敗しました: %v", err)
	}

	// データを整理
	data := model.IndividualData{
		Additional: make(map[string]string),
	}
	for key, value := range parsedHtml {
		switch key {
		case "アルバムタイトル":
			data.AlbumTitle = value
		case "声優":
			data.Actor = value
		case "サークル名":
			data.Brand = value
		case "トラックリスト":
			// トラック情報を抽出するための正規表現
			re := regexp.MustCompile(`([^\s]+) \((\d+):(\d+)\)`)
			matches := re.FindAllStringSubmatch(value, -1)

			tracks := []model.Track{}
			for _, match := range matches {
				title := match[1]
				minutes, _ := strconv.Atoi(match[2])
				seconds, _ := strconv.Atoi(match[3])
				duration := time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second

				tracks = append(tracks, model.Track{
					TrackTitle:    title,
					TrackDuration: fmt.Sprintf("%d分%d秒", int(duration.Minutes()), int(duration.Seconds())%60),
				})
			}
			data.TrackList = tracks
		case "メイン画像":
			data.MainImage = value
		default:
			data.Additional[key] = value
		}
	}

	result = model.IndividualData{
		AlbumTitle: data.AlbumTitle,
		Actor:      data.Actor,
		Brand:      data.Brand,
		MainImage:  data.MainImage,
		TrackList:  data.TrackList,
		Additional: data.Additional,
	}
	return result, nil
}

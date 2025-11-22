package generator

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// InteractiveHTMLGenerator は対話型でHTMLファイルを生成します。
// ユーザーからの入力を受け取り、HTMLテンプレートを使用してファイルを作成します。
func InteractiveHTMLGenerator(htmlDir, imageDir string) error {
	reader := bufio.NewReader(os.Stdin)

	// ファイル名の入力
	fmt.Print("HTMLファイル名を入力してください（.htmlは自動で付加されます）: ")
	filename, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return fmt.Errorf("ファイル名が入力されていません")
	}

	// d_xxxxxx形式かどうかを判定
	isParseD := regexp.MustCompile(`^d_\d{6}$`).MatchString(filename)

	// アルバムタイトルの入力
	fmt.Print("アルバムタイトルを入力してください: ")
	albumTitle, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	albumTitle = strings.TrimSpace(albumTitle)

	// サークル名の入力
	fmt.Print("サークル名を入力してください: ")
	brandName, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	brandName = strings.TrimSpace(brandName)

	// 声優またはactorの入力
	details := make(map[string]string)
	var actorKey string
	if isParseD {
		actorKey = "声優"
	} else {
		actorKey = "actor"
	}
	fmt.Printf("%sを入力してください: ", actorKey)
	actor, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	actor = strings.TrimSpace(actor)
	if actor != "" {
		details[actorKey] = actor
	}

	// 詳細情報の入力
	for {
		fmt.Print("詳細情報を追加しますか？ (y/n): ")
		answer, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(answer)) != "y" {
			break
		}

		fmt.Print("項目名を入力してください: ")
		key, _ := reader.ReadString('\n')
		key = strings.TrimSpace(key)

		fmt.Print("値を入力してください: ")
		value, _ := reader.ReadString('\n')
		value = strings.TrimSpace(value)

		details[key] = value
	}

	// トラックリストの入力
	var tracks []Track
	if !isParseD {
		for {
			fmt.Print("トラックを追加しますか？ (y/n): ")
			answer, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(answer)) != "y" {
				break
			}

			fmt.Print("トラックタイトルを入力してください: ")
			title, _ := reader.ReadString('\n')
			title = strings.TrimSpace(title)

			fmt.Print("再生時間を入力してください (例: 4:30): ")
			duration, _ := reader.ReadString('\n')
			duration = strings.TrimSpace(duration)

			tracks = append(tracks, Track{
				Title:    title,
				Duration: duration,
			})
		}
	}

	// テンプレートデータの作成
	data := &TemplateData{
		AlbumTitle: albumTitle,
		BrandName:  brandName,
		Details:    details,
		Tracks:     tracks,
		IsParseD:   isParseD,
	}

	// HTMLの生成
	html, err := GenerateHTML(data)
	if err != nil {
		return err
	}

	// HTMLファイルの保存
	filePath := filepath.Join(htmlDir, filename+".html")

	// 既存のファイルチェック
	if _, err := os.Stat(filePath); err == nil {
		fmt.Print("同名のファイルが既に存在します。上書きしますか？ (y/n): ")
		answer, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(answer)) != "y" {
			return fmt.Errorf("ファイルの作成を中止しました")
		}
	}

	// HTMLファイルの保存
	if err := os.WriteFile(filePath, []byte(html), 0644); err != nil {
		return err
	}

	fmt.Printf("HTMLファイルを保存しました: %s\n", filePath)

	// メイン画像に関する情報を表示
	fmt.Println("\n【メイン画像について】")
	fmt.Printf("メイン画像を設定する場合は、以下の場所に画像ファイルを配置してください：\n")
	fmt.Printf("  配置ディレクトリ: %s\n", imageDir)
	fmt.Printf("  ファイル名: %s.webp または %s.jpg\n", filename, filename)
	fmt.Printf("  (webp形式が優先されます)\n")

	return nil
}

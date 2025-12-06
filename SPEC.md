# dls-encoder 仕様書 (SPEC.md)

## 概要

dls-encoder は、RJxxxxxxxx または d_xxxxxx 形式のディレクトリからダウンロードしたコンテンツのエンコーダーです。FFmpeg を利用して WAV/FLAC/MP3 ファイルを MP3 形式に変換し、同名の HTML ファイルからメタデータを自動的に設定します。音楽ライブラリやデータベースのメタデータ管理を簡素化し、特に個人用の音楽整理に役立ちます。

## 機能仕様

### 1. 音声変換機能
- **入力形式**: WAV, FLAC, MP3 ファイル
- **出力形式**: MP3 (固定)
- **エンコーディング設定**:
  - ビットレート: 320kbps
  - サンプリングレート: 48kHz
  - エンコーダー: libmp3lame
  - ID3 バージョン: 2.3
- **優先順位**: 同じディレクトリに複数拡張子が存在する場合、WAV > FLAC > MP3 の優先度で1つのみを採用
- **除外ファイル**: 設定ファイルで指定した除外文字列を**ファイルパス全体**に含むファイルは自動的に除外（デフォルト: "SE無し", "SEなし", "効果音無し", "効果音なし", "_MACOSX"）。除外判定はファイル名だけでなく、ディレクトリ名を含むパス全体に対して行われます。`_MACOSX` を指定すると `__MACOSX` ディレクトリにも部分一致でマッチします。

### 2. メタデータ自動設定機能
- **メタデータソース**: 同名の HTML ファイル
- **設定されるメタデータ**:
  - Artist (声優名)
  - Album Artist (サークル名)
  - Album (アルバムタイトル)
  - Title (トラック名、ファイル名から拡張子を除いたもの)
  - Cover Image (メイン画像、設定により)
- **声優名の処理**: 複数の声優がいる場合、以下の区切り文字で自動分割されます
  - カンマ: `,` `，`
  - 中黒: `・`
  - スラッシュ: `/` `／`
  - 読点: `、`
  - 分割後、出力ディレクトリ名では先頭2名のみを「・」区切りで使用し、3名以上の場合は末尾に「他」を付与

#### HTMLパース結果とMP3メタデータの対応表

| HTMLパース結果 | IndividualDataフィールド | MP3メタデータ | 説明 |
|---------------|-------------------------|---------------|------|
| アルバムタイトル | AlbumTitle | AlbumTitle | 作品のタイトル |
| 声優 | Actor | Artist | 声優/アーティスト名 |
| サークル名 | Brand | AlbumArtist | サークル/ブランド名 |
| メイン画像 | MainImage | CoverImage | アルバムカバー画像 |
| トラックリスト | TrackList | - | トラック情報（JSON保存用） |
| その他 | Additional | - | 追加情報（ジャンルなど） |

#### MP3メタデータの設定方法
IndividualData から MP3Metadata への変換は以下の通りです：

```go
baseMetaData := audioconverter.MP3Metadata{
    Artist:      value.Actor,       // 声優名
    AlbumArtist: value.Brand,       // サークル名
    AlbumTitle:  value.AlbumTitle,  // アルバムタイトル
    CoverImage:  coverImage,        // メイン画像のパス（設定によりnil）
}
```

各音声ファイルに対して TrackName を設定：
```go
metaData := baseMetaData
metaData.TrackName = nameWithoutExt  // ファイル名から拡張子を除いたもの
```

FFmpeg コマンドでのメタデータ設定：
- `-metadata artist=<Artist>`
- `-metadata album_artist=<AlbumArtist>`
- `-metadata album=<AlbumTitle>`
- `-metadata title=<TrackName>`
- `-id3v2_version 3` (ID3v2.3 を使用)

画像埋め込みの場合：
- 追加入力: `-i <CoverImage>`
- マッピング: `-map 0:a -map 1:v`
- 画像エンコード: `-c:v mjpeg`
- 画像メタデータ: `-metadata:s:v title=Album cover`

#### サイト別パース仕様

##### RJxxxxxxxx パース仕様
- **アルバムタイトル**: `h1#work_name` のテキスト
- **サークル名**: `span[itemprop='brand'].maker_name a` のテキスト
- **メイン画像**: `.product-slider-data div[data-src]` の `data-src` 属性
- **トラックリスト**: `.work_parts.type_tracklist .work_tracklist_item` のタイトルと時間
- **その他情報**: `#work_outline tr` の th/td ペア

##### d_xxxxxx パース仕様
- **アルバムタイトル**: 
  1. `h1.productTitle__txt` のテキスト（`span.productTitle__txt--campaign` 要素は除去）
  2. フォールバック: `title` タグまたは `meta[property='og:title']` から正規表現で抽出
- **サークル名**: 
  1. `a.circleName__txt` のテキスト
  2. フォールバック: タイトルから正規表現で抽出
- **声優**: 
  1. `dl.informationList dt:contains("声優")` の次の `dd.informationList__txt a` のテキスト（複数の場合は「・」区切り）
  2. フォールバック: タイトルから正規表現で抽出
  3. さらにフォールバック: `.m-productSummary .summary` から「CV」または「声優」を含む行を抽出（":"で分割して取得）
- **メイン画像**: `img[src*="main"]` の `src` 属性、または `meta[property="og:image"]` の `content` 属性
- **トラックリスト**: 現在未実装（d_xxxxxx形式ではトラックリスト取得処理を省略）
- **その他情報**: `#work_outline tr` の th/td ペア

### 3. 設定ファイル管理機能
- **形式**: TOML ファイル (`config/config.toml`)
- **設定項目**:
  - `set_main_image`: メイン画像を MP3 に埋め込むかどうか (bool)
  - `save_parsed_data`: 解析データを JSON ファイルに保存するかどうか (bool)。`true` の場合、`log_dir` 配下に対象ディレクトリごとの JSON (`<dir>.json`) を保存
  - `convert`: 音声ファイルの変換を実行するかどうか (bool)
  - `debug`: デバッグログを出力するかどうか (bool)
  - `exclude_strings`: 除外する文字列のリスト (array)
  - `source_dir`: 変換対象のファイルを配置するディレクトリ (string)
  - `html_dir`: メタデータ取得用の HTML ファイルを配置するディレクトリ (string)
  - `output_dir`: 変換後の MP3 ファイルの出力先 (string)
  - `log_dir`: ログファイルの出力先 (string)
  - `mp3_output_dir_name`: MP3 出力ディレクトリ名 (string)

### 4. 対話型 HTML ファイル生成機能
- **コマンド**: `-create-html` フラグ付きで実行
- **入力項目**:
  - HTML ファイル名 (.html は自動付加)
  - アルバムタイトル
  - サークル名
  - 詳細情報 (ジャンル、作者など、任意のキーバリュー)
  - トラックリスト (タイトルと再生時間)
- **出力**: 指定された HTML ディレクトリに HTML ファイルを生成
- **上書き確認**: 既存ファイルが存在する場合、上書き確認を行う

### 5. メイン画像埋め込み機能
- **条件**: `set_main_image = true` の場合のみ有効
- **ファイル名規則**: `[ディレクトリ名].webp` または `[ディレクトリ名].jpg`
- **検索場所**: `image_dir` 直下
- **優先順位**: webp 形式が優先、存在しない場合 jpg を検索
- **例**: ディレクトリ名が `RJ12345678` の場合
  - 検索場所: `image_dir/`
  - ファイル名: `RJ12345678.webp` または `RJ12345678.jpg`

### 6. デバッグログ機能
- **ログレベル**: DEBUG, INFO, WARN, ERROR, FATAL
- **出力先**: コンソール + ファイル
- **ファイル名形式**: `results_YYYYMMDD_HHMMSS.log`
- **デバッグイベント**: 処理の各段階で詳細なログを出力

### 7. 出力ディレクトリ構成と命名
- **構成**: `output_dir/mp3_output_dir_name/Actor/Brand/【Key】AlbumTitle`
- **Actorディレクトリ**: 声優が複数の場合は先頭2名を「・」区切りで連結し、3名以上は末尾に「他」を付与
- **AlbumTitleの省略**: 20文字を超える場合は20文字で切り取り後に `(…略)` を付与
- **出力先の初期化**: 対象ディレクトリが既に存在する場合は内容をクリーンアップしてから書き込み

## システム要件

### 必須要件
- **OS**: Linux, macOS, Windows (クロスプラットフォーム)
- **Go**: 1.24.0 以上
- **FFmpeg**: 4.2 以上 (PATH に含まれること)
- **依存ライブラリ**:
  - github.com/PuerkitoBio/goquery v1.10.2 (HTML パース用)
  - github.com/spf13/viper v1.19.0 (設定ファイル読み込み用)

### 推奨環境
- **開発環境**: devbox (Go 1.24.0 と FFmpeg を自動セットアップ)
- **ビルドツール**: GNU Make

## 設定ファイル仕様

### ファイル形式
TOML (Tom's Obvious, Minimal Language)

### 必須設定項目
```toml
[setting]
set_main_image = true    # メイン画像を設定するかどうか
save_parsed_data = true  # 解析データをJSONファイルに保存するかどうか
convert = true           # 音声ファイルの変換を実行するかどうか
debug = false           # デバッグログを出力するかどうか
exclude_strings = ["SE無し", "SEなし", "効果音無し", "効果音なし", "_MACOSX"]  # 除外する文字列リスト

[dir_setting]
source_dir = "./data/source/"      # 変換対象のファイルを配置するディレクトリ
html_dir = "./data/html/"          # メタデータ用HTMLファイルを配置するディレクトリ
log_dir = "./data/log/"            # ログファイルの出力先
output_dir = "./data/output/"      # 変換後MP3ファイルの出力先
mp3_output_dir_name = "mp3-output" # MP3出力ディレクトリ名
```

### 設定値の検証
- すべてのディレクトリパスは存在確認を行い、存在しない場合は自動作成
- 相対パスは絶対パスに変換して使用

## 処理フロー

### 1. 初期化フェーズ
1. コマンドライン引数の解析 (`-create-html` フラグ確認)
2. HTML 生成モードの場合: 対話型 HTML 生成を実行して終了
3. 通常モードの場合: 設定ファイル読み込みと検証
4. FFmpeg の依存関係確認
5. ログファイルの初期化

### 2. ディレクトリスキャンフェーズ
1. `source_dir` 内のサブディレクトリを列挙
2. 各ディレクトリに対して以下の処理:
   - 同名の HTML ファイルが存在するか確認
   - HTML ファイルをパースしてメタデータを抽出
   - メイン画像設定が有効な場合、画像ファイルを検索
   - パース結果を JSON として保存 (設定により)

### 3. 変換フェーズ
1. 変換対象ディレクトリをソート
2. 各ディレクトリに対して以下の処理:
   - 出力ディレクトリの準備 (`output_dir/mp3_output_dir_name/Actor/Brand/【Key】AlbumTitle`)
   - 音声ファイルの検索 (優先度: WAV > FLAC > MP3、除外文字列を含むファイルはスキップ)
   - 各音声ファイルに対して:
     - MP3 変換実行 (FFmpeg 使用)
     - ID3 タグ設定
     - メイン画像埋め込み (設定により)

### 4. 終了フェーズ
- 処理結果のログ出力
- 処理対象外ファイルの警告表示

## 内部関数

### splitActorNames 関数
声優名を複数の区切り文字で分割します。

**対応する区切り文字**:
- カンマ: `,` `，`
- 中黒: `・`
- スラッシュ: `/` `／`
- 読点: `、`

**処理**:
1. 上記の区切り文字で文字列を分割
2. 各要素の前後の空白を除去
3. 空文字列を除外
4. 分割結果を文字列スライスで返す

**使用例**:
- 入力: `"声優A・声優B・声優C"`
- 出力: `["声優A", "声優B", "声優C"]`

### truncateAlbumTitle 関数
アルバムタイトルが20文字(ルーン数)を超える場合に省略します。

**処理**:
- 20文字以下の場合: そのまま返す
- 20文字超の場合: 先頭20文字 + "(…略)" を返す

**使用例**:
- 入力: `"これは非常に長いアルバムタイトルのテスト"`
- 出力: `"これは非常に長いアルバムタイトル(…略)"`

## 内部関数

### splitActorNames 関数
声優名を複数の区切り文字で分割します。

**対応する区切り文字**:
- カンマ: `,` `、`
- 中黒: `・`
- スラッシュ: `/` `／`
- 読点: `、`

**処理**:
1. 上記の区切り文字で文字列を分割
2. 各要素の前後の空白を除去
3. 空文字列を除外
4. 分割結果を文字列スライスで返す

**使用例**:
- 入力: `"声優A・声優B・声優C"`
- 出力: `["声優A", "声優B", "声優C"]`

### truncateAlbumTitle 関数
アルバムタイトルが20文字(ルーン数)を超える場合に省略します。

**処理**:
- 20文字以下の場合: そのまま返す
- 20文字超の場合: 先頭20文字 + "(…略)" を返す

**使用例**:
- 入力: `"これは非常に長いアルバムタイトルのテスト"`
- 出力: `"これは非常に長いアルバムタイトル(…略)"`

## データモデル

### IndividualData 構造体
```go
type IndividualData struct {
    AlbumTitle string            `json:"album_title"` // アルバムタイトル
    Actor      string            `json:"actor"`       // 声優名（パース結果のキー「声優」または「actor」から取得）
    Brand      string            `json:"brand"`       // ブランド名
    MainImage  string            `json:"main_image"`  // メイン画像のパス
    TrackList  []Track           `json:"track_list"`  // トラック一覧
    Additional map[string]string `json:"additional"`  // 追加情報
}
```

### Track 構造体
```go
type Track struct {
    TrackTitle    string `json:"track_title"`    // トラックタイトル
    TrackDuration string `json:"track_duration"` // 再生時間 (例: "4分30秒")
}
```

### MP3Metadata 構造体
```go
type MP3Metadata struct {
    Artist      string  // アーティスト名
    AlbumArtist string  // アルバムアーティスト名
    AlbumTitle  string  // アルバムタイトル
    TrackName   string  // トラック名
    CoverImage  *string // 画像ファイルのパス (nil の場合画像なし)
}
```

## HTML パース仕様

### 対象要素
- **アルバムタイトル**: `h1#work_name` のテキスト
- **サークル名**: `span[itemprop='brand'].maker_name a` のテキスト
- **概要情報**: `#work_outline tr` の th/td ペア
- **トラックリスト**: `.work_parts.type_tracklist .work_tracklist_item` のタイトルと時間

### トラック情報抽出
- 正規表現: `([^\s]+) \((\d+):(\d+)\)`
- 形式: `タイトル (分:秒)`
- 変換: 分と秒を time.Duration に変換して "X分Y秒" 形式に

## エラー処理

### 依存関係エラー
- FFmpeg がインストールされていない場合: "ffmpeg がインストールされていない、または PATH に見つかりません"

### 設定ファイルエラー
- ファイルが存在しない場合: "設定ファイルの読み込みエラー"
- TOML パースエラー: "設定のパースエラー"
- 必須項目欠落: 各項目ごとにエラーメッセージ

### HTML 処理エラー
- HTML ファイル不存在: 処理対象外リストに追加
- HTML パースエラー: 処理対象外リストに追加

### 画像処理エラー
- メイン画像不存在: 画像不足リストに追加
- 画像パス変換エラー: 画像不足リストに追加

### 変換エラー
- FFmpeg 実行エラー: 個別ファイルの変換失敗
- 出力ディレクトリ作成エラー: 処理中断

## 制限事項

### パフォーマンス制限
- **並列処理**: 現在は順次処理 (大量ファイルで時間がかかる)
- **メモリ使用量**: FFmpeg プロセスによりメモリ使用量が増加
- **ビットレート**: 320kbps 固定 (高品質設定)

### 機能制限
- **HTML ファイル必須**: 各ディレクトリに対応する HTML ファイルが必要
- **ファイル名制約**: ディレクトリ名と HTML ファイル名が一致している必要
- **画像形式**: webp/jpg のみ対応
- **ID3 バージョン**: v2.3 のみ対応
- **除外ファイル**: 設定ファイルで指定した除外文字列を含むファイルは変換対象外

### システム制限
- **プラットフォーム**: Go がサポートするプラットフォームに依存
- **FFmpeg バージョン**: 4.2 以上が必要
- **ファイルシステム**: UTF-8 エンコーディングを前提

## セキュリティ考慮事項

- **ファイルパス**: ユーザー入力のパスは適切に検証
- **コマンド実行**: FFmpeg コマンドは exec.CommandContext で実行、タイムアウト可能
- **ファイルアクセス**: 設定されたディレクトリ内のみアクセス
- **ログ出力**: デバッグモードでのみ詳細情報出力

## テスト仕様

### ユニットテスト対象
- 設定ファイル読み込み (`config_test.go`)
- HTML パース (`parser_test.go`)
- データモデル (`data_test.go`)
- ログ機能 (`logger_test.go`)
- 音声変換 (`audioconverter_test.go`)
- HTML 生成 (`template_test.go`, `interactive_test.go`)

### テスト実行
```bash
make test      # 通常テスト
make testv     # 詳細出力テスト
make coverage  # カバレッジレポート生成
```

## 開発・運用情報

### ビルドプロセス
1. `make deps`: 依存関係インストール
2. `make build`: バイナリビルド
3. `make clean`: ビルド成果物削除

### ログ仕様
- **ファイル形式**: プレーンテキスト
- **タイムスタンプ**: `20060102_150405` 形式
- **ログレベル**: DEBUG/INFO/WARN/ERROR/FATAL
- **出力内容**: 処理ステップ、ファイルパス、エラー詳細

### バージョン管理
- **Go バージョン**: 1.24.0
- **モジュール管理**: go.mod
- **依存関係**: 固定バージョン使用

---

この仕様書は実装コードに基づいて記述されています。実際の動作はソースコードを参照してください。
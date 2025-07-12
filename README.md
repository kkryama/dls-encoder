# dls-encoder

**DLS(Dynamic Labeling System) Encoder**

dls-encoder(`Dynamic Labeling System-Encoder`)はFFmpegを利用してMP3にエンコードし、ID3タグを自動的に設定するツールです。
音楽ライブラリやデータベースのメタデータ管理を簡素化し、特に個人用の音楽整理に役立ちます。

## 機能

- **音声変換**：WAV、MP3ファイルからMP3への変換
    - 同じディレクトリにWAVとMP3が存在する場合、WAVを優先してエンコード
    - 320kbps、48kHzの高音質設定
- **メタデータ自動設定**：同名のHTMLファイルを参照してID3タグを自動設定
- **設定ファイル管理**：TOMLファイルによる柔軟なディレクトリ管理
- **対話型HTMLファイル生成機能**
    - アルバム情報を対話形式で入力
    - トラックリストの自動生成
    - 既存ファイルの上書き確認
- **画像埋め込み**：メイン画像のMP3への埋め込み
- **デバッグログ**：詳細なログ出力でトラブルシューティングを支援

### メイン画像のファイル名規則

`set_main_image = true` に設定した場合、以下の規則でメイン画像ファイルを自動検索します：

- **ファイル名形式**：`[ディレクトリ名]_img_main.webp` または `[ディレクトリ名]_img_main.jpg`
- **検索場所**：HTMLディレクトリ内の `[ディレクトリ名]_files` フォルダ
- **優先順位**：webp形式が優先され、見つからない場合にjpg形式を検索
- **例**：ディレクトリ名が `RJ12345678` の場合
  - 検索場所：`html_dir/RJ12345678_files/`
  - ファイル名：`RJ12345678_img_main.webp` または `RJ12345678_img_main.jpg`

## 必要条件

- FFmpeg v4.2以上
- Go 1.24.0以上

## インストール

1. リポジトリのクローン:
```bash
git clone https://github.com/kkryama/dls-encoder.git
cd dls-encoder
```

2. 依存関係のインストール:
```bash
make deps
```

3. ビルド:
```bash
make build
```

## 実行方法

### コマンドライン引数

- `-create-html`: HTMLファイルを対話形式で作成します

### エンコード実行

1. 設定ファイル `config/config.toml` の確認、必要に応じて編集
2. 変換したいファイルを含むディレクトリを設定ファイルの `source_dir` に指定した先に配置
3. 元のディレクトリ名と同一のHTMLファイルを `html_dir` に指定した先に配置
4. 以下のコマンドを実行:

```bash
./dls-encoder  # ビルド済みバイナリを使用する場合
# または
go run cmd/main.go  # ソースから直接実行する場合
```

実行するとID3タグを設定しエンコードされたファイルが `output_dir/mp3_output_dir_name/Actor/Brand/AlbumTitle` 以下に配置されます。
Actor, Brand, AlbumTitle はHTMLファイルをパースした結果が利用されます。

### HTMLファイルの生成

対話型のHTMLファイル生成機能を使用して、必要なメタデータを含むHTMLファイルを作成できます：

```bash
./dls-encoder -create-html  # HTMLファイル生成モード
```

以下の情報を対話形式で入力できます：
- HTMLファイル名（.htmlは自動で付加）
- アルバムタイトル
- サークル名
- 詳細情報（ジャンル、作者など）
- トラックリスト（タイトルと再生時間）

生成されたHTMLファイルは `html_dir` で指定されたディレクトリに保存されます。
HTMLファイルを生成してエンコードまで実施する場合、 `set_main_image` の設定は `false` にするか手動で画像の配置をする必要があることに注意してください。

**メイン画像を設定する場合の手順**：
1. HTMLファイル名と同じ名前のディレクトリを作成（例：`test.html` → `test_files/`）
2. そのディレクトリ内に `[HTMLファイル名]_img_main.webp` または `[HTMLファイル名]_img_main.jpg` を配置
3. 例：`test.html` の場合は `test_files/test_img_main.webp` または `test_files/test_img_main.jpg`

### 設定ファイル例

```toml
[setting]
set_main_image = true    # メイン画像を設定するかどうか
save_parsed_data = true  # HTMLをパースしたデータを保存するかどうか
convert = true           # 音声ファイルの変換を実行するかどうか
debug = false           # デバッグログを出力するかどうか

[dir_setting]
source_dir = "./data/source/"      # 変換対象のファイルを配置するディレクトリ
html_dir = "./data/html/"          # メタデータ取得用のHTMLファイルを配置するディレクトリ
log_dir = "./data/log/"            # ログファイルの出力先
output_dir = "./data/output/"      # 変換後のMP3ファイルの出力先
mp3_output_dir_name = "mp3-output" # MP3出力ディレクトリ名
```

### 設定ファイル詳細

`config/config.toml` で以下の設定が可能です：

#### [setting] セクション
- `set_main_image`：メイン画像をMP3に埋め込むかどうか（true/false）
- `save_parsed_data`：HTMLをパースしたデータをJSONファイルとして保存するかどうか（true/false）
- `convert`：音声ファイルの変換を実行するかどうか（true/false）
- `debug`：デバッグログを出力するかどうか（true/false）

#### [dir_setting] セクション
- `source_dir`：変換対象のファイルを配置するディレクトリ
- `html_dir`：メタデータ取得用のHTMLファイルを配置するディレクトリ
- `output_dir`：変換後のMP3ファイルの出力先
- `log_dir`：ログファイルの出力先
- `mp3_output_dir_name`：MP3出力ディレクトリ名

HTMLをパースした結果のみ確認したい場合は `save_parsed_data: true, convert: false` と設定してください。

#### 除外ファイル
以下の文字列を含むファイルは自動的に除外されます：
- `SE無し`、`SEなし`
- `効果音無し`、`効果音なし`
- `__MACOSX`

## パフォーマンスと制限事項

### パフォーマンス
- **並列処理**：現在は順次処理のため、大量のファイルでは時間がかかります
- **メモリ使用量**：FFmpegプロセスによりメモリ使用量が増加することがあります
- **320kbps固定**：高品質設定のため、ファイルサイズが大きくなります

### 制限事項
- **HTMLファイル必須**：各ディレクトリに対応するHTMLファイルが必要
- **ファイル名の制約**：ディレクトリ名とHTMLファイル名が一致している必要があります

## トラブルシューティング

### よくある問題

1. **FFmpegが見つからない場合**:
   ```
   ffmpeg がインストールされていない、または PATH に見つかりません
   ```
   - FFmpegが正しくインストールされているか確認してください
   - PATHが正しく設定されているか確認してください

2. **HTMLパースエラー**:
   ```
   下記のファイルはMP3変換できませんでした。[html_dir]に対応するHTMLファイルが存在するか確認してください
   ```
   - HTMLファイルが正しい形式であることを確認してください
   - ファイル名が対象ディレクトリ名と一致しているか確認してください
   - HTMLファイルに必要な要素が含まれているか確認してください

3. **出力ディレクトリにファイルが生成されない**:
   - 設定ファイルのパスが正しいか確認してください
   - 必要な権限があるか確認してください
   - `convert = true` に設定されているか確認してください

4. **画像が埋め込まれない**:
   ```
   下記のファイルはメイン画像が見つからず、MP3変換処理を実行できませんでした
   ```
   - HTMLディレクトリに画像ファイルが存在するか確認してください
   - 画像ファイル名が正しい形式であるか確認してください：
     - 形式：`[ディレクトリ名]_img_main.webp` または `[ディレクトリ名]_img_main.jpg`
     - 配置場所：`html_dir/[ディレクトリ名]_files/`
   - `set_main_image = false` に設定するか、手動で画像を配置してください

5. **設定ファイルエラー**:
   ```
   設定ファイルの読み込みに失敗
   ```
   - `config/config.toml` ファイルが存在するか確認してください
   - TOML形式が正しいか確認してください
   - 必要なディレクトリが存在するか確認してください

### デバッグ方法

1. **デバッグモードの有効化**:
   ```toml
   [setting]
   debug = true
   ```

2. **ログファイルの確認**:
   - ログファイルは `log_dir` で指定したディレクトリに生成されます
   - ファイル名形式：`results_YYYYMMDD_HHMMSS.log`

3. **パース結果の確認**:
   ```toml
   [setting]
   save_parsed_data = true
   convert = false
   ```
   パースしたデータがJSONファイルとして保存され、変換処理は実行されません

## 開発者向け情報

### テスト実行

```bash
make test      # 通常のテスト実行
make testv     # 詳細な出力でテスト実行
make coverage  # カバレッジレポート付きでテスト実行
```

### コード品質

```bash
make fmt   # コードフォーマット
make lint  # リンター実行
```

### クリーンアップ

```bash
make clean # ビルド成果物、ログ、一時ファイルの削除
```

`make clean` コマンドで以下のファイルが削除されます：
- ビルドされたバイナリ（`dls-encoder`）
- カバレッジレポート（`coverage.out`, `coverage.html`）
- ログファイル（`data/log/*.log`）
- パース結果のJSONファイル（`data/log/*.json`）
- テスト用一時ファイル（`*.test`）
- その他の一時ファイル（`*.tmp`）

## 開発

### 開発環境のセットアップ

#### 方法1: devbox（推奨）

1. **前提条件**
   - [devbox](https://www.jetify.com/devbox)のインストール
   - Git

2. **devboxによる環境構築**
   ```bash
   git clone https://github.com/kkryama/dls-encoder.git
   cd dls-encoder
   devbox shell  # Go 1.24.0とFFmpegが自動でセットアップされます
   ```

3. **依存関係のインストール**
   ```bash
   make deps
   ```

#### 方法2: 手動セットアップ

1. **前提条件**
   - Go 1.24.0以上
   - FFmpeg v4.2以上
   - Git

2. **リポジトリのクローン**
   ```bash
   git clone https://github.com/kkryama/dls-encoder.git
   cd dls-encoder
   ```

3. **依存関係のインストール**
   ```bash
   make deps
   ```

## アーキテクチャ

### プロジェクト構成

```
dls-encoder/
├── cmd/
│   └── main.go                    # エントリーポイント
├── internal/
│   ├── audioconverter/            # 音声変換機能
│   │   ├── create.go              # MP3変換とメタデータ設定
│   │   └── find.go                # 音声ファイル検索
│   ├── config/                    # 設定管理
│   │   ├── config.go              # 設定構造体定義
│   │   └── load_config.go         # 設定ファイル読み込み
│   ├── generator/                 # HTML生成機能
│   │   ├── interactive.go         # 対話型HTMLファイル生成
│   │   └── template.go            # HTMLテンプレート
│   ├── logger/                    # ログ出力機能
│   │   ├── logger.go              # ログ出力の実装
│   │   └── debug.go               # デバッグ機能
│   ├── model/                     # データモデル
│   │   └── data.go                # データ構造体定義
│   ├── parser/                    # HTML解析機能
│   │   ├── parse.go               # HTMLファイル解析
│   │   └── html_extractor.go      # HTML要素抽出
│   └── storage/                   # ファイル管理機能
│       ├── load_target.go         # 対象ディレクトリ読み込み
│       ├── save_json.go           # JSON保存
│       └── find_main_image.go     # メイン画像検索
├── config/
│   └── config.toml                # 設定ファイル
└── data/                          # 処理対象データ
    ├── source/                    # 変換元音声ファイル
    ├── html/                      # メタデータ用HTMLファイル
    ├── output/                    # 変換後MP3ファイル
    └── log/                       # ログファイル
```

### 処理フロー

1. **初期化**：設定読み込み、依存関係確認、ログ初期化
2. **ディレクトリスキャン**：source_dir内の対象ディレクトリを検索
3. **HTML解析**：各ディレクトリに対応するHTMLファイルをパース
4. **画像検索**：メイン画像ファイルを検索（設定により）
5. **音声ファイル検索**：WAV/MP3ファイルを検索、WAV優先
6. **MP3変換**：FFmpegによる変換とメタデータ設定
7. **結果出力**：処理結果とエラー情報をログ出力

---

**注意**: このツールは個人使用向けに設計されており、動作保証やサポートは行っていません。

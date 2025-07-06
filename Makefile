# Makefile for dls-encoder

# Go関連の変数
GO=go
GOFLAGS=-v
GOFMT=gofmt
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOTEST=$(GO) test
GOGET=$(GO) get
GOMOD=$(GO) mod
BINARY_NAME=dls-encoder

# プロジェクトのディレクトリ
CMD_DIR=./cmd
INTERNAL_DIR=./internal

# 全てのパッケージ
ALL_PACKAGES=./...

.PHONY: all build test clean fmt lint deps help coverage devbox-init devbox-shell

# デフォルトターゲット
all: deps test build

# ヘルプを表示
help:
	@echo "利用可能なコマンド:"
	@echo "  make          - 依存関係のインストール、テスト実行、ビルドを行います"
	@echo "  make build    - プログラムをビルドします"
	@echo "  make test     - テストを実行します"
	@echo "  make testv    - 詳細な出力でテストを実行します"
	@echo "  make coverage - カバレッジレポート付きでテストを実行します"
	@echo "  make fmt      - コードをフォーマットします"
	@echo "  make lint     - リンターを実行します"
	@echo "  make deps     - 依存関係をインストールします"
	@echo "  make clean    - ビルド成果物を削除します"
	@echo ""
	@echo "devbox関連:"
	@echo "  make devbox-init  - devbox環境を初期化します"
	@echo "  make devbox-shell - devboxシェルを起動します"

# ビルド
build:
	$(GOBUILD) -o $(BINARY_NAME) $(CMD_DIR)

# テスト実行（標準出力）
test:
	$(GOTEST) $(ALL_PACKAGES)

# テスト実行（詳細出力）
testv:
	$(GOTEST) -v $(ALL_PACKAGES)

# カバレッジ付きテスト
coverage:
	$(GOTEST) -coverprofile=coverage.out $(ALL_PACKAGES)
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "カバレッジレポートが coverage.html に生成されました"

# コードフォーマット
fmt:
	$(GOFMT) -w $(CMD_DIR) $(INTERNAL_DIR)

# リンター実行
lint:
	$(HOME)/go/bin/golangci-lint run

# 依存関係のインストール
deps:
	$(GOMOD) download
	$(GOMOD) tidy
	@if ! command -v $(HOME)/go/bin/golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lintをインストールしています..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi

# ビルド成果物、ログ、一時ファイルのクリーンアップ
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -f data/log/*.log
	rm -f data/log/*.json
	find . -type f -name "*.test" -delete
	find . -type f -name "*.tmp" -delete

# devbox環境の初期化
devbox-init:
	@if ! command -v devbox >/dev/null 2>&1; then \
		echo "devboxがインストールされていません。"; \
		echo "インストール方法: https://www.jetify.com/devbox/docs/installing_devbox/"; \
		exit 1; \
	fi
	devbox install
	@echo "devbox環境が初期化されました。'make devbox-shell' または 'devbox shell' で環境に入ってください。"

# devboxシェルの起動
devbox-shell:
	@if ! command -v devbox >/dev/null 2>&1; then \
		echo "devboxがインストールされていません。'make devbox-init' を先に実行してください。"; \
		exit 1; \
	fi
	devbox shell

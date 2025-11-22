#!/bin/bash

# output_dir をクリーンアップするスクリプト
# config.toml から output_dir を読み取ってクリーンアップします

# config.toml のパス
CONFIG_FILE="./config/config.toml"

# output_dir を抽出 (シンプルなパース)
OUTPUT_DIR=$(grep '^output_dir' "$CONFIG_FILE" | cut -d'"' -f2)

if [ -z "$OUTPUT_DIR" ]; then
    echo "output_dir が config.toml に見つかりません"
    exit 1
fi

# 絶対パスに変換
OUTPUT_DIR=$(realpath "$OUTPUT_DIR")

echo "クリーンアップ対象ディレクトリ: $OUTPUT_DIR"

if [ ! -d "$OUTPUT_DIR" ]; then
    echo "ディレクトリが存在しません: $OUTPUT_DIR"
    exit 1
fi

echo "このディレクトリ内のファイルをすべて削除します。よろしいですか？ (y/n)"
read -p "" confirm

if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
    rm -rf "$OUTPUT_DIR"/*
    echo "クリーンアップ完了"
else
    echo "キャンセルされました"
fi
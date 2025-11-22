#!/usr/bin/env bash
set -euo pipefail

# config.toml から source_dir を取得
CONFIG_FILE="./config/config.toml"
SOURCE_DIR=$(grep '^source_dir' "$CONFIG_FILE" | cut -d'"' -f2)

if [ -z "$SOURCE_DIR" ]; then
    echo "source_dir が config.toml に見つかりません"
    exit 1
fi

# 絶対パスに変換
SOURCE_DIR=$(realpath "$SOURCE_DIR")

echo "展開対象ディレクトリ: $SOURCE_DIR"

if [ ! -d "$SOURCE_DIR" ]; then
    echo "ディレクトリが存在しません: $SOURCE_DIR"
    exit 1
fi

# source_dir に移動
cd "$SOURCE_DIR"

shopt -s nullglob
zipfiles=(*.zip)

for zipfile in "${zipfiles[@]}"; do
  dirname="$(basename "$zipfile" .zip)"
  [ -d "$dirname" ] && rm -rf "$dirname"
  mkdir -p "$dirname"
  echo "Extracting $zipfile into $dirname..."
  unzip -O CP932 "$zipfile" -d "$dirname"
done

echo "All zip files processed!"


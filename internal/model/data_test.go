package model

import (
	"reflect"
	"testing"
)

func TestTrackStruct(t *testing.T) {
	track := Track{
		TrackTitle:    "テストトラック",
		TrackDuration: "5:30",
	}

	// フィールドの値を確認
	if track.TrackTitle != "テストトラック" {
		t.Errorf("TrackTitle: got %v, want %v", track.TrackTitle, "テストトラック")
	}
	if track.TrackDuration != "5:30" {
		t.Errorf("TrackDuration: got %v, want %v", track.TrackDuration, "5:30")
	}
}

func TestIndividualDataStruct(t *testing.T) {
	tracks := []Track{
		{TrackTitle: "トラック1", TrackDuration: "3:45"},
		{TrackTitle: "トラック2", TrackDuration: "4:20"},
	}

	additional := map[string]string{
		"ジャンル": "ドラマ",
		"発売日":  "2025-07-12",
	}

	data := IndividualData{
		AlbumTitle: "テストアルバム",
		Actor:      "テスト声優",
		Brand:      "テストブランド",
		MainImage:  "test.jpg",
		TrackList:  tracks,
		Additional: additional,
	}

	// 基本フィールドの検証
	if data.AlbumTitle != "テストアルバム" {
		t.Errorf("AlbumTitle: got %v, want %v", data.AlbumTitle, "テストアルバム")
	}
	if data.Actor != "テスト声優" {
		t.Errorf("Actor: got %v, want %v", data.Actor, "テスト声優")
	}
	if data.Brand != "テストブランド" {
		t.Errorf("Brand: got %v, want %v", data.Brand, "テストブランド")
	}
	if data.MainImage != "test.jpg" {
		t.Errorf("MainImage: got %v, want %v", data.MainImage, "test.jpg")
	}

	// TrackListの検証
	if !reflect.DeepEqual(data.TrackList, tracks) {
		t.Errorf("TrackList: got %v, want %v", data.TrackList, tracks)
	}

	// Additionalマップの検証
	if !reflect.DeepEqual(data.Additional, additional) {
		t.Errorf("Additional: got %v, want %v", data.Additional, additional)
	}
}

func TestIndividualDataStructEmpty(t *testing.T) {
	// 空のデータ構造のテスト
	data := IndividualData{}

	if data.AlbumTitle != "" {
		t.Error("空のIndividualDataのAlbumTitleは空文字列であるべき")
	}
	if data.Actor != "" {
		t.Error("空のIndividualDataのActorは空文字列であるべき")
	}
	if data.Brand != "" {
		t.Error("空のIndividualDataのBrandは空文字列であるべき")
	}
	if data.MainImage != "" {
		t.Error("空のIndividualDataのMainImageは空文字列であるべき")
	}
	if data.TrackList != nil {
		t.Error("空のIndividualDataのTrackListはnilであるべき")
	}
	if data.Additional != nil {
		t.Error("空のIndividualDataのAdditionalはnilであるべき")
	}
}

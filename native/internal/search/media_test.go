package search

import (
	"strings"
	"testing"
)

func TestNormalizeImageMediaDefaultsKindAndURLs(t *testing.T) {
	media := normalizeMedia(&Media{
		ThumbnailURL: " https://images.example/thumb.jpg#fragment ",
		MIMEType:     " IMAGE/JPEG ",
		Width:        1200,
		Height:       800,
		Alt:          " Golden Gate Bridge ",
	}, CategoryImages)
	if media == nil {
		t.Fatal("image media was rejected")
	}
	if media.Kind != MediaKindImage {
		t.Fatalf("kind = %q", media.Kind)
	}
	if media.ThumbnailURL != "https://images.example/thumb.jpg" || media.ContentURL != media.ThumbnailURL {
		t.Fatalf("normalized URLs = %q / %q", media.ThumbnailURL, media.ContentURL)
	}
	if media.MIMEType != "image/jpeg" || media.Width != 1200 || media.Height != 800 || media.Alt != "Golden Gate Bridge" {
		t.Fatalf("unexpected normalized media: %+v", media)
	}
}

func TestNormalizeMediaRejectsUnsafeOrUnsupportedPayloads(t *testing.T) {
	cases := []*Media{
		{Kind: "video", ContentURL: "https://media.example/video.mp4"},
		{Kind: MediaKindImage, ContentURL: "file:///etc/passwd"},
		{Kind: MediaKindImage, ContentURL: "https://user:pass@images.example/private.jpg"},
	}
	for _, candidate := range cases {
		if got := normalizeMedia(candidate, CategoryImages); got != nil {
			t.Fatalf("unsafe media accepted: %+v", got)
		}
	}
}

func TestNormalizeMediaBoundsProviderMetadata(t *testing.T) {
	media := normalizeMedia(&Media{
		Kind:       MediaKindImage,
		ContentURL: "https://images.example/full.png",
		MIMEType:   "text/html",
		Width:      maxMediaDimension + 1,
		Height:     -1,
		Alt:        strings.Repeat("x", maxMediaAltRunes+20),
	}, CategoryImages)
	if media == nil {
		t.Fatal("valid image URL was rejected")
	}
	if media.MIMEType != "" || media.Width != 0 || media.Height != 0 {
		t.Fatalf("untrusted metadata was not bounded: %+v", media)
	}
	if len([]rune(media.Alt)) != maxMediaAltRunes {
		t.Fatalf("alt length = %d, want %d", len([]rune(media.Alt)), maxMediaAltRunes)
	}
}

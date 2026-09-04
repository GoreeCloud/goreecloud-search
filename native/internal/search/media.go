package search

import (
	"strings"
)

const (
	MediaKindImage  = "image"
	maxMediaAltRunes = 512
	maxMediaDimension = 100000
)

// Media describes provider-supplied media associated with a normalized Search
// result. External media URLs are data only: the browser-facing web layer must
// route them through the GoreeCloud media boundary rather than embedding them
// directly in rendered pages.
type Media struct {
	Kind         string `json:"kind"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	ContentURL   string `json:"content_url,omitempty"`
	MIMEType     string `json:"mime_type,omitempty"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	Alt          string `json:"alt,omitempty"`
}

func normalizeMedia(media *Media, category string) *Media {
	if media == nil {
		return nil
	}

	kind := strings.ToLower(strings.TrimSpace(media.Kind))
	if kind == "" && category == CategoryImages {
		kind = MediaKindImage
	}
	if kind != MediaKindImage {
		return nil
	}

	thumbnailURL, thumbnailOK := normalizeResultURL(media.ThumbnailURL)
	contentURL, contentOK := normalizeResultURL(media.ContentURL)
	if !thumbnailOK && !contentOK {
		return nil
	}
	if !thumbnailOK {
		thumbnailURL = contentURL
	}
	if !contentOK {
		contentURL = thumbnailURL
	}

	mimeType := strings.ToLower(strings.TrimSpace(media.MIMEType))
	if mimeType != "" && !strings.HasPrefix(mimeType, "image/") {
		mimeType = ""
	}

	width := normalizeMediaDimension(media.Width)
	height := normalizeMediaDimension(media.Height)
	alt := strings.TrimSpace(media.Alt)
	if runes := []rune(alt); len(runes) > maxMediaAltRunes {
		alt = string(runes[:maxMediaAltRunes])
	}

	return &Media{
		Kind:         MediaKindImage,
		ThumbnailURL: thumbnailURL,
		ContentURL:   contentURL,
		MIMEType:     mimeType,
		Width:        width,
		Height:       height,
		Alt:          alt,
	}
}

func normalizeMediaDimension(value int) int {
	if value <= 0 || value > maxMediaDimension {
		return 0
	}
	return value
}

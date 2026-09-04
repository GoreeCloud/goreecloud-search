package mediaproxy

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"testing"
)

func TestURLBuildsSignedSameOriginRoute(t *testing.T) {
	proxy, err := New()
	if err != nil {
		t.Fatal(err)
	}
	got := proxy.URL("https://images.example/photo.jpg#ignored")
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Path != "/media/image" {
		t.Fatalf("path = %q", parsed.Path)
	}
	if parsed.Query().Get("u") != "https://images.example/photo.jpg" || parsed.Query().Get("sig") == "" {
		t.Fatalf("unexpected signed URL: %q", got)
	}
	if strings.Contains(got, "#ignored") {
		t.Fatalf("fragment leaked into proxy URL: %q", got)
	}
}

func TestURLRejectsUnsafeTargetSyntax(t *testing.T) {
	proxy, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{
		"file:///etc/passwd",
		"https://user:pass@example.com/image.jpg",
		"https://example.com:8443/image.jpg",
		"http://localhost/image.jpg",
		"http://internal/image.jpg",
	} {
		if got := proxy.URL(target); got != "" {
			t.Fatalf("target %q produced proxy URL %q", target, got)
		}
	}
}

func TestHandlerRejectsTamperedSignatureBeforeFetch(t *testing.T) {
	proxy, err := New()
	if err != nil {
		t.Fatal(err)
	}
	signed, err := url.Parse(proxy.URL("https://images.example/photo.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	query := signed.Query()
	query.Set("u", "http://127.0.0.1/private.jpg")
	signed.RawQuery = query.Encode()

	response := httptest.NewRecorder()
	proxy.ServeHTTP(response, httptest.NewRequest(http.MethodGet, signed.String(), nil))
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func TestHandlerRejectsSignedPrivateAddress(t *testing.T) {
	proxy, err := New()
	if err != nil {
		t.Fatal(err)
	}
	signed := proxy.URL("http://127.0.0.1/image.jpg")
	if signed == "" {
		t.Fatal("numeric target should be signed so request-time address policy is exercised")
	}
	response := httptest.NewRecorder()
	proxy.ServeHTTP(response, httptest.NewRequest(http.MethodGet, signed, nil))
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusForbidden, response.Body.String())
	}
}

func TestPublicAddressPolicyRejectsReservedRanges(t *testing.T) {
	for _, raw := range []string{"127.0.0.1", "10.0.0.1", "100.64.0.1", "192.0.2.10", "198.51.100.20", "203.0.113.30", "::1", "2001:db8::1"} {
		if isPublicAddress(netip.MustParseAddr(raw)) {
			t.Fatalf("reserved address accepted: %s", raw)
		}
	}
	for _, raw := range []string{"1.1.1.1", "8.8.8.8", "2606:4700:4700::1111"} {
		if !isPublicAddress(netip.MustParseAddr(raw)) {
			t.Fatalf("public address rejected: %s", raw)
		}
	}
}

func TestAllowedImageTypesExcludeActiveDocumentFormats(t *testing.T) {
	for _, contentType := range []string{"image/jpeg", "image/png", "image/gif", "image/webp"} {
		if !allowedImageType(contentType) {
			t.Fatalf("expected allowed image type %q", contentType)
		}
	}
	for _, contentType := range []string{"image/svg+xml", "text/html", "application/octet-stream"} {
		if allowedImageType(contentType) {
			t.Fatalf("unexpected allowed media type %q", contentType)
		}
	}
}

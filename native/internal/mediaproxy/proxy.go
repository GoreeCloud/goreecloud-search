package mediaproxy

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	maxImageBytes = 12 << 20
	maxRedirects  = 3
)

var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("2001::/32"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
}

var errTargetNotAllowed = errors.New("media target is not allowed")

// Proxy is a request-local privacy and SSRF boundary for provider-supplied
// image URLs. Rendered Search pages receive only signed same-origin URLs; the
// browser never receives an external image URL as an image source.
type Proxy struct {
	key    []byte
	client *http.Client
}

func New() (*Proxy, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("create media proxy signing key: %w", err)
	}
	proxy := &Proxy{key: key}
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           proxy.dialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          32,
		MaxIdleConnsPerHost:   4,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 8 * time.Second,
	}
	proxy.client = &http.Client{
		Transport: transport,
		Timeout:   12 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return errors.New("media redirect limit exceeded")
			}
			if err := validateTargetURL(req.URL); err != nil {
				return err
			}
			_, err := publicAddresses(req.Context(), req.URL.Hostname())
			return err
		},
	}
	return proxy, nil
}

// URL returns a signed same-origin media endpoint for a syntactically allowed
// external image URL. Network address validation is deliberately repeated at
// request time so DNS changes cannot be trusted from render time.
func (p *Proxy) URL(raw string) string {
	if p == nil || len(p.key) == 0 {
		return ""
	}
	target, err := canonicalTarget(raw)
	if err != nil {
		return ""
	}
	values := url.Values{}
	values.Set("u", target)
	values.Set("sig", p.signature(target))
	return "/media/image?" + values.Encode()
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p == nil || len(p.key) == 0 || p.client == nil {
		http.Error(w, "Media is unavailable", http.StatusServiceUnavailable)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	target, err := canonicalTarget(r.URL.Query().Get("u"))
	if err != nil {
		http.Error(w, "Invalid media request", http.StatusBadRequest)
		return
	}
	if !hmac.Equal([]byte(p.signature(target)), []byte(r.URL.Query().Get("sig"))) {
		http.Error(w, "Invalid media request", http.StatusForbidden)
		return
	}
	parsed, _ := url.Parse(target)
	if _, err := publicAddresses(r.Context(), parsed.Hostname()); err != nil {
		http.Error(w, "Media target is not allowed", http.StatusForbidden)
		return
	}

	request, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
	if err != nil {
		http.Error(w, "Invalid media request", http.StatusBadRequest)
		return
	}
	request.Header.Set("Accept", "image/avif,image/webp,image/png,image/jpeg,image/gif;q=0.9,*/*;q=0.1")
	request.Header.Set("User-Agent", "GoreeCloud-Search-Media/1")

	response, err := p.client.Do(request)
	if err != nil {
		http.Error(w, "Unable to load media", http.StatusBadGateway)
		return
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		http.Error(w, "Unable to load media", http.StatusBadGateway)
		return
	}
	if response.ContentLength > maxImageBytes {
		http.Error(w, "Media is too large", http.StatusRequestEntityTooLarge)
		return
	}

	payload, err := io.ReadAll(io.LimitReader(response.Body, maxImageBytes+1))
	if err != nil {
		http.Error(w, "Unable to load media", http.StatusBadGateway)
		return
	}
	if len(payload) > maxImageBytes {
		http.Error(w, "Media is too large", http.StatusRequestEntityTooLarge)
		return
	}
	contentType := http.DetectContentType(payload)
	if !allowedImageType(contentType) {
		http.Error(w, "Unsupported media type", http.StatusUnsupportedMediaType)
		return
	}

	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
	w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func (p *Proxy) signature(target string) string {
	mac := hmac.New(sha256.New, p.key)
	_, _ = mac.Write([]byte(target))
	return hex.EncodeToString(mac.Sum(nil))
}

func canonicalTarget(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || validateTargetURL(parsed) != nil {
		return "", errTargetNotAllowed
	}
	parsed.Fragment = ""
	return parsed.String(), nil
}

func validateTargetURL(target *url.URL) error {
	if target == nil || target.User != nil || target.Hostname() == "" {
		return errTargetNotAllowed
	}
	if target.Scheme != "http" && target.Scheme != "https" {
		return errTargetNotAllowed
	}
	port := target.Port()
	if port != "" && port != "80" && port != "443" {
		return errTargetNotAllowed
	}
	host := strings.TrimSuffix(strings.ToLower(target.Hostname()), ".")
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || (!strings.Contains(host, ".") && net.ParseIP(host) == nil) {
		return errTargetNotAllowed
	}
	return nil
}

func (p *Proxy) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, errTargetNotAllowed
	}
	addresses, err := publicAddresses(ctx, host)
	if err != nil {
		return nil, err
	}
	dialer := net.Dialer{Timeout: 6 * time.Second, KeepAlive: 20 * time.Second}
	var lastErr error
	for _, address := range addresses {
		connection, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(address.String(), port))
		if dialErr == nil {
			return connection, nil
		}
		lastErr = dialErr
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errTargetNotAllowed
}

func publicAddresses(ctx context.Context, host string) ([]netip.Addr, error) {
	host = strings.TrimSpace(strings.TrimSuffix(host, "."))
	if host == "" {
		return nil, errTargetNotAllowed
	}
	if parsed, err := netip.ParseAddr(host); err == nil {
		parsed = parsed.Unmap()
		if !isPublicAddress(parsed) {
			return nil, errTargetNotAllowed
		}
		return []netip.Addr{parsed}, nil
	}
	addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil || len(addresses) == 0 {
		return nil, errTargetNotAllowed
	}
	public := make([]netip.Addr, 0, len(addresses))
	for _, address := range addresses {
		address = address.Unmap()
		if !isPublicAddress(address) {
			return nil, errTargetNotAllowed
		}
		public = append(public, address)
	}
	return public, nil
}

func isPublicAddress(address netip.Addr) bool {
	if !address.IsValid() || !address.IsGlobalUnicast() || address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() || address.IsLinkLocalMulticast() || address.IsMulticast() || address.IsUnspecified() {
		return false
	}
	for _, prefix := range blockedPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}

func allowedImageType(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	switch contentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "image/avif":
		return true
	default:
		return false
	}
}

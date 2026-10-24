// Package icons proxies vendor architecture icons from icon.icepanel.io
// through our own origin so client-side canvas export (PNG / SVG / PDF)
// can serialize them.
//
// The CDN itself doesn't send CORS headers. Without this proxy, DOM-to-
// image libraries (html-to-image, dom-to-image) trip the browser's
// tainted-canvas guard when they try to inline our AWS/GCP node icons,
// and PNG/SVG export fails with SecurityError.
//
// The proxy also gives us a single place to swap the underlying source
// later (e.g. move to a self-hosted CDN, or serve pre-vendored SVGs
// from disk once scripts/vendor-icons.mjs has been run).
package icons

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/oche/diagram/pkg/response"
)

const (
	upstreamHost = "icon.icepanel.io"
	upstream     = "https://" + upstreamHost
	maxBytes     = 1 << 20 // 1MB — real vendor SVGs are 5-30KB
	fetchTimeout = 8 * time.Second
	// clientCacheSeconds — how long browsers cache proxied icons.
	// Icons never change, only get added, so aggressive caching is safe.
	// One week matches typical CDN defaults.
	clientCacheSeconds = 60 * 60 * 24 * 7
)

// Handler proxies /api/v1/icons/vendor/{...path} to icon.icepanel.io/{...path}.
// The vendor/ prefix is fixed so we can later add /api/v1/icons/mirror/...
// for the self-hosted variant without breaking clients.
type Handler struct {
	client *http.Client
}

func NewHandler() *Handler {
	return &Handler{
		client: &http.Client{Timeout: fetchTimeout},
	}
}

// Proxy handles GET /api/v1/icons/vendor/{path...}.
// Cross-origin fetches are stripped, only GET is allowed, and paths must
// stay under the vendor directory — no scheme, no host, no ../ traversal.
func (h *Handler) Proxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	const routePrefix = "/api/v1/icons/vendor/"
	if !strings.HasPrefix(r.URL.Path, routePrefix) {
		response.Error(w, http.StatusBadRequest, "unexpected route")
		return
	}
	path := r.URL.Path[len(routePrefix):]
	if !safeUpstreamPath(path) {
		response.Error(w, http.StatusBadRequest, "invalid icon path")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstream+"/"+path, nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "proxy build request")
		return
	}
	req.Header.Set("Accept", "image/svg+xml,image/png,image/*;q=0.5")

	resp, err := h.client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			response.Error(w, http.StatusGatewayTimeout, "upstream timeout")
			return
		}
		response.Error(w, http.StatusBadGateway, "upstream error")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		response.Error(w, http.StatusNotFound, "icon not found")
		return
	}
	if resp.StatusCode >= 400 {
		response.Error(w, http.StatusBadGateway, "upstream error")
		return
	}

	// Copy the content-type; default to svg since that's what we serve
	// 99% of the time. Length limit protects the client from a rogue
	// upstream returning something huge.
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/svg+xml"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", cacheControl(clientCacheSeconds))
	// Explicit CORS so html-to-image's `useCORS` / anonymous crossOrigin
	// paths behave — the whole point of this proxy is to make the
	// canvas API happy. Since we're on our own origin, this is really
	// belt-and-braces for any future embed / iframe use.
	w.Header().Set("Access-Control-Allow-Origin", "*")

	_, _ = io.Copy(w, io.LimitReader(resp.Body, maxBytes))
}

// safeUpstreamPath rejects anything that could be used to reach out of
// the vendor tree: absolute schemes, host authority, dotdot segments,
// or paths starting with '/'. Icepanel paths look like
//   AWS/svg/Compute/Lambda.svg
//   GCP/svg/Cloud-Run.svg
// Letters, digits, /, -, ., _ are all we need.
func safeUpstreamPath(p string) bool {
	if p == "" || len(p) > 200 {
		return false
	}
	if strings.Contains(p, "://") || strings.HasPrefix(p, "/") {
		return false
	}
	if strings.Contains(p, "..") {
		return false
	}
	for _, r := range p {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '/' || r == '-' || r == '.' || r == '_':
		default:
			return false
		}
	}
	return true
}

func cacheControl(seconds int) string {
	if seconds <= 0 {
		return "no-cache"
	}
	return "public, max-age=" + itoa(seconds) + ", immutable"
}

// itoa avoids a strconv import for this one call.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

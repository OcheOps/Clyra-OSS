package pricing

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Detector resolves a region code (ISO-3166-1 alpha-2) for an inbound HTTP request.
//
// Resolution order:
//  1. Explicit `?region=XX` query param (for QA / override)
//  2. Cloudflare's CF-IPCountry header (production behind CF)
//  3. Accept-Language region tag (e.g. "en-NG")
//  4. Upstream ip-api.com lookup on the client IP (cached per-IP)
//  5. Configured default (for localhost dev); empty means global/USD
type Detector struct {
	client        *http.Client
	cache         sync.Map // map[string]cacheEntry keyed by IP
	ttl           time.Duration
	lookup        func(ctx context.Context, ip string) string // overridable in tests
	disabled      bool
	defaultRegion string
}

type cacheEntry struct {
	country string
	exp     time.Time
}

// NewDetector builds a default detector with a 2s lookup timeout and 1h cache TTL.
// defaultRegion is returned when every other signal is inconclusive — useful
// for local development (e.g. "NG" when running on localhost).
func NewDetector(defaultRegion string) *Detector {
	d := &Detector{
		client:        &http.Client{Timeout: 2 * time.Second},
		ttl:           1 * time.Hour,
		defaultRegion: strings.ToUpper(strings.TrimSpace(defaultRegion)),
	}
	d.lookup = d.lookupUpstream
	return d
}

// WithLookup lets tests inject a deterministic IP→country function.
func (d *Detector) WithLookup(fn func(ctx context.Context, ip string) string) *Detector {
	d.lookup = fn
	return d
}

// Detect inspects the request and returns a region code.
//
// Precedence (strongest → weakest):
//  1. ?region=XX           explicit override (QA)
//  2. CF-IPCountry         Cloudflare, authoritative
//  3. IP geo lookup        upstream, authoritative (skipped for private IPs)
//  4. PRICING_DEFAULT_REGION  admin-configured intent (wins on localhost)
//  5. Accept-Language      last-ditch hint (only when nothing stronger exists)
func (d *Detector) Detect(r *http.Request) string {
	if override := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("region"))); override != "" {
		return override
	}

	if cf := strings.ToUpper(strings.TrimSpace(r.Header.Get("CF-IPCountry"))); cf != "" && cf != "XX" {
		return cf
	}

	ip := clientIP(r)
	if ip != "" && !d.disabled {
		parsed := net.ParseIP(ip)
		// Only hit upstream for public IPs.
		if parsed != nil && !parsed.IsLoopback() && !parsed.IsPrivate() && !parsed.IsLinkLocalUnicast() {
			if v, ok := d.cache.Load(ip); ok {
				if e := v.(cacheEntry); time.Now().Before(e.exp) {
					if e.country != "" {
						return e.country
					}
				}
			}
			country := d.lookup(r.Context(), ip)
			d.cache.Store(ip, cacheEntry{country: country, exp: time.Now().Add(d.ttl)})
			if country != "" {
				return country
			}
		}
	}

	// Admin-configured default beats a soft Accept-Language hint so a dev in
	// Nigeria can force NG locally even when their browser says en-US.
	if d.defaultRegion != "" {
		return d.defaultRegion
	}

	return regionFromAcceptLanguage(r.Header.Get("Accept-Language"))
}

// regionFromAcceptLanguage pulls a region tag out of the first language with one.
// "en-NG,en;q=0.8" → "NG".
func regionFromAcceptLanguage(h string) string {
	if h == "" {
		return ""
	}
	for _, raw := range strings.Split(h, ",") {
		tag := strings.TrimSpace(raw)
		if semi := strings.Index(tag, ";"); semi >= 0 {
			tag = tag[:semi]
		}
		if dash := strings.Index(tag, "-"); dash >= 0 && dash+1 < len(tag) {
			region := strings.ToUpper(strings.TrimSpace(tag[dash+1:]))
			// Region tags are 2 letters (or a 3-digit M49 code); we only accept the former.
			if len(region) == 2 {
				return region
			}
		}
	}
	return ""
}

// clientIP extracts the best-guess caller IP, preferring X-Forwarded-For.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// First entry in the comma-separated list is the original client.
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return strings.TrimSpace(xr)
	}
	// RemoteAddr is "host:port" — strip the port.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// lookupUpstream queries ip-api.com. Free tier, no auth, rate-limited (~45/min/IP).
// Returns "" on failure — the cache still records this so we don't hammer the API.
func (d *Detector) lookupUpstream(ctx context.Context, ip string) string {
	url := "http://ip-api.com/json/" + ip + "?fields=status,countryCode"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ""
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	var payload struct {
		Status      string `json:"status"`
		CountryCode string `json:"countryCode"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return ""
	}
	if payload.Status != "success" {
		return ""
	}
	return strings.ToUpper(payload.CountryCode)
}

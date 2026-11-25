// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-10-26, v1.7.1 — caldera-keys-service#2 G16): pins the
// redirect rule. Go's default redirect policy strips only Authorization and
// Cookie on a cross-host hop, so the x-api-key header used to be forwarded to
// whatever host a redirect named. These tests prove a cross-origin redirect
// is not followed (the caller gets the 3xx as an *APIError) and the second
// server never sees the key; a same-origin redirect (same scheme, host and
// port) is followed; and a same-origin chain stops after three hops.
//
// CHANGELOG (2026-10-26, review of !243): the origin is now compared as
// scheme + hostname + effective port, not the raw Host string. New tests:
// an explicit default port (https://h vs https://h:443) is the same origin,
// another port is not, two spellings of one IPv6 literal are the same
// origin, and userinfo ("origin@evil") cannot fake the host — unit tests on
// the policy plus one httptest proof that the userinfo redirect never
// reaches the other server.
// =============================================================================

package platformxe

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func redirectReq(t *testing.T, raw string) *http.Request {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse %q: %v", raw, err)
	}
	return &http.Request{URL: u}
}

func TestRedirectPolicyOriginComparison(t *testing.T) {
	cases := []struct {
		name, from, to string
		follow         bool
	}{
		{"explicit default https port", "https://platformxe.com/a", "https://platformxe.com:443/b", true},
		{"explicit default http port", "http://platformxe.com:80/a", "http://platformxe.com/b", true},
		{"host case differs", "https://PlatformXe.com/a", "https://platformxe.com/b", true},
		{"different port", "https://platformxe.com/a", "https://platformxe.com:8443/b", false},
		{"https to http", "https://platformxe.com/a", "http://platformxe.com/b", false},
		{"other host", "https://platformxe.com/a", "https://evil.example/b", false},
		{"same IPv6 literal, two spellings", "http://[::1]:8080/a", "http://[0:0:0:0:0:0:0:1]:8080/b", true},
		{"IPv6 different port", "http://[::1]:8080/a", "http://[::1]:8081/b", false},
		{"userinfo cannot fake the host", "https://platformxe.com/a", "https://platformxe.com@evil.example/b", false},
		{"userinfo on the same host", "https://platformxe.com/a", "https://user@platformxe.com/b", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := sameOriginRedirectsOnly(redirectReq(t, tc.to), []*http.Request{redirectReq(t, tc.from)})
			if tc.follow && err != nil {
				t.Fatalf("expected the redirect to be followed, got %v", err)
			}
			if !tc.follow && err != http.ErrUseLastResponse {
				t.Fatalf("expected ErrUseLastResponse, got %v", err)
			}
		})
	}
}

func TestUserinfoRedirectDoesNotReachOtherServer(t *testing.T) {
	otherHits := 0
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		otherHits++
		w.Write([]byte(`{"success":true,"data":{}}`))
	}))
	t.Cleanup(other.Close)

	var origin *httptest.Server
	origin = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originHost := strings.TrimPrefix(origin.URL, "http://")
		otherHost := strings.TrimPrefix(other.URL, "http://")
		http.Redirect(w, r, "http://"+originHost+"@"+otherHost+"/x", http.StatusFound)
	}))
	t.Cleanup(origin.Close)

	c := NewClient(ClientConfig{APIKey: redirectTestKey, BaseURL: origin.URL, DisableRetries: true})
	_, err := c.doRequest("GET", "/x", nil, nil)
	if asAPIError(t, err).StatusCode != http.StatusFound {
		t.Fatalf("expected the 302 to reach the caller")
	}
	if otherHits != 0 {
		t.Fatalf("userinfo redirect reached the other server (%d hits)", otherHits)
	}
}

const redirectTestKey = "pxk_test_redirect_key"

func TestCrossOriginRedirectIsNotFollowedAndKeyIsNotSent(t *testing.T) {
	otherHits := 0
	otherSawKey := false
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		otherHits++
		if r.Header.Get("x-api-key") != "" {
			otherSawKey = true
		}
		w.Write([]byte(`{"success":true,"data":{}}`))
	}))
	t.Cleanup(other.Close)

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+"/api/v1/whoami", http.StatusFound)
	}))
	t.Cleanup(origin.Close)

	c := NewClient(ClientConfig{APIKey: redirectTestKey, BaseURL: origin.URL, DisableRetries: true})
	_, err := c.doRequest("GET", "/api/v1/whoami", nil, nil)

	apiErr := asAPIError(t, err)
	if apiErr.StatusCode != http.StatusFound {
		t.Fatalf("expected the 302 to reach the caller, got status %d", apiErr.StatusCode)
	}
	if otherHits != 0 || otherSawKey {
		t.Fatalf("redirect target was contacted (hits=%d, sawKey=%v)", otherHits, otherSawKey)
	}
}

func TestSameOriginRedirectIsFollowedWithKey(t *testing.T) {
	var keyOnTarget string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/old" {
			http.Redirect(w, r, "/new", http.StatusMovedPermanently)
			return
		}
		keyOnTarget = r.Header.Get("x-api-key")
		w.Write([]byte(`{"success":true,"data":{"ok":true}}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(ClientConfig{APIKey: redirectTestKey, BaseURL: srv.URL, DisableRetries: true})
	data, err := c.doRequest("GET", "/old", nil, nil)
	if err != nil {
		t.Fatalf("same-origin redirect should be followed: %v", err)
	}
	if data["ok"] != true {
		t.Fatalf("unexpected data: %v", data)
	}
	if keyOnTarget != redirectTestKey {
		t.Fatalf("same-origin target should receive the key, got %q", keyOnTarget)
	}
}

func TestSameOriginRedirectChainStopsAfterThreeHops(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		http.Redirect(w, r, "/loop", http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(ClientConfig{APIKey: redirectTestKey, BaseURL: srv.URL, DisableRetries: true})
	_, err := c.doRequest("GET", "/loop", nil, nil)

	apiErr := asAPIError(t, err)
	if apiErr.StatusCode != http.StatusFound {
		t.Fatalf("expected the last 302, got status %d", apiErr.StatusCode)
	}
	// The original request plus three followed hops.
	if hits != 4 {
		t.Fatalf("expected 4 requests (1 + 3 hops), got %d", hits)
	}
}

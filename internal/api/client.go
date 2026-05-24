// Package api is a dependency-free client for the Hack The Box API (v4 and v5),
// built on the standard library only.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

// apiBaseURL is the root of the HTB API. The version segment ("v4"/"v5") is
// appended per request because HTB spreads endpoints across both versions.
const apiBaseURL = "https://labs.hackthebox.com/api"

// minRequestInterval paces requests so that concurrent refreshes don't burst
// HTB's API into a rate-limit. The official SDK rate-limits too; we do it with
// a simple serialized minimum gap.
const minRequestInterval = 130 * time.Millisecond

// Client is an authenticated HTB API client. It is safe for concurrent use.
type Client struct {
	token string
	http  *http.Client

	mu   sync.Mutex // serializes pacing
	last time.Time  // time of the last dispatched request
}

// pace blocks until at least minRequestInterval has elapsed since the previous
// request, smoothing concurrent callers into a steady, rate-limit-safe stream.
func (c *Client) pace() {
	c.mu.Lock()
	if wait := minRequestInterval - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
	c.mu.Unlock()
}

// NewClient creates a new HTB API client for the given API token.
func NewClient(token string) *Client {
	return &Client{
		token: strings.TrimSpace(token),
		http:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Token returns the raw API token (used, for example, to parse the JWT subject).
func (c *Client) Token() string { return c.token }

// APIError is returned for any non-2xx response from the HTB API.
type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string {
	body := strings.TrimSpace(e.Body)
	if len(body) > 300 {
		body = body[:300] + "…"
	}
	return fmt.Sprintf("HTB API %d: %s", e.Status, body)
}

// FlagResponse is the common shape returned by flag-submission endpoints
// (machines, challenges, sherlocks).
type FlagResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// doRequest performs an authenticated request against {base}/{version}{path}.
// version is "v4" or "v5"; path must begin with "/". query and body may be nil.
func (c *Client) doRequest(method, version, path string, query url.Values, body any) ([]byte, error) {
	endpoint := apiBaseURL + "/" + version + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var rawBody []byte
	if body != nil {
		var err error
		if rawBody, err = json.Marshal(body); err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
	}

	const maxAttempts = 3
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		c.pace()

		var reader io.Reader
		if rawBody != nil {
			reader = bytes.NewReader(rawBody)
		}
		req, err := http.NewRequest(method, endpoint, reader)
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "htb-tray/2.0")
		if rawBody != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request %s %s: %w", method, path, err)
		}
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read response: %w", readErr)
		}

		// Back off and retry only on an explicit 429. (Pacing is what prevents
		// bursts; retrying a 403 would amplify an active rate-limit.)
		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = &APIError{Status: resp.StatusCode, Body: string(data)}
			delay := time.Duration(attempt+1) * 800 * time.Millisecond
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, e := strconv.Atoi(strings.TrimSpace(ra)); e == nil && secs > 0 {
					delay = time.Duration(secs) * time.Second
				}
			}
			time.Sleep(delay)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, &APIError{Status: resp.StatusCode, Body: string(data)}
		}
		return data, nil
	}
	return nil, lastErr
}

// getJSON issues a GET and unmarshals the JSON response into out (out may be nil).
func (c *Client) getJSON(version, path string, query url.Values, out any) error {
	data, err := c.doRequest(http.MethodGet, version, path, query, nil)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode %s response: %w", path, err)
	}
	return nil
}

// sendJSON issues a POST/DELETE/PUT with an optional JSON body and unmarshals
// the response into out (body and out may each be nil).
func (c *Client) sendJSON(method, version, path string, body, out any) error {
	data, err := c.doRequest(method, version, path, nil, body)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode %s response: %w", path, err)
	}
	return nil
}

// getRaw issues a GET and returns the raw (possibly binary) response body,
// used for file downloads (VPN configs, challenge files).
func (c *Client) getRaw(version, p string) ([]byte, error) {
	return c.doRequest(http.MethodGet, version, p, nil, nil)
}

// FetchURL downloads an arbitrary absolute URL (not an API path) without
// attaching the API token - used for pre-signed download links and public
// storage assets (avatars), where an extra Authorization header could break a
// signed request. It returns the body and a suggested filename derived from the
// Content-Disposition header or the URL path.
func (c *Client) FetchURL(rawurl string) (data []byte, filename string, err error) {
	req, err := http.NewRequest(http.MethodGet, rawurl, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "htb-tray/2.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", &APIError{Status: resp.StatusCode, Body: string(data)}
	}
	return data, filenameFromResponse(resp, rawurl), nil
}

// filenameFromResponse picks a download filename from Content-Disposition, then
// falls back to the last path segment of the URL.
func filenameFromResponse(resp *http.Response, rawurl string) string {
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			if fn := params["filename"]; fn != "" {
				return fn
			}
		}
	}
	if u, err := url.Parse(rawurl); err == nil {
		if b := path.Base(u.Path); b != "" && b != "/" && b != "." {
			return b
		}
	}
	return ""
}

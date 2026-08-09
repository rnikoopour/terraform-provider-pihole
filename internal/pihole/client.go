package pihole

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	BaseURL    string
	password   string
	httpClient *http.Client
	sid        string
	mu         sync.Mutex
}

type requestResult struct {
	Body       []byte
	StatusCode int
}

func NewClient(baseURL, password string, insecure bool) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}, //nolint:gosec
	}
	return &Client{
		BaseURL:  baseURL,
		password: password,
		// Client.Request holds c.mu for the full call, serializing every
		// operation against this provider instance. Without a timeout here,
		// a single request that hangs server-side (FTL wedged, network
		// blip) blocks every other resource for this instance forever
		// instead of failing fast. 5 minutes accommodates a real gravity
		// update (fetches + processes every configured blocklist, which
		// can legitimately take minutes) while still catching genuine
		// hangs.
		httpClient: &http.Client{Transport: transport, Timeout: 5 * time.Minute},
	}
}

func (c *Client) logout() {
	if c.sid == "" {
		return
	}
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+"/api/auth", nil)
	if err == nil {
		req.Header.Set("X-FTL-SID", c.sid)
		resp, err := c.httpClient.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}
	c.sid = ""
}

func (c *Client) authenticate() error {
	body, err := json.Marshal(map[string]string{"password": c.password})
	if err != nil {
		return fmt.Errorf("marshal auth body: %w", err)
	}

	var lastErr error
	for attempt := range 5 {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		resp, err := c.httpClient.Post(c.BaseURL+"/api/auth", "application/json", bytes.NewBuffer(body))
		if err != nil {
			lastErr = fmt.Errorf("auth request to %s: %w", c.BaseURL, err)
			continue
		}

		var result struct {
			Session struct {
				SID   string `json:"sid"`
				Valid bool   `json:"valid"`
			} `json:"session"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("decode auth response: %w", err)
			continue
		}
		if !result.Session.Valid {
			lastErr = fmt.Errorf("authentication failed for %s", c.BaseURL)
			continue
		}
		c.sid = result.Session.SID
		return nil
	}
	return lastErr
}

func (c *Client) Request(method, path string, body interface{}) (*requestResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.doRequest(method, path, body, true)
}

func (c *Client) doRequest(method, path string, body interface{}, retry bool) (*requestResult, error) {
	if c.sid == "" {
		if err := c.authenticate(); err != nil {
			return nil, err
		}
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-FTL-SID", c.sid)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request to %s%s: %w", c.BaseURL, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized && retry {
		c.logout()
		return c.doRequest(method, path, body, false)
	}

	return &requestResult{Body: respBody, StatusCode: resp.StatusCode}, nil
}

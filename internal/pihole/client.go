package pihole

import (
	"bytes"
	"context"
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

// NewClient builds a client with a bounded HTTP timeout. Request holds c.mu for
// the full call, serializing every operation against this provider instance;
// without a timeout a server-side hang would block every other resource for
// this instance forever. Five minutes accommodates a real gravity update
// (fetching + processing every blocklist) while still catching genuine hangs.
func NewClient(in *NewClientInput) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: in.Insecure}, //nolint:gosec
	}
	c := &Client{
		BaseURL:    in.BaseURL,
		password:   in.Password,
		httpClient: &http.Client{Transport: transport, Timeout: 5 * time.Minute},
	}

	registryMu.Lock()
	registry = append(registry, c)
	registryMu.Unlock()

	return c
}

var (
	registryMu sync.Mutex
	registry   []*Client
)

// CloseAll logs out every client created in this process. Terraform starts one
// provider instance per Pi-hole server, and each session it opens counts
// against Pi-hole's API seat limit until it expires, so logging out at shutdown
// frees the seats immediately.
func CloseAll(ctx context.Context) {
	registryMu.Lock()
	defer registryMu.Unlock()

	for _, c := range registry {
		c.mu.Lock()
		c.logout(ctx)
		c.mu.Unlock()
	}
	registry = nil
}

func (c *Client) logout(ctx context.Context) {
	if c.sid == "" {
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.BaseURL+"/api/auth", nil)
	if err == nil {
		req.Header.Set("X-FTL-SID", c.sid)
		resp, err := c.httpClient.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}
	c.sid = ""
}

func (c *Client) authenticate(ctx context.Context) error {
	body, err := json.Marshal(map[string]string{"password": c.password})
	if err != nil {
		return fmt.Errorf("marshal auth body: %w", err)
	}

	var lastErr error
	for attempt := range 5 {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/auth", bytes.NewBuffer(body))
		if err != nil {
			lastErr = fmt.Errorf("create auth request to %s: %w", c.BaseURL, err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
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

// Request serializes all calls against this client instance so a shared session
// id is never raced.
func (c *Client) Request(in *RequestInput) (*RequestOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.doRequest(&doRequestInput{
		Ctx:    in.Ctx,
		Method: in.Method,
		Path:   in.Path,
		Body:   in.Body,
		Retry:  true,
	})
}

// doRequestInput is the package-private form of RequestInput plus the
// single-retry flag used after a session expiry.
type doRequestInput struct {
	Ctx    context.Context
	Method string
	Path   string
	Body   any
	Retry  bool
}

func (c *Client) doRequest(in *doRequestInput) (*RequestOutput, error) {
	ctx := ctxOrBackground(in.Ctx)
	if c.sid == "" {
		if err := c.authenticate(ctx); err != nil {
			return nil, err
		}
	}

	var bodyReader io.Reader
	if in.Body != nil {
		b, err := json.Marshal(in.Body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequestWithContext(ctx, in.Method, c.BaseURL+in.Path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-FTL-SID", c.sid)
	if in.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request to %s%s: %w", c.BaseURL, in.Path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized && in.Retry {
		c.logout(ctx)
		return c.doRequest(&doRequestInput{
			Ctx:    ctx,
			Method: in.Method,
			Path:   in.Path,
			Body:   in.Body,
			Retry:  false,
		})
	}

	return &RequestOutput{StatusCode: resp.StatusCode, Body: respBody}, nil
}

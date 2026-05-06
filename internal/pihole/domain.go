package pihole

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Domain struct {
	ID      int    `json:"id,omitempty"`
	Domain  string `json:"domain"`
	Type    string `json:"type"` // "allow" or "deny"
	Kind    string `json:"kind"` // "exact" or "regex"
	Enabled bool   `json:"enabled"`
	Comment string `json:"comment"`
	Groups  []int  `json:"groups,omitempty"`
}

func (c *Client) GetDomain(domain, domainType, kind string) (*Domain, error) {
	result, err := c.Request(http.MethodGet, "/api/domains", nil)
	if err != nil {
		return nil, err
	}
	if result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /api/domains: status %d: %s", result.StatusCode, string(result.Body))
	}
	var resp struct {
		Domains []Domain `json:"domains"`
	}
	if err := json.Unmarshal(result.Body, &resp); err != nil {
		return nil, fmt.Errorf("parse domains response: %w", err)
	}
	for i := range resp.Domains {
		d := &resp.Domains[i]
		if d.Domain == domain && d.Type == domainType && d.Kind == kind {
			return d, nil
		}
	}
	return nil, nil
}

func (c *Client) CreateDomain(d Domain) (*Domain, error) {
	path := fmt.Sprintf("/api/domains/%s/%s", d.Type, d.Kind)
	result, err := c.Request(http.MethodPost, path, d)
	if err != nil {
		return nil, err
	}
	if result.StatusCode != http.StatusCreated && result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("POST %s: status %d: %s", path, result.StatusCode, string(result.Body))
	}
	var resp struct {
		Domains []Domain `json:"domains"`
	}
	if err := json.Unmarshal(result.Body, &resp); err != nil {
		return nil, fmt.Errorf("parse create domain response: %w", err)
	}
	if len(resp.Domains) == 0 {
		return nil, fmt.Errorf("create domain response contained no domains")
	}
	return &resp.Domains[0], nil
}

func (c *Client) UpdateDomain(id int, d Domain) error {
	result, err := c.Request(http.MethodPut, fmt.Sprintf("/api/domains/%d", id), d)
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusOK {
		return fmt.Errorf("PUT /api/domains/%d: status %d: %s", id, result.StatusCode, string(result.Body))
	}
	return nil
}

func (c *Client) DeleteDomain(domainType, kind, domain string) error {
	path := fmt.Sprintf("/api/domains/%s/%s/%s", domainType, kind, url.PathEscape(domain))
	result, err := c.Request(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	if result.StatusCode == http.StatusNotFound {
		return nil
	}
	if result.StatusCode != http.StatusNoContent && result.StatusCode != http.StatusOK {
		return fmt.Errorf("DELETE %s: status %d: %s", path, result.StatusCode, string(result.Body))
	}
	return nil
}

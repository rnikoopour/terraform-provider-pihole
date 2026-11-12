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

// GetDomain returns a nil Domain in the output when no matching entry exists.
func (c *Client) GetDomain(in *GetDomainInput) (*GetDomainOutput, error) {
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodGet, Path: "/api/domains"})
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
		if d.Domain == in.Domain && d.Type == in.Type && d.Kind == in.Kind {
			return &GetDomainOutput{Domain: d}, nil
		}
	}
	return &GetDomainOutput{}, nil
}

func (c *Client) CreateDomain(in *CreateDomainInput) (*CreateDomainOutput, error) {
	path := fmt.Sprintf("/api/domains/%s/%s", in.Domain.Type, in.Domain.Kind)
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodPost, Path: path, Body: in.Domain})
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
	return &CreateDomainOutput{Domain: &resp.Domains[0]}, nil
}

func (c *Client) UpdateDomain(in *UpdateDomainInput) error {
	path := fmt.Sprintf("/api/domains/%d", in.ID)
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodPut, Path: path, Body: in.Domain})
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusOK {
		return fmt.Errorf("PUT %s: status %d: %s", path, result.StatusCode, string(result.Body))
	}
	return nil
}

func (c *Client) DeleteDomain(in *DeleteDomainInput) error {
	path := fmt.Sprintf("/api/domains/%s/%s/%s", in.Type, in.Kind, url.PathEscape(in.Domain))
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodDelete, Path: path})
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

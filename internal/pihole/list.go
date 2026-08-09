package pihole

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type List struct {
	ID      int    `json:"id,omitempty"`
	Address string `json:"address"`
	Type    string `json:"type"` // "block" or "allow"
	Enabled bool   `json:"enabled"`
	Comment string `json:"comment"`
	Groups  []int  `json:"groups,omitempty"`
}

func (c *Client) GetListByAddress(address string) (*List, error) {
	result, err := c.Request(http.MethodGet, "/api/lists", nil)
	if err != nil {
		return nil, err
	}
	if result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /api/lists: status %d: %s", result.StatusCode, string(result.Body))
	}
	var resp struct {
		Lists []List `json:"lists"`
	}
	if err := json.Unmarshal(result.Body, &resp); err != nil {
		return nil, fmt.Errorf("parse lists response: %w", err)
	}
	for i := range resp.Lists {
		if resp.Lists[i].Address == address {
			return &resp.Lists[i], nil
		}
	}
	return nil, nil
}

func (c *Client) CreateList(list List) (*List, error) {
	result, err := c.Request(http.MethodPost, "/api/lists?type="+list.Type, list)
	if err != nil {
		return nil, err
	}
	if result.StatusCode != http.StatusCreated && result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("POST /api/lists: status %d: %s", result.StatusCode, string(result.Body))
	}
	var resp struct {
		Lists []List `json:"lists"`
	}
	if err := json.Unmarshal(result.Body, &resp); err != nil {
		return nil, fmt.Errorf("parse create list response: %w", err)
	}
	if len(resp.Lists) == 0 {
		return nil, fmt.Errorf("create list response contained no lists")
	}
	return &resp.Lists[0], nil
}

func (c *Client) UpdateList(_ int, list List) error {
	result, err := c.Request(http.MethodPut, fmt.Sprintf("/api/lists/%s?type=%s", url.PathEscape(list.Address), list.Type), list)
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusOK {
		return fmt.Errorf("PUT /api/lists/%s: status %d: %s", list.Address, result.StatusCode, string(result.Body))
	}
	return nil
}

func (c *Client) TriggerGravity() error {
	result, err := c.Request(http.MethodPost, "/api/action/gravity", nil)
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusOK && result.StatusCode != http.StatusNoContent {
		return fmt.Errorf("POST /api/action/gravity: status %d", result.StatusCode)
	}
	return nil
}

func (c *Client) DeleteList(address string, listType string) error {
	path := fmt.Sprintf("/api/lists/%s?type=%s", url.PathEscape(address), listType)
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

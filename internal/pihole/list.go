package pihole

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type List struct {
	ID      int    `json:"id,omitempty"`
	Address string `json:"address"`
	Type    string `json:"type"` // "block" or "allow"
	Enabled bool   `json:"enabled"`
	Comment string `json:"comment"`
	Groups  []int  `json:"groups,omitempty"`
}

// GetListByAddress returns a nil List in the output when no matching list exists.
func (c *Client) GetListByAddress(in *GetListByAddressInput) (*GetListByAddressOutput, error) {
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodGet, Path: "/api/lists"})
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
		if resp.Lists[i].Address == in.Address {
			return &GetListByAddressOutput{List: &resp.Lists[i]}, nil
		}
	}
	return &GetListByAddressOutput{}, nil
}

func (c *Client) CreateList(in *CreateListInput) (*CreateListOutput, error) {
	result, err := c.Request(&RequestInput{
		Ctx:    in.Ctx,
		Method: http.MethodPost,
		Path:   "/api/lists?type=" + in.List.Type,
		Body:   in.List,
	})
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
	return &CreateListOutput{List: &resp.Lists[0]}, nil
}

func (c *Client) UpdateList(in *UpdateListInput) error {
	path := fmt.Sprintf("/api/lists/%s?type=%s", escapePathSegment(in.List.Address), in.List.Type)
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodPut, Path: path, Body: in.List})
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusOK {
		return fmt.Errorf("PUT /api/lists/%s: status %d: %s", in.List.Address, result.StatusCode, string(result.Body))
	}
	return nil
}

func (c *Client) TriggerGravity(in *TriggerGravityInput) error {
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodPost, Path: "/api/action/gravity"})
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusOK && result.StatusCode != http.StatusNoContent {
		return fmt.Errorf("POST /api/action/gravity: status %d", result.StatusCode)
	}
	return nil
}

func (c *Client) DeleteList(in *DeleteListInput) error {
	path := fmt.Sprintf("/api/lists/%s?type=%s", escapePathSegment(in.Address), in.Type)
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

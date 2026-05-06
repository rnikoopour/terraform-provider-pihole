package pihole

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Group struct {
	ID      int    `json:"id,omitempty"`
	Name    string `json:"name"`
	Comment string `json:"comment"`
	Enabled bool   `json:"enabled"`
}

func (c *Client) ListGroups() ([]Group, error) {
	result, err := c.Request(http.MethodGet, "/api/groups", nil)
	if err != nil {
		return nil, err
	}
	if result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /api/groups: status %d: %s", result.StatusCode, string(result.Body))
	}
	var resp struct {
		Groups []Group `json:"groups"`
	}
	if err := json.Unmarshal(result.Body, &resp); err != nil {
		return nil, fmt.Errorf("parse groups response: %w", err)
	}
	return resp.Groups, nil
}

func (c *Client) GetGroup(name string) (*Group, error) {
	groups, err := c.ListGroups()
	if err != nil {
		return nil, err
	}
	for i := range groups {
		if groups[i].Name == name {
			return &groups[i], nil
		}
	}
	return nil, nil
}

func (c *Client) CreateGroup(group Group) (*Group, error) {
	result, err := c.Request(http.MethodPost, "/api/groups", group)
	if err != nil {
		return nil, err
	}
	if result.StatusCode != http.StatusCreated && result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("POST /api/groups: status %d: %s", result.StatusCode, string(result.Body))
	}
	var resp struct {
		Group Group `json:"group"`
	}
	if err := json.Unmarshal(result.Body, &resp); err != nil {
		return nil, fmt.Errorf("parse create group response: %w", err)
	}
	return &resp.Group, nil
}

func (c *Client) UpdateGroup(name string, group Group) error {
	result, err := c.Request(http.MethodPut, "/api/groups/"+name, group)
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusOK {
		return fmt.Errorf("PUT /api/groups/%s: status %d: %s", name, result.StatusCode, string(result.Body))
	}
	return nil
}

func (c *Client) DeleteGroup(name string) error {
	result, err := c.Request(http.MethodDelete, "/api/groups/"+name, nil)
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusNoContent && result.StatusCode != http.StatusOK {
		return fmt.Errorf("DELETE /api/groups/%s: status %d: %s", name, result.StatusCode, string(result.Body))
	}
	return nil
}

// GroupIDByName resolves a group name to its integer ID on this server.
func (c *Client) GroupIDByName(name string) (int, error) {
	group, err := c.GetGroup(name)
	if err != nil {
		return 0, err
	}
	if group == nil {
		return 0, fmt.Errorf("group %q not found", name)
	}
	return group.ID, nil
}

// GroupNameByID resolves a group integer ID to its name on this server.
func (c *Client) GroupNameByID(id int) (string, error) {
	groups, err := c.ListGroups()
	if err != nil {
		return "", err
	}
	for _, g := range groups {
		if g.ID == id {
			return g.Name, nil
		}
	}
	return "", fmt.Errorf("group id %d not found", id)
}

// ResolveGroupNames converts a slice of group names to integer IDs for this server.
func (c *Client) ResolveGroupNames(names []string) ([]int, error) {
	ids := make([]int, 0, len(names))
	for _, name := range names {
		id, err := c.GroupIDByName(name)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// ResolveGroupIDs converts a slice of integer group IDs to names for this server.
func (c *Client) ResolveGroupIDs(ids []int) ([]string, error) {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		name, err := c.GroupNameByID(id)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

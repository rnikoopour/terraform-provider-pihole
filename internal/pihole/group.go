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

func (c *Client) ListGroups(in *ListGroupsInput) (*ListGroupsOutput, error) {
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodGet, Path: "/api/groups"})
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
	return &ListGroupsOutput{Groups: resp.Groups}, nil
}

// GetGroup returns a nil Group in the output when the name does not exist.
func (c *Client) GetGroup(in *GetGroupInput) (*GetGroupOutput, error) {
	groups, err := c.ListGroups(&ListGroupsInput{Ctx: in.Ctx})
	if err != nil {
		return nil, err
	}
	for i := range groups.Groups {
		if groups.Groups[i].Name == in.Name {
			return &GetGroupOutput{Group: &groups.Groups[i]}, nil
		}
	}
	return &GetGroupOutput{}, nil
}

func (c *Client) CreateGroup(in *CreateGroupInput) (*CreateGroupOutput, error) {
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodPost, Path: "/api/groups", Body: in.Group})
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
	return &CreateGroupOutput{Group: &resp.Group}, nil
}

func (c *Client) UpdateGroup(in *UpdateGroupInput) error {
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodPut, Path: "/api/groups/" + in.Name, Body: in.Group})
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusOK {
		return fmt.Errorf("PUT /api/groups/%s: status %d: %s", in.Name, result.StatusCode, string(result.Body))
	}
	return nil
}

func (c *Client) DeleteGroup(in *DeleteGroupInput) error {
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodDelete, Path: "/api/groups/" + in.Name})
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusNoContent && result.StatusCode != http.StatusOK {
		return fmt.Errorf("DELETE /api/groups/%s: status %d: %s", in.Name, result.StatusCode, string(result.Body))
	}
	return nil
}

// GroupIDByName resolves a group name to its integer ID on this server.
func (c *Client) GroupIDByName(in *GroupIDByNameInput) (*GroupIDByNameOutput, error) {
	out, err := c.GetGroup(&GetGroupInput{Ctx: in.Ctx, Name: in.Name})
	if err != nil {
		return nil, err
	}
	if out.Group == nil {
		return nil, fmt.Errorf("group %q not found", in.Name)
	}
	return &GroupIDByNameOutput{ID: out.Group.ID}, nil
}

// GroupNameByID resolves a group integer ID to its name on this server.
func (c *Client) GroupNameByID(in *GroupNameByIDInput) (*GroupNameByIDOutput, error) {
	groups, err := c.ListGroups(&ListGroupsInput{Ctx: in.Ctx})
	if err != nil {
		return nil, err
	}
	for _, g := range groups.Groups {
		if g.ID == in.ID {
			return &GroupNameByIDOutput{Name: g.Name}, nil
		}
	}
	return nil, fmt.Errorf("group id %d not found", in.ID)
}

// ResolveGroupNames converts group names to integer IDs for this server.
func (c *Client) ResolveGroupNames(in *ResolveGroupNamesInput) (*ResolveGroupNamesOutput, error) {
	ids := make([]int, 0, len(in.Names))
	for _, name := range in.Names {
		id, err := c.GroupIDByName(&GroupIDByNameInput{Ctx: in.Ctx, Name: name})
		if err != nil {
			return nil, err
		}
		ids = append(ids, id.ID)
	}
	return &ResolveGroupNamesOutput{IDs: ids}, nil
}

// ResolveGroupIDs converts integer group IDs to names for this server.
func (c *Client) ResolveGroupIDs(in *ResolveGroupIDsInput) (*ResolveGroupIDsOutput, error) {
	names := make([]string, 0, len(in.IDs))
	for _, id := range in.IDs {
		name, err := c.GroupNameByID(&GroupNameByIDInput{Ctx: in.Ctx, ID: id})
		if err != nil {
			return nil, err
		}
		names = append(names, name.Name)
	}
	return &ResolveGroupIDsOutput{Names: names}, nil
}

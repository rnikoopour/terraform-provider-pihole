package pihole

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// ClientEntry is a Pi-hole client record used for group-scoped blocking.
// Groups holds numeric group IDs on the wire; the provider layer translates
// to and from group names for readability in Terraform configuration.
type ClientEntry struct {
	ID      int    `json:"id,omitempty"`
	Client  string `json:"client"`
	Name    string `json:"name,omitempty"`
	Comment string `json:"comment"`
	Groups  []int  `json:"groups"`
}

// clientResponse accommodates the two shapes Pi-hole v6 returns for client
// endpoints: a single "client" object and a "clients" array.
type clientResponse struct {
	Client  *ClientEntry  `json:"client"`
	Clients []ClientEntry `json:"clients"`
}

func (r clientResponse) pick(want string) *ClientEntry {
	if r.Client != nil {
		return r.Client
	}
	for i := range r.Clients {
		if r.Clients[i].Client == want {
			return &r.Clients[i]
		}
	}
	if len(r.Clients) > 0 {
		return &r.Clients[0]
	}
	return nil
}

func (c *Client) ListClients(in *ListClientsInput) (*ListClientsOutput, error) {
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodGet, Path: "/api/clients"})
	if err != nil {
		return nil, err
	}
	if result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /api/clients: status %d: %s", result.StatusCode, string(result.Body))
	}
	var resp clientResponse
	if err := json.Unmarshal(result.Body, &resp); err != nil {
		return nil, fmt.Errorf("parse clients response: %w", err)
	}
	if resp.Client != nil {
		return &ListClientsOutput{Clients: []ClientEntry{*resp.Client}}, nil
	}
	return &ListClientsOutput{Clients: resp.Clients}, nil
}

// GetClient returns a nil Entry in the output when no explicit record exists;
// such a client falls back to the Default group.
func (c *Client) GetClient(in *GetClientInput) (*GetClientOutput, error) {
	out, err := c.ListClients(&ListClientsInput{Ctx: in.Ctx})
	if err != nil {
		return nil, err
	}
	for i := range out.Clients {
		if out.Clients[i].Client == in.Addr {
			return &GetClientOutput{Entry: &out.Clients[i]}, nil
		}
	}
	return &GetClientOutput{}, nil
}

func (c *Client) CreateClient(in *CreateClientInput) (*CreateClientOutput, error) {
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodPost, Path: "/api/clients", Body: in.Entry})
	if err != nil {
		return nil, err
	}
	if result.StatusCode != http.StatusCreated && result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("POST /api/clients: status %d: %s", result.StatusCode, string(result.Body))
	}
	var resp clientResponse
	if err := json.Unmarshal(result.Body, &resp); err != nil {
		return nil, fmt.Errorf("parse create client response: %w", err)
	}
	created := resp.pick(in.Entry.Client)
	if created == nil {
		return nil, fmt.Errorf("POST /api/clients: response did not include the created client")
	}
	return &CreateClientOutput{Entry: created}, nil
}

func (c *Client) UpdateClient(in *UpdateClientInput) error {
	path := "/api/clients/" + url.PathEscape(in.Addr)
	// Only comment and groups are mutable; the identifier lives in the path.
	payload := map[string]any{
		"comment": in.Entry.Comment,
		"groups":  in.Entry.Groups,
	}
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodPut, Path: path, Body: payload})
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusOK {
		return fmt.Errorf("PUT %s: status %d: %s", path, result.StatusCode, string(result.Body))
	}
	return nil
}

func (c *Client) DeleteClient(in *DeleteClientInput) error {
	path := "/api/clients/" + url.PathEscape(in.Addr)
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

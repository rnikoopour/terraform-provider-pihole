package pihole

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// DNSRecord represents a custom local DNS A/AAAA record ("IP hostname").
type DNSRecord struct {
	IP       string
	Hostname string
}

func (r DNSRecord) String() string {
	return r.IP + " " + r.Hostname
}

type parseHostOutput struct {
	IP       string
	Hostname string
}

func (c *Client) ListDNSRecords(in *ListDNSRecordsInput) (*ListDNSRecordsOutput, error) {
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodGet, Path: "/api/config/dns/hosts"})
	if err != nil {
		return nil, err
	}
	if result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /api/config/dns/hosts: status %d: %s", result.StatusCode, string(result.Body))
	}
	var resp struct {
		Config struct {
			DNS struct {
				Hosts []string `json:"hosts"`
			} `json:"dns"`
		} `json:"config"`
	}
	if err := json.Unmarshal(result.Body, &resp); err != nil {
		return nil, fmt.Errorf("parse dns hosts response: %w", err)
	}
	records := make([]DNSRecord, 0, len(resp.Config.DNS.Hosts))
	for _, h := range resp.Config.DNS.Hosts {
		out, ok := parseHost(h)
		if ok {
			records = append(records, DNSRecord{IP: out.IP, Hostname: out.Hostname})
		}
	}
	return &ListDNSRecordsOutput{Records: records}, nil
}

// GetDNSRecord matches the exact IP/hostname pair; a nil Record means no match.
func (c *Client) GetDNSRecord(in *GetDNSRecordInput) (*GetDNSRecordOutput, error) {
	out, err := c.ListDNSRecords(&ListDNSRecordsInput{Ctx: in.Ctx})
	if err != nil {
		return nil, err
	}
	for i := range out.Records {
		if out.Records[i].IP == in.IP && out.Records[i].Hostname == in.Hostname {
			return &GetDNSRecordOutput{Record: &out.Records[i]}, nil
		}
	}
	return &GetDNSRecordOutput{}, nil
}

func (c *Client) CreateDNSRecord(in *CreateDNSRecordInput) error {
	path := "/api/config/dns/hosts/" + escapePathSegment(in.Record.String())
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodPut, Path: path})
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusNoContent && result.StatusCode != http.StatusOK && result.StatusCode != http.StatusCreated {
		return fmt.Errorf("PUT %s: status %d: %s", path, result.StatusCode, string(result.Body))
	}
	return nil
}

func (c *Client) DeleteDNSRecord(in *DeleteDNSRecordInput) error {
	// Pi-hole v6 uses dns%2Fhosts (encoded slash) for the delete path.
	path := "/api/config/dns%2Fhosts/" + escapePathSegment(in.Record.String())
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodDelete, Path: path})
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusNoContent && result.StatusCode != http.StatusOK {
		return fmt.Errorf("DELETE %s: status %d: %s", path, result.StatusCode, string(result.Body))
	}
	return nil
}

func parseHost(s string) (parseHostOutput, bool) {
	for i, c := range s {
		if c == ' ' {
			return parseHostOutput{IP: s[:i], Hostname: s[i+1:]}, true
		}
	}
	return parseHostOutput{}, false
}

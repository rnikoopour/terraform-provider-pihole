package pihole

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

func (c *Client) ListDNSRecords() ([]DNSRecord, error) {
	result, err := c.Request(http.MethodGet, "/api/config/dns/hosts", nil)
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
	return records, nil
}

func (c *Client) GetDNSRecord(ip, hostname string) (*DNSRecord, error) {
	records, err := c.ListDNSRecords()
	if err != nil {
		return nil, err
	}
	for i := range records {
		if records[i].IP == ip && records[i].Hostname == hostname {
			return &records[i], nil
		}
	}
	return nil, nil
}

func (c *Client) CreateDNSRecord(r DNSRecord) error {
	entry := r.String()
	path := "/api/config/dns/hosts/" + url.PathEscape(entry)
	result, err := c.Request(http.MethodPut, path, nil)
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusNoContent && result.StatusCode != http.StatusOK && result.StatusCode != http.StatusCreated {
		return fmt.Errorf("PUT %s: status %d: %s", path, result.StatusCode, string(result.Body))
	}
	return nil
}

func (c *Client) DeleteDNSRecord(r DNSRecord) error {
	entry := r.String()
	// Pi-hole v6 uses dns%2Fhosts (encoded slash) for the delete path
	path := "/api/config/dns%2Fhosts/" + url.PathEscape(entry)
	result, err := c.Request(http.MethodDelete, path, nil)
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

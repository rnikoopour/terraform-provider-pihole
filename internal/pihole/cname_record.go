package pihole

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// CNAMERecord represents a local CNAME record ("domain,target").
type CNAMERecord struct {
	Domain string
	Target string
}

func (r CNAMERecord) String() string {
	return r.Domain + "," + r.Target
}

func (c *Client) ListCNAMERecords() ([]CNAMERecord, error) {
	result, err := c.Request(http.MethodGet, "/api/config/dns/cnameRecords", nil)
	if err != nil {
		return nil, err
	}
	if result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /api/config/dns/cnameRecords: status %d: %s", result.StatusCode, string(result.Body))
	}
	var resp struct {
		Config struct {
			DNS struct {
				CNAMERecords []string `json:"cnameRecords"`
			} `json:"dns"`
		} `json:"config"`
	}
	if err := json.Unmarshal(result.Body, &resp); err != nil {
		return nil, fmt.Errorf("parse cname records response: %w", err)
	}
	records := make([]CNAMERecord, 0, len(resp.Config.DNS.CNAMERecords))
	for _, entry := range resp.Config.DNS.CNAMERecords {
		parts := strings.SplitN(entry, ",", 2)
		if len(parts) == 2 {
			records = append(records, CNAMERecord{Domain: parts[0], Target: parts[1]})
		}
	}
	return records, nil
}

func (c *Client) GetCNAMERecord(domain string) (*CNAMERecord, error) {
	records, err := c.ListCNAMERecords()
	if err != nil {
		return nil, err
	}
	for i := range records {
		if records[i].Domain == domain {
			return &records[i], nil
		}
	}
	return nil, nil
}

func (c *Client) CreateCNAMERecord(r CNAMERecord) error {
	entry := r.String()
	path := "/api/config/dns/cnameRecords/" + url.PathEscape(entry)
	result, err := c.Request(http.MethodPut, path, nil)
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusNoContent && result.StatusCode != http.StatusOK && result.StatusCode != http.StatusCreated {
		return fmt.Errorf("PUT %s: status %d: %s", path, result.StatusCode, string(result.Body))
	}
	return nil
}

func (c *Client) DeleteCNAMERecord(r CNAMERecord) error {
	entry := r.String()
	path := "/api/config/dns%2FcnameRecords/" + url.PathEscape(entry)
	result, err := c.Request(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusNoContent && result.StatusCode != http.StatusOK {
		return fmt.Errorf("DELETE %s: status %d: %s", path, result.StatusCode, string(result.Body))
	}
	return nil
}

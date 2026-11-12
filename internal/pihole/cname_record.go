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

func (c *Client) ListCNAMERecords(in *ListCNAMERecordsInput) (*ListCNAMERecordsOutput, error) {
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodGet, Path: "/api/config/dns/cnameRecords"})
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
	return &ListCNAMERecordsOutput{Records: records}, nil
}

// GetCNAMERecord matches on the domain; a nil Record means no match.
func (c *Client) GetCNAMERecord(in *GetCNAMERecordInput) (*GetCNAMERecordOutput, error) {
	out, err := c.ListCNAMERecords(&ListCNAMERecordsInput{Ctx: in.Ctx})
	if err != nil {
		return nil, err
	}
	for i := range out.Records {
		if out.Records[i].Domain == in.Domain {
			return &GetCNAMERecordOutput{Record: &out.Records[i]}, nil
		}
	}
	return &GetCNAMERecordOutput{}, nil
}

func (c *Client) CreateCNAMERecord(in *CreateCNAMERecordInput) error {
	path := "/api/config/dns/cnameRecords/" + url.PathEscape(in.Record.String())
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodPut, Path: path})
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusNoContent && result.StatusCode != http.StatusOK && result.StatusCode != http.StatusCreated {
		return fmt.Errorf("PUT %s: status %d: %s", path, result.StatusCode, string(result.Body))
	}
	return nil
}

func (c *Client) DeleteCNAMERecord(in *DeleteCNAMERecordInput) error {
	path := "/api/config/dns%2FcnameRecords/" + url.PathEscape(in.Record.String())
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodDelete, Path: path})
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusNoContent && result.StatusCode != http.StatusOK {
		return fmt.Errorf("DELETE %s: status %d: %s", path, result.StatusCode, string(result.Body))
	}
	return nil
}

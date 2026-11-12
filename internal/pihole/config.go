package pihole

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// DNSConfig holds the syncable subset of Pi-hole DNS configuration.
type DNSConfig struct {
	Upstreams        []string          `json:"upstreams,omitempty"`
	DomainNeeded     bool              `json:"domainNeeded"`
	ExpandHosts      bool              `json:"expandHosts"`
	BogusPriv        bool              `json:"bogusPriv"`
	DNSSEC           bool              `json:"dnssec"`
	QueryLogging     bool              `json:"queryLogging"`
	CNAMEDeepInspect bool              `json:"CNAMEdeepInspect"`
	BlockESNI        bool              `json:"blockESNI"`
	Blocking         DNSBlocking       `json:"blocking"`
	RateLimit        DNSRateLimit      `json:"rateLimit"`
	Cache            DNSCache          `json:"cache"`
	SpecialDomains   DNSSpecialDomains `json:"specialDomains"`
	PiholePTR        string            `json:"piholePTR"`
	HostRecord       string            `json:"hostRecord"`
	ListeningMode    string            `json:"listeningMode"`
	Interface        string            `json:"interface"`
	ReplyWhenBusy    string            `json:"replyWhenBusy"`
	BlockTTL         int               `json:"blockTTL"`
}

type DNSBlocking struct {
	Active bool   `json:"active"`
	Mode   string `json:"mode"` // NULL, NXDOMAIN, NODATA, IP, IP-NODATA-AAAA
	EDNS   string `json:"edns"` // NONE, CODE, TEXT
}

type DNSRateLimit struct {
	Count    int `json:"count"`
	Interval int `json:"interval"`
}

type DNSCache struct {
	Size      int `json:"size"`
	Optimizer int `json:"optimizer"`
}

type DNSSpecialDomains struct {
	MozillaCanary      bool `json:"mozillaCanary"`
	ICloudPrivateRelay bool `json:"iCloudPrivateRelay"`
	DesignatedResolver bool `json:"designatedResolver"`
}

type ServerConfig struct {
	DNS             DNSConfig
	WebserverDomain string
	Theme           string
	Boxed           bool
	PrivacyLevel    int
	MaxDBDays       int
}

func (c *Client) GetServerConfig(in *GetServerConfigInput) (*GetServerConfigOutput, error) {
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodGet, Path: "/api/config"})
	if err != nil {
		return nil, err
	}
	if result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /api/config: status %d: %s", result.StatusCode, string(result.Body))
	}
	var resp struct {
		Config struct {
			DNS       DNSConfig `json:"dns"`
			Webserver struct {
				Domain    string `json:"domain"`
				Interface struct {
					Theme string `json:"theme"`
					Boxed bool   `json:"boxed"`
				} `json:"interface"`
			} `json:"webserver"`
			Misc struct {
				PrivacyLevel int `json:"privacylevel"`
			} `json:"misc"`
			Database struct {
				MaxDBDays int `json:"maxDBdays"`
			} `json:"database"`
		} `json:"config"`
	}
	if err := json.Unmarshal(result.Body, &resp); err != nil {
		return nil, fmt.Errorf("parse config response: %w", err)
	}
	return &GetServerConfigOutput{Config: &ServerConfig{
		DNS:             resp.Config.DNS,
		WebserverDomain: resp.Config.Webserver.Domain,
		Theme:           resp.Config.Webserver.Interface.Theme,
		Boxed:           resp.Config.Webserver.Interface.Boxed,
		PrivacyLevel:    resp.Config.Misc.PrivacyLevel,
		MaxDBDays:       resp.Config.Database.MaxDBDays,
	}}, nil
}

func (c *Client) UpdateServerConfig(in *UpdateServerConfigInput) error {
	payload := map[string]any{
		"config": map[string]any{
			"dns": in.Config.DNS,
			"webserver": map[string]any{
				"domain": in.Config.WebserverDomain,
				"interface": map[string]any{
					"theme": in.Config.Theme,
					"boxed": in.Config.Boxed,
				},
			},
			"misc": map[string]any{
				"privacylevel": in.Config.PrivacyLevel,
			},
			"database": map[string]any{
				"maxDBdays": in.Config.MaxDBDays,
			},
		},
	}
	result, err := c.Request(&RequestInput{Ctx: in.Ctx, Method: http.MethodPatch, Path: "/api/config", Body: payload})
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusOK && result.StatusCode != http.StatusNoContent {
		return fmt.Errorf("PATCH /api/config: status %d: %s", result.StatusCode, string(result.Body))
	}
	return nil
}

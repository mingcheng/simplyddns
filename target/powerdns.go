package target

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	ddns "github.com/mingcheng/simplyddns"
)

type powerDNSRecord struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`
}

type powerDNSComment struct {
	Content    string `json:"content"`
	Account    string `json:"account"`
	ModifiedAt int64  `json:"modified_at"`
}

type powerDNSRecordSet struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	TTL        int64             `json:"ttl"`
	ChangeType string            `json:"changetype,omitempty"`
	Records    []powerDNSRecord  `json:"records"`
	Comments   []powerDNSComment `json:"comments,omitempty"`
}

type powerDNSZone struct {
	Name   string              `json:"name,omitempty"`
	RRSets []powerDNSRecordSet `json:"rrsets"`
}

func init() {
	registerTarget("powerdns", updatePowerDNS)
}

func updatePowerDNS(ctx context.Context, addr *net.IP, config *ddns.TargetConfig) error {
	kind, err := recordType(addr)
	if err != nil {
		return err
	}
	if config == nil || config.Token == "" || config.Zone == "" || config.Endpoint == "" {
		return fmt.Errorf("powerdns requires token, zone and endpoint")
	}
	endpoint, err := url.Parse(config.Endpoint)
	if err != nil || (endpoint.Scheme != "https" && endpoint.Scheme != "http") ||
		endpoint.Hostname() == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return fmt.Errorf("powerdns endpoint must be an absolute HTTP(S) server API URL without credentials, query or fragment")
	}
	if _, err := dnsNames(config, ""); err != nil {
		return err
	}
	client, err := targetHTTPClient(config)
	if err != nil {
		return err
	}
	if config.Proxy != "" {
		defer client.CloseIdleConnections()
	}
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	zoneURL := strings.TrimSuffix(config.Endpoint, "/") + "/zones/" + url.PathEscape(config.Zone)
	headers := map[string]string{"X-API-Key": config.Token}
	var zone powerDNSZone
	if err := requestJSON(ctx, client, http.MethodGet, zoneURL, headers, nil, &zone, http.StatusOK); err != nil {
		return fmt.Errorf("powerdns: %w", err)
	}
	if zone.Name == "" || zone.RRSets == nil {
		return fmt.Errorf("powerdns returned an incomplete zone")
	}
	for _, record := range zone.RRSets {
		if record.Name == "" || record.Type == "" {
			return fmt.Errorf("powerdns returned an incomplete RRset")
		}
	}
	names, err := dnsNames(config, zone.Name)
	if err != nil {
		return err
	}
	var updates []powerDNSRecordSet
	for _, name := range names {
		update := powerDNSRecordSet{
			Name: name.domain + ".", Type: kind, TTL: defaultRecordTTL, ChangeType: "REPLACE",
			Records: []powerDNSRecord{{Content: addr.String()}},
		}
		found, unchanged := false, false
		for _, existing := range zone.RRSets {
			if !strings.EqualFold(strings.TrimSuffix(existing.Name, "."), name.domain) || existing.Type != kind {
				continue
			}
			if found || existing.TTL <= 0 || len(existing.Records) == 0 {
				return fmt.Errorf("powerdns %s: invalid or duplicate RRset", name.domain)
			}
			found = true
			disabled := existing.Records[0].Disabled
			for _, record := range existing.Records {
				if record.Disabled != disabled {
					return fmt.Errorf("powerdns %s: cannot merge records with mixed disabled states", name.domain)
				}
			}
			unchanged = len(existing.Records) == 1 && net.ParseIP(existing.Records[0].Content).Equal(*addr)
			update.TTL, update.Comments = existing.TTL, existing.Comments
			update.Records[0].Disabled = disabled
		}
		if !unchanged {
			updates = append(updates, update)
		}
	}
	if len(updates) == 0 {
		return nil
	}
	if err := requestJSON(ctx, client, http.MethodPatch, zoneURL, headers, powerDNSZone{RRSets: updates}, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("powerdns: %w", err)
	}
	return nil
}

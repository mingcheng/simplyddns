package target

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	ddns "github.com/mingcheng/simplyddns"
)

type goDaddyRecord struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	Data string `json:"data"`
	TTL  int64  `json:"ttl"`
}

func init() {
	registerTarget("godaddy", updateGoDaddy)
}

func updateGoDaddy(ctx context.Context, addr *net.IP, config *ddns.TargetConfig) error {
	kind, err := recordType(addr)
	if err != nil {
		return err
	}
	client, err := newHTTPClient(config)
	if err != nil {
		return err
	}
	if config.Proxy != "" {
		defer client.CloseIdleConnections()
	}
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	names, err := dnsNames(config, config.Zone)
	if err != nil {
		return err
	}
	headers := map[string]string{"Authorization": "sso-key " + config.Key + ":" + config.Token}
	for _, name := range names {
		endpoint := "https://api.godaddy.com/v1/domains/" + url.PathEscape(name.zone) +
			"/records/" + kind + "/" + url.PathEscape(name.name)
		records, err := goDaddyRecords(ctx, client, endpoint, headers)
		if err != nil {
			return fmt.Errorf("godaddy %s: %w", name.domain, err)
		}
		for _, record := range records {
			if record.Type != kind || !strings.EqualFold(record.Name, name.name) || record.TTL <= 0 {
				return fmt.Errorf("godaddy %s: record has an unexpected name, type or TTL", name.domain)
			}
		}
		if len(records) == 1 && net.ParseIP(records[0].Data).Equal(*addr) {
			continue
		}
		ttl := int64(defaultRecordTTL)
		if len(records) != 0 {
			ttl = records[0].TTL
		}
		body := []goDaddyRecord{{Data: addr.String(), TTL: ttl}}
		if err := requestJSON(ctx, client, http.MethodPut, endpoint, headers, body, nil, http.StatusOK); err != nil {
			return fmt.Errorf("godaddy %s: %w", name.domain, err)
		}
	}
	return nil
}

func goDaddyRecords(ctx context.Context, client *http.Client, endpoint string, headers map[string]string) ([]goDaddyRecord, error) {
	var records []goDaddyRecord
	previousPage := ""
	for offset := 0; ; {
		query := url.Values{"limit": {"100"}, "offset": {strconv.Itoa(offset)}}
		var page []goDaddyRecord
		if err := requestJSON(ctx, client, http.MethodGet, endpoint+"?"+query.Encode(), headers, nil, &page, http.StatusOK); err != nil {
			return nil, err
		}
		if page == nil {
			return nil, fmt.Errorf("godaddy returned a null record list")
		}
		signature := fmt.Sprint(page)
		if offset != 0 && signature == previousPage && len(page) != 0 {
			return nil, fmt.Errorf("godaddy returned a repeated record page")
		}
		for _, record := range page {
			if record.Type == "" || record.Name == "" || record.Data == "" {
				return nil, fmt.Errorf("godaddy returned an incomplete record")
			}
		}
		records = append(records, page...)
		if len(page) < 100 {
			return records, nil
		}
		previousPage = signature
		offset += len(page)
	}
}

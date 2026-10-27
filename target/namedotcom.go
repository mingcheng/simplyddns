/**
 * File: namedotcom.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, December 26th 2020, 6:33:41 pm
 * Last Modified: Sunday, December 27th 2020, 8:39:19 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package target

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	ddns "github.com/mingcheng/simplyddns"
	namedotcom "github.com/namedotcom/go"
)

type NameCom struct {
	Key     string
	Token   string
	Records map[string]*namedotcom.Record
	client  *http.Client
	ctx     context.Context
}

func newNameCom(ctx context.Context, config *ddns.TargetConfig) (*NameCom, error) {
	client, err := newHTTPClient(config)
	if err != nil {
		return nil, err
	}
	return &NameCom{
		Key: config.Key, Token: config.Token, Records: make(map[string]*namedotcom.Record),
		client: client, ctx: ctx,
	}, nil
}

// The legacy SDK drops pagination query parameters and does not close bodies.
// Keep its public record types, but own the HTTP request lifecycle here.
func (c *NameCom) request(method, path string, body, result any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(c.ctx, method, "https://api.name.com/v4"+path, reader)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.Key, c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", ddns.DefaultUserAgent)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var failure namedotcom.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&failure); err != nil {
			return fmt.Errorf("name.com returned %s: %w", resp.Status, err)
		}
		return fmt.Errorf("name.com returned %s: %s", resp.Status, failure.Error())
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

// Ping function is used to test if the namedotcom instance is available
func (c *NameCom) Ping() error {
	var response namedotcom.HelloResponse
	return c.request(http.MethodGet, "/hello", nil, &response)
}

func nameComRecordKey(domain, kind string) string {
	if kind == "A" {
		return domain
	}
	return kind + " " + domain
}

// FindRecord function is used to find the record from namedotcom instance
func (c *NameCom) FindRecord(domain string) (*namedotcom.Record, error) {
	return c.findRecord(domain, "A")
}

func (c *NameCom) findRecord(domain, kind string) (*namedotcom.Record, error) {
	tld, err := ddns.ParseDomain(domain)
	if err != nil {
		return nil, err
	}

	domain = tld.Hostname()
	key := nameComRecordKey(domain, kind)
	if record := c.Records[key]; record != nil {
		return record, nil
	}
	for page := int32(1); ; {
		query := url.Values{"page": {strconv.Itoa(int(page))}, "perPage": {"100"}}
		path := "/domains/" + url.PathEscape(tld.Domain+"."+tld.TLD) + "/records?" + query.Encode()
		var result namedotcom.ListRecordsResponse
		if err := c.request(http.MethodGet, path, nil, &result); err != nil {
			return nil, err
		}
		for _, record := range result.Records {
			if record == nil {
				return nil, fmt.Errorf("name.com returned a null DNS record")
			}
			if record.Type != "A" && record.Type != "AAAA" {
				continue
			}
			name := record.DomainName
			if record.Host != "" && record.Host != "@" {
				name = record.Host + "." + name
			}
			name = strings.ToLower(strings.TrimSuffix(name, "."))
			c.Records[nameComRecordKey(name, record.Type)] = record
		}
		if record := c.Records[key]; record != nil {
			return record, nil
		}
		if result.NextPage == 0 {
			return nil, nil
		}
		if result.NextPage <= page {
			return nil, fmt.Errorf("name.com returned a non-advancing page number")
		}
		page = result.NextPage
	}
}

// UpdateRecord function is used to update the record from namedotcom instance
func (c *NameCom) UpdateRecord(record *namedotcom.Record) (*namedotcom.Record, error) {
	if record == nil || record.ID <= 0 {
		return nil, fmt.Errorf("name.com record ID is required")
	}
	path := fmt.Sprintf("/domains/%s/records/%d", url.PathEscape(record.DomainName), record.ID)
	var result namedotcom.Record
	if err := c.request(http.MethodPut, path, record, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateRecord function is used to create the record from namedotcom instance
func (c *NameCom) CreateRecord(domain string, addr *net.IP) (*namedotcom.Record, error) {
	kind, err := recordType(addr)
	if err != nil {
		return nil, err
	}
	tld, err := ddns.ParseDomain(domain)
	if err != nil {
		return nil, err
	}

	record := namedotcom.Record{
		Type:       kind,
		Answer:     addr.String(),
		DomainName: fmt.Sprintf("%s.%s", tld.Domain, tld.TLD),
		Host:       tld.Subdomain,
		TTL:        600,
	}

	var result namedotcom.Record
	path := "/domains/" + url.PathEscape(record.DomainName) + "/records"
	if err := c.request(http.MethodPost, path, &record, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func init() {
	target := func(ctx context.Context, addr *net.IP, config *ddns.TargetConfig) error {
		kind, err := recordType(addr)
		if err != nil {
			return err
		}
		namecom, err := newNameCom(ctx, config)
		if err != nil {
			return err
		}
		if config.Proxy != "" {
			defer namecom.client.CloseIdleConnections()
		}
		if len(config.Domains) == 0 {
			return fmt.Errorf("name.com target requires at least one domain")
		}
		for _, domain := range config.Domains {
			record, err := namecom.findRecord(domain, kind)
			if err != nil {
				return fmt.Errorf("name.com %s: %w", domain, err)
			}
			if record == nil {
				record, err = namecom.CreateRecord(domain, addr)
			} else if net.ParseIP(record.Answer).Equal(*addr) {
				continue
			} else {
				update := *record
				update.Answer = addr.String()
				record, err = namecom.UpdateRecord(&update)
			}
			if err != nil {
				return fmt.Errorf("name.com %s: %w", domain, err)
			}
			namecom.Records[nameComRecordKey(domain, kind)] = record
		}

		return nil
	}

	registerTarget("namedotcom", target)
}

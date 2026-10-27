/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: cloudflare.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: 2025-11-27 17:04:40
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:14:07
 */

package target

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	ddns "github.com/mingcheng/simplyddns"
)

func init() {
	registerTarget("cloudflare",
		func(ctx context.Context, ip *net.IP, config *ddns.TargetConfig) error {
			if _, err := recordType(ip); err != nil {
				return err
			}
			client, err := NewCloudflareDNSClient(config)
			if err != nil {
				return err
			}
			if config.Proxy != "" {
				defer client.httpClient.CloseIdleConnections()
			}
			if len(config.Domains) == 0 {
				return fmt.Errorf("cloudflare target requires at least one domain")
			}
			for _, domain := range config.Domains {
				if err := client.CreateOrUpdateRecord(ctx, domain, ip); err != nil {
					return fmt.Errorf("cloudflare %s: %w", domain, err)
				}
			}

			return nil
		})
}

type CloudflareClient struct {
	client         *cloudflare.API
	zoneIdentifier *cloudflare.ResourceContainer
	httpClient     *http.Client
}

func (c CloudflareClient) CreateOrUpdateRecord(ctx context.Context, domain string, ip *net.IP) error {
	kind, err := recordType(ip)
	if err != nil {
		return err
	}
	parsed, err := ddns.ParseDomain(domain)
	if err != nil {
		return err
	}
	domain = parsed.Hostname()
	records, _, err := c.client.ListDNSRecords(ctx, c.zoneIdentifier, cloudflare.ListDNSRecordsParams{
		Type: kind,
		Name: domain,
	})

	if err != nil {
		return err
	}

	found := false
	for _, record := range records {
		if strings.EqualFold(strings.TrimSuffix(record.Name, "."), domain) && record.Type == kind {
			found = true
			if net.ParseIP(record.Content).Equal(*ip) {
				continue
			}
			if err := c.updateRecord(ctx, record, ip); err != nil {
				return err
			}
		}
	}
	if found {
		return nil
	}
	return c.CreateRecord(ctx, domain, ip)
}

func (c CloudflareClient) UpdateRecord(ctx context.Context, recordID string, ip *net.IP) error {
	if _, err := recordType(ip); err != nil {
		return err
	}
	record, err := c.client.GetDNSRecord(ctx, c.zoneIdentifier, recordID)
	if err != nil {
		return err
	}
	return c.updateRecord(ctx, record, ip)
}

func (c CloudflareClient) updateRecord(ctx context.Context, record cloudflare.DNSRecord, ip *net.IP) error {
	kind, err := recordType(ip)
	if err != nil {
		return err
	}
	if record.Type != kind {
		return fmt.Errorf("cannot update a %s record with a %s address", record.Type, kind)
	}
	log.Infof("update the cloudflare record %s with ip %s", record.ID, ip)
	_, err = c.client.UpdateDNSRecord(ctx, c.zoneIdentifier, cloudflare.UpdateDNSRecordParams{
		ID:       record.ID,
		Content:  ip.String(),
		TTL:      record.TTL,
		Proxied:  record.Proxied,
		Priority: record.Priority,
		Comment:  &record.Comment,
		Tags:     record.Tags,
		Settings: record.Settings,
	})
	return err
}

func (c CloudflareClient) DeleteRecord(ctx context.Context, recordID string) error {
	log.Warnf("remove the old record %s", recordID)
	return c.client.DeleteDNSRecord(ctx, c.zoneIdentifier, recordID)
}

func (c CloudflareClient) CreateRecord(ctx context.Context, domain string, ip *net.IP) error {
	kind, err := recordType(ip)
	if err != nil {
		return err
	}
	parsed, err := ddns.ParseDomain(domain)
	if err != nil {
		return err
	}
	log.Infof("create a new record for domain %s with ip %s", domain, ip)
	_, err = c.client.CreateDNSRecord(ctx, c.zoneIdentifier, cloudflare.CreateDNSRecordParams{
		Type:    kind,
		Name:    parsed.Hostname(),
		Content: ip.String(),
		TTL:     60,
	})

	return err
}

func NewCloudflareDNSClient(config *ddns.TargetConfig) (*CloudflareClient, error) {
	httpClient, err := newHTTPClient(config)
	if err != nil {
		return nil, err
	}
	client, err := cloudflare.NewWithAPIToken(config.Token, cloudflare.HTTPClient(httpClient),
		cloudflare.UserAgent(ddns.DefaultUserAgent))
	if err != nil {
		return nil, err
	}

	return &CloudflareClient{
		client:         client,
		zoneIdentifier: cloudflare.ZoneIdentifier(config.Key),
		httpClient:     httpClient,
	}, nil
}

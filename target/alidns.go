/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: alidns.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: Thursday, December 24th 2020, 10:58:18 am
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:24:20
 */

package target

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	alidns "github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/mingcheng/simplyddns"
)

// AliDNS is the target for aliyun dns
type AliDNS struct {
	Client  *alidns.Client
	Config  *simplyddns.TargetConfig
	records map[string]*alidns.Record
}

var errRecordNotFound = errors.New("DNS record not found")

// NewAliDNS for creating a new instance of AliDNS
func NewAliDNS(config *simplyddns.TargetConfig) (*AliDNS, error) {
	return newAliDNS(context.Background(), config)
}

func newAliDNS(ctx context.Context, config *simplyddns.TargetConfig) (*AliDNS, error) {
	httpClient, err := newHTTPClient(config)
	if err != nil {
		return nil, err
	}
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", config.Key, config.Token)
	if err != nil {
		return nil, err
	}
	client.SetTransport(contextTransport{ctx: ctx, base: httpClient.Transport})
	client.SetReadTimeout(simplyddns.DefaultTimeout)
	client.SetConnectTimeout(simplyddns.DefaultTimeout)
	return &AliDNS{
		Client:  client,
		Config:  config,
		records: map[string]*alidns.Record{},
	}, nil
}

// All to get all domains
func (a *AliDNS) All() ([]alidns.DomainInDescribeDomains, error) {
	var all []alidns.DomainInDescribeDomains
	for page := 1; ; page++ {
		req := alidns.CreateDescribeDomainsRequest()
		req.Scheme = "https"
		req.PageNumber = requests.NewInteger(page)
		req.PageSize = requests.NewInteger(100)
		resp, err := a.Client.DescribeDomains(req)
		if err != nil {
			return nil, err
		}
		all = append(all, resp.Domains.Domain...)
		if int64(len(all)) >= resp.TotalCount {
			return all, nil
		}
		if len(resp.Domains.Domain) == 0 {
			return nil, fmt.Errorf("alidns returned an empty domain page before the end of the list")
		}
	}
}

// GetRecord to get a single DNS record
func (a *AliDNS) GetRecord(domain string) (*alidns.Record, error) {
	return a.getRecord(domain, "A")
}

func (a *AliDNS) getRecord(domain, kind string) (*alidns.Record, error) {
	tld, err := simplyddns.ParseDomain(domain)
	if err != nil {
		return nil, err
	}

	key := kind + " " + tld.Hostname()
	if a.records[key] != nil {
		return a.records[key], nil
	}

	req := alidns.CreateDescribeSubDomainRecordsRequest()
	req.Scheme = "https"
	req.SubDomain = tld.Hostname()
	req.DomainName = tld.Domain + "." + tld.TLD
	req.Type = kind

	resp, err := a.Client.DescribeSubDomainRecords(req)
	if err != nil {
		return nil, err
	}

	if len(resp.DomainRecords.Record) == 0 {
		if resp.TotalCount > 0 {
			return nil, fmt.Errorf("alidns returned an empty record list with a nonzero count")
		}
		return nil, fmt.Errorf("%s: %w", domain, errRecordNotFound)
	}
	for _, record := range resp.DomainRecords.Record {
		if record.Type == kind {
			a.records[key] = &record
			return &record, nil
		}
	}
	return nil, fmt.Errorf("alidns returned no %s records for %s despite a nonempty result", kind, domain)
}

// UpdateRecord updates an existing DNS record.
func (a *AliDNS) UpdateRecord(domain string, ip *net.IP) (*alidns.UpdateDomainRecordResponse, error) {
	kind, err := recordType(ip)
	if err != nil {
		return nil, err
	}
	tld, err := simplyddns.ParseDomain(domain)
	if err != nil {
		return nil, err
	}

	record, err := a.getRecord(domain, kind)
	if err != nil {
		return nil, err
	}

	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"

	request.Value = ip.String()
	request.Type = kind
	request.TTL = requests.NewInteger64(record.TTL)
	request.Line = record.Line

	if tld.Subdomain == "" {
		request.RR = "@"
	} else {
		request.RR = tld.Subdomain
	}
	request.RecordId = record.RecordId

	resp, err := a.Client.UpdateDomainRecord(request)
	if err == nil {
		record.Value = ip.String()
	}
	return resp, err
}

// CreateRecord to create a new DNS record
func (a *AliDNS) CreateRecord(domain string, ip *net.IP) (*alidns.AddDomainRecordResponse, error) {
	kind, err := recordType(ip)
	if err != nil {
		return nil, err
	}
	tld, err := simplyddns.ParseDomain(domain)
	if err != nil {
		return nil, err
	}

	request := alidns.CreateAddDomainRecordRequest()
	request.Scheme = "https"
	request.Type = kind
	request.Value = ip.String()
	request.DomainName = fmt.Sprintf("%s.%s", tld.Domain, tld.TLD)

	if tld.Subdomain == "" {
		request.RR = "@"
	} else {
		request.RR = tld.Subdomain
	}

	return a.Client.AddDomainRecord(request)
}

// init to register the function to dispatcher
func init() {
	registerTarget("alidns",
		func(ctx context.Context, ip *net.IP, config *simplyddns.TargetConfig) error {
			kind, err := recordType(ip)
			if err != nil {
				return err
			}
			client, err := newAliDNS(ctx, config)
			if err != nil {
				return err
			}
			if len(config.Domains) == 0 {
				return fmt.Errorf("alidns target requires at least one domain")
			}
			for _, domain := range config.Domains {
				record, err := client.getRecord(domain, kind)
				switch {
				case errors.Is(err, errRecordNotFound):
					_, err = client.CreateRecord(domain, ip)
				case err != nil:
					return fmt.Errorf("alidns %s: %w", domain, err)
				case net.ParseIP(record.Value).Equal(*ip):
					continue
				default:
					_, err = client.UpdateRecord(domain, ip)
				}
				if err != nil {
					return fmt.Errorf("alidns %s: %w", domain, err)
				}
			}

			return nil
		})
}

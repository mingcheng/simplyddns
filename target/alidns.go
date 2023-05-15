/**
 * File: alidns.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Thursday, December 24th 2020, 10:58:18 am
 * Last Modified: Monday, December 28th 2020, 2:49:23 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package target

import (
	"context"
	"fmt"
	"net"

	alidns "github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/mingcheng/simplyddns"
)

// AliDNS is the target for aliyun dns
type AliDNS struct {
	Client  *alidns.Client
	Config  *simplyddns.TargetConfig
	domains map[string]*alidns.DomainInDescribeDomains
	records map[string]*alidns.Record
}

// NewAliDNS for creating a new instance of AliDNS
func NewAliDNS(config *simplyddns.TargetConfig) (*AliDNS, error) {
	var (
		err        error
		client     *alidns.Client
		allDomains []alidns.DomainInDescribeDomains
	)
	client, err = alidns.NewClientWithAccessKey("cn-hangzhou", config.Key, config.Token)
	if err != nil {
		return nil, err
	}

	instance := &AliDNS{
		Client:  client,
		Config:  config,
		domains: map[string]*alidns.DomainInDescribeDomains{},
		records: map[string]*alidns.Record{},
	}

	if allDomains, err = instance.All(); err != nil {
		return nil, err
	}

	// cache the domain
	if len(allDomains) > 0 {
		for _, v := range allDomains {
			instance.domains[v.DomainName] = &v
		}
	}

	return instance, nil
}

// All to get all domains
func (a *AliDNS) All() ([]alidns.DomainInDescribeDomains, error) {
	req := alidns.CreateDescribeDomainsRequest()
	domains, err := a.Client.DescribeDomains(req)
	if err != nil {
		return nil, err
	}

	if domains.TotalCount == int64(len(domains.Domains.Domain)) {
		return domains.Domains.Domain, nil
	}

	return []alidns.DomainInDescribeDomains{}, nil
}

// GetRecord to get a single DNS record
func (a *AliDNS) GetRecord(domain string) (*alidns.Record, error) {
	tld, err := simplyddns.ParseDomain(domain)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%s.%s", tld.Domain, tld.TLD)
	if a.domains[key] == nil {
		return nil, fmt.Errorf("domian %s is not found", key)
	}

	if a.records[domain] != nil {
		return a.records[domain], nil
	}

	req := alidns.CreateDescribeSubDomainRecordsRequest()
	req.SubDomain = domain
	req.DomainName = tld.Subdomain
	req.Type = "A"
	// req.Type= "AAAA"

	resp, err := a.Client.DescribeSubDomainRecords(req)
	if err != nil {
		return nil, err
	}

	if resp.TotalCount <= 0 {
		return nil, fmt.Errorf("%v record is not found", domain)
	}

	// only get first record
	record := resp.DomainRecords.Record[0]
	a.records[domain] = &record
	return &record, nil
}

// UpdateRecord for update a existsed record
func (a *AliDNS) UpdateRecord(domain string, ip *net.IP) (*alidns.UpdateDomainRecordResponse, error) {
	tld, err := simplyddns.ParseDomain(domain)
	if err != nil {
		return nil, err
	}

	record, err := a.GetRecord(domain)
	if err != nil {
		return nil, err
	}

	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"

	request.Value = ip.String()
	request.Type = "A"

	if tld.Subdomain == "" {
		request.RR = "@"
	} else {
		request.RR = tld.Subdomain
	}
	request.RecordId = record.RecordId

	return a.Client.UpdateDomainRecord(request)
}

// CreateRecord to create a new DNS record
func (a *AliDNS) CreateRecord(domain string, ip *net.IP) (*alidns.AddDomainRecordResponse, error) {
	tld, err := simplyddns.ParseDomain(domain)
	if err != nil {
		return nil, err
	}

	request := alidns.CreateAddDomainRecordRequest()
	request.Scheme = "https"
	request.Type = "A"
	request.Value = ip.String()
	request.DomainName = fmt.Sprintf("%s.%s", tld.Domain, tld.TLD)

	if tld.Subdomain == "" {
		request.RR = "@"
	} else {
		request.RR = tld.Subdomain
	}

	return a.Client.AddDomainRecord(request)
}

// init to register the function to dispather
func init() {
	_ = simplyddns.RegisterTargetFunc("alidns", func(ctx context.Context, ip *net.IP, config *simplyddns.TargetConfig) error {
		var (
			err    error
			client *AliDNS
		)

		if client, err = NewAliDNS(config); err != nil {
			return err
		}

		for _, domain := range config.Domains {
			if record, _ := client.GetRecord(domain); record != nil {
				_, err = client.UpdateRecord(domain, ip)
			} else {
				_, err = client.CreateRecord(domain, ip)
			}
		}

		return err
	})
}

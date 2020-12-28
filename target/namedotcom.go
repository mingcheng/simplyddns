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
	"context"
	"fmt"
	"net"

	ddns "github.com/mingcheng/simplyddns"
	namedotcom "github.com/namedotcom/go"
)

type NameCom struct {
	Key      string
	Token    string
	Records  map[string]*namedotcom.Record
	instance *namedotcom.NameCom
}

func (c *NameCom) Ping() error {
	if _, err := c.instance.HelloFunc(); err != nil {
		return err
	}

	return nil
}

func (c *NameCom) FindRecord(domain string) (*namedotcom.Record, error) {
	// if already cached
	if c.Records[domain] != nil {
		return c.Records[domain], nil
	}

	// request from dot name service
	tld, err := ddns.ParseDomain(domain)
	if err != nil {
		return nil, err
	}

	// get all records
	all, err := c.instance.ListRecords(&namedotcom.ListRecordsRequest{
		DomainName: fmt.Sprintf("%s.%s", tld.Domain, tld.TLD),
		PerPage:    100,
	})

	if err != nil {
		return nil, err
	}

	var find *namedotcom.Record
	for _, v := range all.Records {
		indexName := v.DomainName
		if v.Host != "" {
			indexName = fmt.Sprintf("%s.%s", v.Host, v.DomainName)
		}

		// cache the record into map, record type must A
		if v.Type == "A" {
			(c.Records)[indexName] = v

			// mark if match the record
			if indexName == domain {
				find = v
			}
		}
	}

	return find, nil
}

func (c *NameCom) UpdateRecord(record *namedotcom.Record) (*namedotcom.Record, error) {
	return c.instance.UpdateRecord(record)
}

func (c *NameCom) CreateRecord(domain string, addr *net.IP) (*namedotcom.Record, error) {
	tld, err := ddns.ParseDomain(domain)
	if err != nil {
		return nil, err
	}

	record := namedotcom.Record{
		Type:       "A",
		Answer:     addr.String(),
		DomainName: fmt.Sprintf("%s.%s", tld.Domain, tld.TLD),
		Host:       tld.Subdomain,
		TTL:        600,
	}

	return c.instance.CreateRecord(&record)
}

func init() {
	// cache domain records for every namedotcom instance
	domainRecords := map[string]*namedotcom.Record{}

	// NewNameCom for instance a new namedotcom client
	NewNameCom := func(key, token string, config *ddns.TargetConfig) (*NameCom, error) {
		instance := namedotcom.New(key, token)
		if config.Proxy != "" {
			if client, err := ddns.ProxyHttpClient(config.Proxy); client != nil {
				instance.Client = client
			} else {
				return nil, err
			}
		}

		return &NameCom{
			Key:      key,
			Token:    token,
			Records:  domainRecords,
			instance: instance,
		}, nil
	}

	target := func(ctx context.Context, addr *net.IP, config *ddns.TargetConfig) error {
		namecom, err := NewNameCom(config.Key, config.Token, config)
		if err != nil {
			return err
		}

		if err := namecom.Ping(); err != nil {
			return err
		}

		for _, domain := range config.Domains {
			findRecord, _ := namecom.FindRecord(domain)

			if findRecord != nil {
				if findRecord.Answer == addr.String() {
					return fmt.Errorf("record %s already set the address %s", domain, findRecord.Answer)
				}

				findRecord.Answer = addr.String()
				if _, err = namecom.UpdateRecord(findRecord); err != nil {
					return err
				}
			} else {
				if _, err = namecom.CreateRecord(domain, addr); err != nil {
					return err
				}
			}
		}

		return nil
	}

	_ = ddns.RegisterTargetFunc("namedotcom", target)
}

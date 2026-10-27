package target

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	httpconfig "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	dns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"
	ddns "github.com/mingcheng/simplyddns"
)

func init() {
	registerTarget("huaweicloud", updateHuaweiCloud)
}

func updateHuaweiCloud(ctx context.Context, addr *net.IP, config *ddns.TargetConfig) error {
	kind, err := recordType(addr)
	if err != nil {
		return err
	}
	httpClient, err := newHTTPClient(config)
	if err != nil {
		return err
	}
	if config.Proxy != "" {
		defer httpClient.CloseIdleConnections()
	}
	if config.Zone == "" || config.Region == "" || config.ProjectID == "" {
		return fmt.Errorf("huaweicloud requires zone, region and project_id")
	}
	if _, err := dnsNames(config, ""); err != nil {
		return err
	}
	serviceRegion, err := region.SafeValueOf(config.Region)
	if err != nil {
		return err
	}
	credential, err := basic.NewCredentialsBuilder().WithAk(config.Key).WithSk(config.Token).
		WithProjectId(config.ProjectID).WithSecurityToken(config.SessionToken).SafeBuild()
	if err != nil {
		return err
	}
	clientConfig := httpconfig.DefaultHttpConfig().WithTimeout(ddns.DefaultTimeout).
		WithHttpRoundTripper(contextTransport{ctx: ctx, base: httpClient.Transport}).
		WithUserAgent(ddns.DefaultUserAgent)
	base, err := dns.DnsClientBuilder().WithRegion(serviceRegion).WithCredential(credential).WithHttpConfig(clientConfig).SafeBuild()
	if err != nil {
		return err
	}
	client := dns.NewDnsClient(base)
	zone, err := client.ShowPublicZone(&model.ShowPublicZoneRequest{ZoneId: config.Zone})
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("huaweicloud: %w", err)
	}
	if zone == nil || zone.Name == nil || *zone.Name == "" {
		return fmt.Errorf("huaweicloud returned a zone without a name")
	}
	names, err := dnsNames(config, *zone.Name)
	if err != nil {
		return err
	}
	for _, name := range names {
		if err := syncHuaweiRecord(client, config.Zone, name.domain, kind, addr); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("huaweicloud %s: %w", name.domain, err)
		}
	}
	return nil
}

func syncHuaweiRecord(client *dns.DnsClient, zone, domain, kind string, addr *net.IP) error {
	records, err := huaweiRecords(client, zone, domain, kind)
	if err != nil {
		return err
	}
	values := []string{addr.String()}
	if len(records) == 0 {
		ttl := int32(defaultRecordTTL)
		response, err := client.CreateRecordSet(&model.CreateRecordSetRequest{
			ZoneId: zone,
			Body: &model.CreateRecordSetRequestBody{
				Name: domain + ".", Type: kind, Ttl: &ttl, Records: values,
			},
		})
		if err != nil {
			return err
		}
		if response == nil || response.Id == nil || *response.Id == "" {
			return fmt.Errorf("create response is missing the record ID")
		}
		if response.Status != nil && *response.Status == "ERROR" {
			return fmt.Errorf("created record is in ERROR state")
		}
		return nil
	}
	for _, record := range records {
		if len(*record.Records) == 1 && net.ParseIP((*record.Records)[0]).Equal(*addr) {
			continue
		}
		response, err := client.UpdateRecordSet(&model.UpdateRecordSetRequest{
			ZoneId: zone, RecordsetId: *record.Id,
			Body: &model.UpdateRecordSetReq{
				Records: &values, Ttl: record.Ttl, Description: record.Description,
			},
		})
		if err != nil {
			return err
		}
		if response == nil || response.Id == nil || *response.Id != *record.Id {
			return fmt.Errorf("update response has an invalid record ID")
		}
		if response.Status != nil && *response.Status == "ERROR" {
			return fmt.Errorf("updated record is in ERROR state")
		}
	}
	return nil
}

func huaweiRecords(client *dns.DnsClient, zone, domain, kind string) ([]model.ListRecordSets, error) {
	limit, searchMode, name := int32(100), "equal", domain+"."
	request := &model.ListRecordSetsByZoneRequest{
		ZoneId: zone, Name: &name, Type: &kind, SearchMode: &searchMode, Limit: &limit,
	}
	var records []model.ListRecordSets
	seen := make(map[string]bool)
	var count int64
	for {
		result, err := client.ListRecordSetsByZone(request)
		if err != nil {
			return nil, err
		}
		if result == nil || result.Recordsets == nil || result.Metadata == nil || result.Metadata.TotalCount == nil ||
			*result.Metadata.TotalCount < 0 {
			return nil, fmt.Errorf("record list is missing records or pagination metadata")
		}
		page := *result.Recordsets
		for _, record := range page {
			if record.Id == nil || *record.Id == "" || record.Name == nil || record.Type == nil {
				return nil, fmt.Errorf("record list contains an incomplete record")
			}
			if seen[*record.Id] {
				return nil, fmt.Errorf("record list contains a repeated record ID")
			}
			seen[*record.Id] = true
			if !strings.EqualFold(strings.TrimSuffix(*record.Name, "."), domain) || *record.Type != kind {
				continue
			}
			if record.Ttl == nil || *record.Ttl <= 0 || record.Records == nil || len(*record.Records) == 0 {
				return nil, fmt.Errorf("record %s is missing TTL or address data", *record.Id)
			}
			records = append(records, record)
		}
		count += int64(len(page))
		if count >= int64(*result.Metadata.TotalCount) {
			return records, nil
		}
		if len(page) == 0 {
			return nil, fmt.Errorf("record list ended before its advertised count")
		}
		request.Marker = page[len(page)-1].Id
	}
}

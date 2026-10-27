package target

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	ddns "github.com/mingcheng/simplyddns"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

func init() {
	registerTarget("dnspod", updateDNSPod)
}

func updateDNSPod(ctx context.Context, addr *net.IP, config *ddns.TargetConfig) error {
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
	names, err := dnsNames(config, config.Zone)
	if err != nil {
		return err
	}
	clientProfile := profile.NewClientProfile()
	clientProfile.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	clientProfile.HttpProfile.ReqTimeout = int(ddns.DefaultTimeout / time.Second)
	credential := common.NewTokenCredential(config.Key, config.Token, config.SessionToken)
	client, err := dnspod.NewClient(credential, "", clientProfile)
	if err != nil {
		return err
	}
	client.WithHttpTransport(httpClient.Transport)
	for _, name := range names {
		if err := syncDNSPodRecord(ctx, client, name, kind, addr); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("dnspod %s: %w", name.domain, err)
		}
	}
	return nil
}

func syncDNSPodRecord(ctx context.Context, client *dnspod.Client, name dnsName, kind string, addr *net.IP) error {
	records, err := dnsPodRecords(ctx, client, name, kind)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		request := dnspod.NewCreateRecordRequest()
		request.Domain, request.SubDomain = common.StringPtr(name.zone), common.StringPtr(name.name)
		request.RecordType, request.Value = common.StringPtr(kind), common.StringPtr(addr.String())
		request.RecordLine, request.RecordLineId = common.StringPtr("\u9ed8\u8ba4"), common.StringPtr("0")
		request.TTL = common.Uint64Ptr(defaultRecordTTL)
		response, err := client.CreateRecordWithContext(ctx, request)
		if err != nil {
			return err
		}
		if response == nil || response.Response == nil || response.Response.RecordId == nil || *response.Response.RecordId == 0 {
			return fmt.Errorf("create response is missing the record ID")
		}
		return nil
	}
	for _, record := range records {
		if net.ParseIP(*record.Value).Equal(*addr) {
			continue
		}
		request := dnspod.NewModifyRecordRequest()
		request.Domain, request.SubDomain = common.StringPtr(name.zone), common.StringPtr(name.name)
		request.RecordId, request.RecordType = record.RecordId, common.StringPtr(kind)
		request.RecordLine, request.RecordLineId = record.Line, record.LineId
		request.Value, request.TTL = common.StringPtr(addr.String()), record.TTL
		request.Status, request.Weight, request.Remark = record.Status, record.Weight, record.Remark
		response, err := client.ModifyRecordWithContext(ctx, request)
		if err != nil {
			return err
		}
		if response == nil || response.Response == nil || response.Response.RecordId == nil ||
			*response.Response.RecordId != *record.RecordId {
			return fmt.Errorf("update response has an invalid record ID")
		}
	}
	return nil
}

func dnsPodRecords(ctx context.Context, client *dnspod.Client, name dnsName, kind string) ([]*dnspod.RecordListItem, error) {
	var records []*dnspod.RecordListItem
	seen := make(map[uint64]bool)
	for offset := uint64(0); ; {
		request := dnspod.NewDescribeRecordListRequest()
		request.Domain, request.SubDomain = common.StringPtr(name.zone), common.StringPtr(name.name)
		request.RecordType, request.ErrorOnEmpty = common.StringPtr(kind), common.StringPtr("no")
		request.Offset, request.Limit = common.Uint64Ptr(offset), common.Uint64Ptr(100)
		result, err := client.DescribeRecordListWithContext(ctx, request)
		if err != nil {
			return nil, err
		}
		if result == nil || result.Response == nil || result.Response.RecordCountInfo == nil ||
			result.Response.RecordCountInfo.TotalCount == nil {
			return nil, fmt.Errorf("record list is missing its count")
		}
		page := result.Response.RecordList
		for _, record := range page {
			if record == nil || record.RecordId == nil || *record.RecordId == 0 || record.Name == nil || record.Type == nil {
				return nil, fmt.Errorf("record list contains an incomplete record")
			}
			if seen[*record.RecordId] {
				return nil, fmt.Errorf("record list contains a repeated record ID")
			}
			seen[*record.RecordId] = true
			if !strings.EqualFold(*record.Name, name.name) || *record.Type != kind {
				continue
			}
			if record.Value == nil || record.Line == nil || *record.Line == "" || record.TTL == nil || *record.TTL == 0 ||
				record.Status == nil || (*record.Status != "ENABLE" && *record.Status != "DISABLE") {
				return nil, fmt.Errorf("record %d is missing required value, line, TTL or status metadata", *record.RecordId)
			}
			records = append(records, record)
		}
		offset += uint64(len(page))
		if offset >= *result.Response.RecordCountInfo.TotalCount {
			return records, nil
		}
		if len(page) == 0 {
			return nil, fmt.Errorf("record list ended before its advertised count")
		}
	}
}

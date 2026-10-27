package target

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	ddns "github.com/mingcheng/simplyddns"
)

func init() {
	registerTarget("route53", updateRoute53)
}

func updateRoute53(ctx context.Context, addr *net.IP, config *ddns.TargetConfig) error {
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
	if config.Zone == "" {
		return fmt.Errorf("route53 requires a hosted zone ID in target.zone")
	}
	if _, err := dnsNames(config, ""); err != nil {
		return err
	}
	region := config.Region
	if region == "" {
		region = "us-east-1"
	}
	client := route53.NewFromConfig(aws.Config{
		Region: region, HTTPClient: httpClient,
		Credentials: aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID: config.Key, SecretAccessKey: config.Token,
				SessionToken: config.SessionToken, Source: "simplyddns",
			}, nil
		}),
	})
	zone, err := client.GetHostedZone(ctx, &route53.GetHostedZoneInput{Id: aws.String(config.Zone)})
	if err != nil {
		return fmt.Errorf("route53: %w", err)
	}
	if zone == nil || zone.HostedZone == nil || aws.ToString(zone.HostedZone.Name) == "" {
		return fmt.Errorf("route53 returned a zone without a name")
	}
	names, err := dnsNames(config, *zone.HostedZone.Name)
	if err != nil {
		return err
	}
	for _, name := range names {
		records, err := route53Records(ctx, client, config.Zone, name.domain, types.RRType(kind))
		if err != nil {
			return fmt.Errorf("route53 %s: %w", name.domain, err)
		}
		if len(records) == 0 {
			records = append(records, types.ResourceRecordSet{
				Name: aws.String(name.domain + "."), Type: types.RRType(kind), TTL: aws.Int64(defaultRecordTTL),
			})
		}
		var changes []types.Change
		for _, record := range records {
			if record.AliasTarget != nil {
				return fmt.Errorf("route53 %s: refusing to replace an alias record with an IP address", name.domain)
			}
			if record.TTL == nil || *record.TTL <= 0 || record.TrafficPolicyInstanceId != nil {
				return fmt.Errorf("route53 %s: missing TTL or unsupported traffic-policy record", name.domain)
			}
			if len(record.ResourceRecords) == 1 && net.ParseIP(aws.ToString(record.ResourceRecords[0].Value)).Equal(*addr) {
				continue
			}
			record.ResourceRecords = []types.ResourceRecord{{Value: aws.String(addr.String())}}
			changes = append(changes, types.Change{Action: types.ChangeActionUpsert, ResourceRecordSet: &record})
		}
		if len(changes) == 0 {
			continue
		}
		response, err := client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(config.Zone), ChangeBatch: &types.ChangeBatch{Changes: changes},
		})
		if err != nil {
			return fmt.Errorf("route53 %s: %w", name.domain, err)
		}
		if response == nil || response.ChangeInfo == nil || aws.ToString(response.ChangeInfo.Id) == "" {
			return fmt.Errorf("route53 %s: change response is missing its ID", name.domain)
		}
		if response.ChangeInfo.Status != types.ChangeStatusPending && response.ChangeInfo.Status != types.ChangeStatusInsync {
			return fmt.Errorf("route53 %s: unexpected change status %q", name.domain, response.ChangeInfo.Status)
		}
	}
	return nil
}

func route53Records(ctx context.Context, client *route53.Client, zone, domain string, kind types.RRType) ([]types.ResourceRecordSet, error) {
	request := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zone), StartRecordName: aws.String(domain + "."),
		StartRecordType: kind, MaxItems: aws.Int32(100),
	}
	var records []types.ResourceRecordSet
	seen := make(map[string]bool)
	for {
		result, err := client.ListResourceRecordSets(ctx, request)
		if err != nil {
			return nil, err
		}
		if result == nil || result.MaxItems == nil {
			return nil, fmt.Errorf("record list is missing pagination metadata")
		}
		for _, record := range result.ResourceRecordSets {
			if aws.ToString(record.Name) == "" || record.Type == "" {
				return nil, fmt.Errorf("record list contains an incomplete record")
			}
			if !strings.EqualFold(strings.TrimSuffix(*record.Name, "."), domain) || record.Type != kind {
				return records, nil
			}
			records = append(records, record)
		}
		if !result.IsTruncated {
			return records, nil
		}
		if len(result.ResourceRecordSets) == 0 || aws.ToString(result.NextRecordName) == "" || result.NextRecordType == "" {
			return nil, fmt.Errorf("truncated record list has no valid continuation")
		}
		next := *result.NextRecordName + " " + string(result.NextRecordType) + " " + aws.ToString(result.NextRecordIdentifier)
		if seen[next] {
			return nil, fmt.Errorf("record list returned a repeated continuation")
		}
		seen[next] = true
		request.StartRecordName, request.StartRecordType = result.NextRecordName, result.NextRecordType
		request.StartRecordIdentifier = result.NextRecordIdentifier
	}
}

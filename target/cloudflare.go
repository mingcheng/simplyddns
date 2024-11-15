package target

import (
	"context"
	cloudflare "github.com/cloudflare/cloudflare-go"
	ddns "github.com/mingcheng/simplyddns"
	"net"
)

func init() {
	_ = ddns.RegisterTargetFunc("cloudflare",
		func(ctx context.Context, ip *net.IP, config *ddns.TargetConfig) error {
			var (
				err    error
				client *CloudflareClient
			)

			if client, err = NewCloudflareDNSClient(config); err != nil {
				return err
			}

			for _, domain := range config.Domains {
				err = client.CreateOrUpdateRecord(context.Background(), domain, ip)
				if err != nil {
					return err
				}
			}

			return nil
		})
}

type CloudflareClient struct {
	client         *cloudflare.API
	zoneIdentifier *cloudflare.ResourceContainer
}

func (c CloudflareClient) CreateOrUpdateRecord(ctx context.Context, domain string, ip *net.IP) error {
	records, _, err := c.client.ListDNSRecords(ctx, c.zoneIdentifier, cloudflare.ListDNSRecordsParams{
		Type: "A",
	})

	if err != nil {
		return err
	}

	// prune the old records
	for _, record := range records {
		if record.Name == domain {
			log.Warnf("remove the old records %s(%s)", record.Name, record.Content)
			_ = c.DeleteRecord(ctx, record.ID)
		}
	}

	return c.CreateRecord(ctx, domain, ip)
}

func (c CloudflareClient) UpdateRecord(ctx context.Context, recordID string, ip *net.IP) error {
	log.Infof("update the cloudflare record %s with ip %s", recordID, ip)
	_, err := c.client.UpdateDNSRecord(ctx, c.zoneIdentifier, cloudflare.UpdateDNSRecordParams{
		ID:      recordID,
		Content: ip.String(),
	})
	return err
}

func (c CloudflareClient) DeleteRecord(ctx context.Context, recordID string) error {
	log.Warnf("remove the old record %s", recordID)
	return c.client.DeleteDNSRecord(ctx, c.zoneIdentifier, recordID)
}

func (c CloudflareClient) CreateRecord(ctx context.Context, domain string, ip *net.IP) error {
	log.Infof("create a new record for domain %s with ip %s", domain, ip)
	_, err := c.client.CreateDNSRecord(ctx, c.zoneIdentifier, cloudflare.CreateDNSRecordParams{
		Type:      "A",
		Name:      domain,
		Content:   ip.String(),
		TTL:       60,
		Proxiable: false,
	})

	return err
}

func NewCloudflareDNSClient(config *ddns.TargetConfig) (*CloudflareClient, error) {
	client, err := cloudflare.NewWithAPIToken(config.Token)
	if err != nil {
		return nil, err
	}

	zoneIdentifier := cloudflare.ZoneIdentifier(config.Key)

	// check zone identifier whether is fine
	_, _, err = client.ListDNSRecords(context.TODO(), zoneIdentifier, cloudflare.ListDNSRecordsParams{})
	if err != nil {
		return nil, err
	}

	return &CloudflareClient{
		client:         client,
		zoneIdentifier: zoneIdentifier,
	}, nil
}

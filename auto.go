package simplyddns

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
)

// Use one family-specific endpoint per provider; ident's tnedi mirror must
// not receive a second vote.
var autoIPv4Sources = []string{
	"ipify", "ipsb", "opendns", "icanhazip-ipv4",
	"ident-ipv4", "myexternalip-ipv4", "wtfismyip-ipv4",
}

var autoIPv6Sources = []string{
	"ipify-ipv6", "ipsb-ipv6", "opendns-ipv6", "icanhazip-ipv6",
	"ident-ipv6", "ipinfo-ipv6", "wtfismyip-ipv6",
}

func init() {
	for _, name := range []string{DefaultSourceType, "auto-ipv4", "auto-ipv6"} {
		pool, family := automaticSourcePool(name)
		err := RegisterSourceFunc(name, func(ctx context.Context, config *SourceConfig) (*net.IP, error) {
			count, err := automaticSourceCount(len(pool))
			if err != nil {
				return nil, err
			}
			return queryAutomaticSources(ctx, config, pool, family, count)
		})
		if err != nil {
			panic(err)
		}
	}
}

func automaticSourcePool(name string) ([]string, int) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case DefaultSourceType, "auto-ipv4":
		return autoIPv4Sources, 4
	case "auto-ipv6":
		return autoIPv6Sources, 6
	default:
		return nil, 0
	}
}

func automaticSourceCount(poolSize int) (int, error) {
	value, configured := os.LookupEnv(AutoSourceCountEnv)
	if !configured {
		value = strconv.Itoa(DefaultAutoSourceCount)
	}
	count, err := strconv.Atoi(value)
	if err != nil || count < 3 || count%2 == 0 || count > poolSize {
		return 0, fmt.Errorf("%s must be an odd integer between 3 and %d (got %q)", AutoSourceCountEnv, poolSize, value)
	}
	return count, nil
}

func validateSourceType(sourceType string) error {
	if strings.TrimSpace(sourceType) == "" {
		sourceType = DefaultSourceType
	}
	names := strings.Split(sourceType, ",")
	for _, name := range names {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("source names cannot be empty")
		}
		pool, _ := automaticSourcePool(name)
		if pool == nil {
			continue
		}
		if len(names) != 1 {
			return fmt.Errorf("automatic source types cannot be combined with other sources")
		}
		if _, err := automaticSourceCount(len(pool)); err != nil {
			return err
		}
	}
	return nil
}

func queryAutomaticSources(ctx context.Context, config *SourceConfig, pool []string, family, count int) (*net.IP, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	var workers sync.WaitGroup
	defer func() {
		cancel()
		workers.Wait()
	}()

	type result struct {
		name string
		addr net.IP
		err  error
	}
	results := make(chan result, count)
	selected := rand.Perm(len(pool))[:count]
	for _, index := range selected {
		name := pool[index]
		workers.Go(func() {
			r := result{name: name}
			fn, err := SourceFuncByName(name)
			if err == nil {
				err = ctx.Err()
			}
			if err == nil {
				var addr *net.IP
				addr, err = fn(ctx, config)
				if err == nil && addr != nil {
					r.addr = slices.Clone(*addr)
				}
			}
			r.err = err
			results <- r
		})
	}

	quorum := count/2 + 1
	votes := make(map[string]int)
	var failures []error
	for range count {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case r := <-results:
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if r.err == nil && r.addr.To16() == nil {
				r.err = fmt.Errorf("source returned an invalid IP address")
			}
			if r.err == nil && (r.addr.To4() != nil) != (family == 4) {
				r.err = fmt.Errorf("source returned the wrong address family, expected IPv%d", family)
			}
			if r.err != nil {
				err := fmt.Errorf("source %s: %w", r.name, r.err)
				log.Warnf("automatic source failed: %v", err)
				failures = append(failures, err)
				continue
			}
			address := r.addr.String()
			votes[address]++
			log.Debugf("automatic IPv%d source %s voted for %s (%d/%d)", family, r.name, address, votes[address], quorum)
			// Once a strict majority agrees, remaining votes cannot change it.
			if votes[address] >= quorum {
				return &r.addr, nil
			}
		}
	}
	err := fmt.Errorf("automatic IPv%d source quorum not reached: need %d of %d matching votes, got %v", family, quorum, count, votes)
	return nil, errors.Join(err, errors.Join(failures...))
}

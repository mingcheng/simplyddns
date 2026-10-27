package simplyddns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func registerAutomaticTestSources(t *testing.T, funcs []SourceFunc) []string {
	t.Helper()
	names := make([]string, len(funcs))
	for i, fn := range funcs {
		names[i] = fmt.Sprintf("auto-test-%s-%d", strings.ToLower(t.Name()), i)
		require.NoError(t, RegisterSourceFunc(names[i], fn))
	}
	t.Cleanup(func() {
		registryMu.Lock()
		defer registryMu.Unlock()
		for _, name := range names {
			delete(sourceFuncs, name)
		}
	})
	return names
}

func TestAutomaticSourceCount(t *testing.T) {
	t.Run("unset", func(t *testing.T) {
		t.Setenv(AutoSourceCountEnv, "")
		require.NoError(t, os.Unsetenv(AutoSourceCountEnv))
		count, err := automaticSourceCount(7)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})
	for _, value := range []string{
		"3", "5", "7", "", " ", "0", "1", "2", "4", "6", "8", "9",
		"-3", "three", "3.0", "3e0", "999999999999999999999999999999",
	} {
		t.Run(fmt.Sprintf("value=%q", value), func(t *testing.T) {
			t.Setenv(AutoSourceCountEnv, value)
			count, err := automaticSourceCount(7)
			switch value {
			case "3":
				require.NoError(t, err)
				assert.Equal(t, 3, count)
			case "5":
				require.NoError(t, err)
				assert.Equal(t, 5, count)
			case "7":
				require.NoError(t, err)
				assert.Equal(t, 7, count)
			default:
				require.ErrorContains(t, err, AutoSourceCountEnv)
				assert.Zero(t, count)
			}
		})
	}
	t.Run("candidate limit", func(t *testing.T) {
		t.Setenv(AutoSourceCountEnv, "7")
		_, err := automaticSourceCount(5)
		require.ErrorContains(t, err, "between 3 and 5")
	})
}

func TestAutomaticSourceDocumentation(t *testing.T) {
	for _, filename := range []string{"README.md", "README_ZH.md"} {
		t.Run(filename, func(t *testing.T) {
			data, err := os.ReadFile(filename)
			require.NoError(t, err)
			text := string(data)
			assert.Contains(t, text, AutoSourceCountEnv)
			for _, count := range []int{3, 5, 7} {
				row := fmt.Sprintf(`(?m)^\|[ \t]*%d[ \t]*\|[ \t]*%d[ \t]*\|[ \t]*\r?$`, count, count/2+1)
				assert.Regexp(t, row, text, "missing configuration choice for %d sources", count)
			}
			for _, name := range []string{"auto", "auto-ipv4", "auto-ipv6"} {
				assert.Contains(t, text, "`"+name+"`")
			}
			assert.Contains(t, text, "example/auto.yml")
		})
	}
}

func TestAutomaticSourceVoting(t *testing.T) {
	unavailable := errors.New("provider unavailable")
	source := func(value string, err error) SourceFunc {
		return func(context.Context, *SourceConfig) (*net.IP, error) {
			if value == "" {
				return nil, err
			}
			ip := net.ParseIP(value)
			return &ip, err
		}
	}
	a := source("192.0.2.1", nil)
	b := source("192.0.2.2", nil)
	c := source("192.0.2.3", nil)
	v6 := source("2001:db8::1", nil)
	bad := source("", unavailable)
	for _, tc := range []struct {
		name    string
		sources []SourceFunc
		family  int
		want    string
	}{
		{"unanimous", []SourceFunc{a, a, a}, 4, "192.0.2.1"},
		{"majority", []SourceFunc{a, b, a}, 4, "192.0.2.1"},
		{"majority with failure", []SourceFunc{bad, a, a}, 4, "192.0.2.1"},
		{"only one valid result", []SourceFunc{bad, a, bad}, 4, ""},
		{"tie with failure", []SourceFunc{a, bad, b}, 4, ""},
		{"three different addresses", []SourceFunc{a, b, c}, 4, ""},
		{"all failed", []SourceFunc{bad, bad, bad}, 4, ""},
		{"nil results", []SourceFunc{a, source("", nil), source("", nil)}, 4, ""},
		{"invalid results", []SourceFunc{a, source("invalid", nil), source("invalid", nil)}, 4, ""},
		{"errors invalidate addresses", []SourceFunc{a, source("192.0.2.1", unavailable), bad}, 4, ""},
		{"mapped IPv4", []SourceFunc{a, source("::ffff:192.0.2.1", nil), b}, 4, "192.0.2.1"},
		{"IPv6", []SourceFunc{v6, source("2001:0db8:0:0:0:0:0:1", nil), bad}, 6, "2001:db8::1"},
		{"IPv4 rejects IPv6 votes", []SourceFunc{a, v6, v6}, 4, ""},
		{"IPv6 rejects IPv4 votes", []SourceFunc{a, a, v6}, 6, ""},
		{"IPv6 rejects mapped IPv4", []SourceFunc{a, source("::ffff:192.0.2.1", nil), v6}, 6, ""},
		{"five require three votes", []SourceFunc{a, b, a, c, a}, 4, "192.0.2.1"},
		{"five tolerate two failures", []SourceFunc{a, bad, a, bad, a}, 4, "192.0.2.1"},
		{"five reject plurality", []SourceFunc{a, a, b, c, bad}, 4, ""},
		{"five reject tie", []SourceFunc{a, a, b, b, c}, 4, ""},
		{"failures do not lower quorum", []SourceFunc{a, a, bad, bad, bad}, 4, ""},
		{"seven require four votes", []SourceFunc{a, a, a, a, b, b, b}, 4, "192.0.2.1"},
		{"seven reject three votes", []SourceFunc{a, a, a, bad, bad, bad, bad}, 4, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := registerAutomaticTestSources(t, tc.sources)
			ip, err := queryAutomaticSources(context.Background(), nil, pool, tc.family, len(pool))
			if tc.want == "" {
				require.ErrorContains(t, err, "quorum not reached")
				assert.Nil(t, ip)
				if tc.name == "all failed" {
					assert.ErrorIs(t, err, unavailable)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, ip)
			assert.Equal(t, tc.want, ip.String())
		})
	}
}

func TestAutomaticSourceSamplingAndConcurrency(t *testing.T) {
	for _, count := range []int{3, 5, 7} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				calls := make(chan int, 7)
				release := make(chan struct{})
				funcs := make([]SourceFunc, 7)
				for i := range funcs {
					funcs[i] = func(ctx context.Context, _ *SourceConfig) (*net.IP, error) {
						calls <- i
						select {
						case <-release:
							ip := net.ParseIP("192.0.2.1")
							return &ip, nil
						case <-ctx.Done():
							return nil, ctx.Err()
						}
					}
				}
				pool := registerAutomaticTestSources(t, funcs)
				original := slices.Clone(pool)
				done := make(chan error, 1)
				go func() {
					ip, err := queryAutomaticSources(context.Background(), nil, pool, 4, count)
					assert.NotNil(t, ip)
					done <- err
				}()
				synctest.Wait()
				assert.Len(t, calls, count, "all selected sources must start before any responds")
				seen := make(map[int]bool)
				for range len(calls) {
					index := <-calls
					assert.False(t, seen[index], "a provider must not get more than one vote")
					seen[index] = true
				}
				close(release)
				require.NoError(t, <-done)
				assert.Equal(t, original, pool, "sampling must not mutate the candidate pool")
			})
		})
	}
}

func TestAutomaticSourceCancellation(t *testing.T) {
	t.Run("before requests", func(t *testing.T) {
		calls := 0
		fn := func(context.Context, *SourceConfig) (*net.IP, error) {
			calls++
			return nil, nil
		}
		pool := registerAutomaticTestSources(t, []SourceFunc{fn, fn, fn})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := queryAutomaticSources(ctx, nil, pool, 4, 3)
		require.ErrorIs(t, err, context.Canceled)
		assert.Zero(t, calls)
	})
	for _, name := range []string{"shutdown", "deadline"} {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				entered := make(chan struct{}, 3)
				finished := make(chan struct{}, 3)
				fn := func(ctx context.Context, _ *SourceConfig) (*net.IP, error) {
					entered <- struct{}{}
					<-ctx.Done()
					finished <- struct{}{}
					return nil, ctx.Err()
				}
				pool := registerAutomaticTestSources(t, []SourceFunc{fn, fn, fn})
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				done := make(chan error, 1)
				go func() {
					_, err := queryAutomaticSources(ctx, nil, pool, 4, 3)
					done <- err
				}()
				synctest.Wait()
				assert.Len(t, entered, 3)
				want := context.Canceled
				if name == "deadline" {
					want = context.DeadlineExceeded
					time.Sleep(DefaultTimeout)
				} else {
					cancel()
				}
				require.ErrorIs(t, <-done, want)
				assert.Len(t, finished, 3, "all in-flight requests must be stopped before returning")
			})
		})
	}
	t.Run("majority cancels remaining request", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			entered, finished := make(chan struct{}), make(chan struct{})
			slow := func(ctx context.Context, _ *SourceConfig) (*net.IP, error) {
				close(entered)
				<-ctx.Done()
				close(finished)
				return nil, ctx.Err()
			}
			good := func(context.Context, *SourceConfig) (*net.IP, error) {
				<-entered
				ip := net.ParseIP("192.0.2.1")
				return &ip, nil
			}
			pool := registerAutomaticTestSources(t, []SourceFunc{slow, good, good})
			ip, err := queryAutomaticSources(context.Background(), nil, pool, 4, 3)
			require.NoError(t, err)
			assert.Equal(t, "192.0.2.1", ip.String())
			select {
			case <-finished:
			default:
				t.Fatal("the remaining request was not cancelled")
			}
		})
	})
	t.Run("cancellation overrides successful responses", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		fn := func(context.Context, *SourceConfig) (*net.IP, error) {
			cancel()
			ip := net.ParseIP("192.0.2.1")
			return &ip, nil
		}
		pool := registerAutomaticTestSources(t, []SourceFunc{fn, fn, fn})
		ip, err := queryAutomaticSources(ctx, nil, pool, 4, 3)
		require.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, ip)
	})
}

func TestAutomaticJobRequiresQuorumBeforeUpdating(t *testing.T) {
	good := func(context.Context, *SourceConfig) (*net.IP, error) {
		ip := net.ParseIP("192.0.2.1")
		return &ip, nil
	}
	bad := func(context.Context, *SourceConfig) (*net.IP, error) {
		return nil, errors.New("provider unavailable")
	}
	pool := registerAutomaticTestSources(t, []SourceFunc{good, bad, bad})
	job, err := NewJob(testJobConfig(t))
	require.NoError(t, err)
	defer job.Stop()
	job.SourceFunc = []SourceFunc{func(ctx context.Context, config *SourceConfig) (*net.IP, error) {
		return queryAutomaticSources(ctx, config, pool, 4, 3)
	}}
	updates := 0
	job.TargetFunc = func(context.Context, *net.IP, *TargetConfig) error {
		updates++
		return nil
	}
	require.ErrorContains(t, job.run(context.Background()), "need 2 of 3 matching votes")
	assert.Zero(t, updates)
	assert.Nil(t, job.lastIP)
}

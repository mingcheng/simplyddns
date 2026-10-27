package simplyddns

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testJobConfig(t *testing.T) JobConfig {
	t.Helper()
	NewLogger()
	name := strings.ToLower(t.Name())
	require.NoError(t, RegisterSourceFunc(name, func(context.Context, *SourceConfig) (*net.IP, error) {
		ip := net.ParseIP("192.0.2.1")
		return &ip, nil
	}))
	require.NoError(t, RegisterTargetFunc(name, func(context.Context, *net.IP, *TargetConfig) error {
		return nil
	}))
	t.Cleanup(func() {
		registryMu.Lock()
		defer registryMu.Unlock()
		delete(sourceFuncs, name)
		delete(targetFuncs, name)
	})
	return JobConfig{Source: SourceConfig{Type: name}, Target: TargetConfig{Type: name}}
}

func TestNewJob(t *testing.T) {
	for _, tc := range []struct {
		name     string
		interval uint
		want     uint
	}{
		{"default", 0, DefaultInterval},
		{"minimum", 1, MinInterval},
		{"explicit", 120, 120},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := testJobConfig(t)
			config.Source.Interval = tc.interval
			job, err := NewJob(config)
			require.NoError(t, err)
			t.Cleanup(job.Stop)
			assert.Equal(t, tc.want, job.Config.Source.Interval)
		})
	}
}

func TestNewJobRejectsInvalidConfig(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*JobConfig)
	}{
		{"domain", func(c *JobConfig) { c.Target.Domains = []string{"https://example.com/path"} }},
		{"empty source", func(c *JobConfig) { c.Source.Type = "" }},
		{"unknown source", func(c *JobConfig) { c.Source.Type = "missing-source" }},
		{"empty target", func(c *JobConfig) { c.Target.Type = "" }},
		{"unknown target", func(c *JobConfig) { c.Target.Type = "missing-target" }},
		{"overflowing interval", func(c *JobConfig) { c.Source.Interval = ^uint(0) }},
		{"webhook URL", func(c *JobConfig) { c.WebHook.Url = "://invalid" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := testJobConfig(t)
			tc.change(&config)
			assert.NotPanics(t, func() {
				job, err := NewJob(config)
				if job != nil {
					job.Stop()
				}
				assert.Error(t, err)
			})
		})
	}
}

func TestJobStopIsIdempotent(t *testing.T) {
	job, err := NewJob(testJobConfig(t))
	require.NoError(t, err)
	stopped := make(chan struct{})
	go func() {
		job.Stop()
		job.Stop()
		job.Start(context.Background())
		job.Stop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("stopping an idle or already stopped job must not block")
	}
}

func TestJobSource(t *testing.T) {
	NewLogger()
	valid := net.ParseIP("192.0.2.1")
	other := net.ParseIP("192.0.2.2")
	invalid := net.IP{1, 2, 3}
	var empty net.IP
	source := func(ip *net.IP, err error) SourceFunc {
		return func(context.Context, *SourceConfig) (*net.IP, error) { return ip, err }
	}
	good := source(&valid, nil)
	bad := source(nil, errors.New("source unavailable"))
	for _, tc := range []struct {
		name    string
		sources []SourceFunc
		wantErr bool
	}{
		{"empty", nil, true},
		{"single", []SourceFunc{good}, false},
		{"agreement", []SourceFunc{good, good}, false},
		{"disagreement", []SourceFunc{good, source(&other, nil)}, true},
		{"fallback", []SourceFunc{bad, good}, false},
		{"all failed", []SourceFunc{bad, bad}, true},
		{"nil result", []SourceFunc{source(nil, nil)}, true},
		{"empty IP", []SourceFunc{source(&empty, nil)}, true},
		{"malformed IP", []SourceFunc{source(&invalid, nil)}, true},
		{"too many errors", []SourceFunc{bad, bad, good, good}, true},
		{"invalid results count as errors", []SourceFunc{source(nil, nil), source(&empty, nil), good, good}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			job := Job{SourceFunc: tc.sources}
			ip, err := job.Source(context.Background(), &SourceConfig{})
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, ip)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, ip)
			assert.True(t, valid.Equal(*ip))
		})
	}
	t.Run("cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		job := Job{SourceFunc: []SourceFunc{good}}
		_, err := job.Source(ctx, &SourceConfig{})
		require.ErrorIs(t, err, context.Canceled)
	})
}

func TestJobRunCachesOnlySuccessfulUpdates(t *testing.T) {
	job, err := NewJob(testJobConfig(t))
	require.NoError(t, err)
	defer job.Stop()
	calls := 0
	job.TargetFunc = func(context.Context, *net.IP, *TargetConfig) error {
		calls++
		if calls == 1 {
			return errors.New("temporary provider error")
		}
		return nil
	}
	require.Error(t, job.run(context.Background()))
	assert.Nil(t, job.lastIP)
	require.NoError(t, job.run(context.Background()))
	require.NoError(t, job.run(context.Background()))
	assert.Equal(t, 2, calls)

	job.SourceFunc = []SourceFunc{func(context.Context, *SourceConfig) (*net.IP, error) {
		ip := net.ParseIP("192.0.2.2")
		return &ip, nil
	}}
	require.NoError(t, job.run(context.Background()))
	assert.Equal(t, 3, calls)
	assert.Equal(t, "192.0.2.2", job.lastIP.String())
}

func TestJobWebhookFailureDoesNotRepeatDNSUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()
	job, err := NewJob(testJobConfig(t))
	require.NoError(t, err)
	defer job.Stop()
	job.Config.WebHook.Url = server.URL
	calls := 0
	job.TargetFunc = func(context.Context, *net.IP, *TargetConfig) error { calls++; return nil }
	require.ErrorContains(t, job.run(context.Background()), "webhook")
	require.NoError(t, job.run(context.Background()))
	assert.Equal(t, 1, calls)
}

func TestJobStopsInFlightSource(t *testing.T) {
	for _, stopJob := range []bool{false, true} {
		name := "context"
		if stopJob {
			name = "stop"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				job, err := NewJob(testJobConfig(t))
				require.NoError(t, err)
				defer job.Stop()
				entered := make(chan struct{})
				job.SourceFunc = []SourceFunc{func(ctx context.Context, _ *SourceConfig) (*net.IP, error) {
					close(entered)
					<-ctx.Done()
					return nil, ctx.Err()
				}}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				done := make(chan struct{})
				go func() { job.Start(ctx); close(done) }()
				time.Sleep(time.Duration(job.Config.Source.Interval) * time.Second)
				synctest.Wait()
				select {
				case <-entered:
				default:
					t.Fatal("the source did not run on the first configured tick")
				}
				if stopJob {
					job.Stop()
				} else {
					cancel()
				}
				synctest.Wait()
				select {
				case <-done:
				default:
					t.Fatal("shutdown did not cancel the in-flight source")
				}
			})
		})
	}
}

func TestJobCycleTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		job, err := NewJob(testJobConfig(t))
		require.NoError(t, err)
		defer job.Stop()
		job.timeout = time.Second
		timedOut := make(chan error, 1)
		job.SourceFunc = []SourceFunc{func(ctx context.Context, _ *SourceConfig) (*net.IP, error) {
			<-ctx.Done()
			timedOut <- ctx.Err()
			return nil, ctx.Err()
		}}
		done := make(chan struct{})
		go func() { job.Start(context.Background()); close(done) }()
		time.Sleep(time.Duration(job.Config.Source.Interval)*time.Second + time.Second)
		synctest.Wait()
		select {
		case err := <-timedOut:
			require.ErrorIs(t, err, context.DeadlineExceeded)
		default:
			t.Fatal("job did not enforce its per-cycle timeout")
		}
		job.Stop()
		<-done
	})
}

func TestNewJobCopiesAndNormalizesDomains(t *testing.T) {
	config := testJobConfig(t)
	config.Target.Domains = []string{"Home.Example.COM."}
	job, err := NewJob(config)
	require.NoError(t, err)
	defer job.Stop()
	config.Target.Domains[0] = "other.example.com"
	assert.Equal(t, []string{"home.example.com"}, job.Config.Target.Domains)
}

func TestJobRunWebhook(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusCreated, http.StatusNoContent, http.StatusBadRequest} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var payload struct {
				Address string    `json:"address"`
				Domains string    `json:"domains"`
				Now     time.Time `json:"now"`
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "192.0.2.1", r.Header.Get("Address"))
				assert.Equal(t, "a.example.com,b.example.com", r.Header.Get("Domains"))
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
				w.WriteHeader(status)
			}))
			defer server.Close()
			job := Job{Config: &JobConfig{WebHook: WebHook{Url: server.URL, Token: "test-token"}}}
			err := job.RunWebhook(context.Background(), "192.0.2.1", []string{"a.example.com", "b.example.com"})
			if status >= 300 {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, "192.0.2.1", payload.Address)
			assert.Equal(t, "a.example.com,b.example.com", payload.Domains)
			assert.False(t, payload.Now.IsZero())
		})
	}
	t.Run("basic auth", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, password, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "user", user)
			assert.Equal(t, "password", password)
		}))
		defer server.Close()
		job := Job{Config: &JobConfig{WebHook: WebHook{Url: server.URL, UserName: "user", Password: "password"}}}
		require.NoError(t, job.RunWebhook(context.Background(), "192.0.2.1", nil))
	})
}

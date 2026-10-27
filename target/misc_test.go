package target

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return fn(r) }

func mockTransport(t *testing.T, transport http.RoundTripper) {
	t.Helper()
	original := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = original })
}

func mockAPI(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	mockTransport(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if err := r.Context().Err(); err != nil {
			return nil, err
		}
		recorder := httptest.NewRecorder()
		handler(recorder, r)
		response := recorder.Result()
		response.Request = r
		return response, nil
	}))
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	assert.NoError(t, json.NewEncoder(w).Encode(value))
}

func TestRecordType(t *testing.T) {
	for _, tc := range []struct{ address, want string }{
		{"192.0.2.1", "A"}, {"::ffff:192.0.2.1", "A"}, {"2001:db8::1", "AAAA"}, {"invalid", ""},
	} {
		ip := net.ParseIP(tc.address)
		kind, err := recordType(&ip)
		if tc.want == "" {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.Equal(t, tc.want, kind)
		}
	}
	_, err := recordType(nil)
	require.Error(t, err)
}

func TestTargetsRejectInvalidInput(t *testing.T) {
	for _, name := range []string{"cloudflare", "alidns", "namedotcom"} {
		t.Run(name, func(t *testing.T) {
			mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("invalid target configuration must not make HTTP requests")
				w.WriteHeader(http.StatusBadRequest)
			})
			fn, err := simplyddns.TargetFuncByName(name)
			require.NoError(t, err)
			ip := net.ParseIP("192.0.2.1")
			require.Error(t, fn(context.Background(), nil, &simplyddns.TargetConfig{}))
			require.Error(t, fn(context.Background(), &ip, nil))
			require.Error(t, fn(context.Background(), &ip, &simplyddns.TargetConfig{}))
			require.Error(t, fn(context.Background(), &ip, &simplyddns.TargetConfig{Key: "key", Token: "token"}))
		})
	}
}

func TestTargetCancellation(t *testing.T) {
	for _, name := range []string{"cloudflare", "alidns", "namedotcom", "sleep"} {
		t.Run(name, func(t *testing.T) {
			mockTransport(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
				<-r.Context().Done()
				return nil, r.Context().Err()
			}))
			fn, err := simplyddns.TargetFuncByName(name)
			require.NoError(t, err)
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()
			ip := net.ParseIP("192.0.2.1")
			err = fn(ctx, &ip, &simplyddns.TargetConfig{Key: "key", Token: "token", Domains: []string{"home.example.com"}})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "context deadline exceeded")
		})
	}
}

func TestContextTransportKeepsBodyContextAlive(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	var requestContext context.Context
	transport := contextTransport{ctx: parent, base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requestContext = r.Context()
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("body"))}, nil
	})}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)
	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	require.NoError(t, requestContext.Err())
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "body", string(data))
	require.NoError(t, resp.Body.Close())
	require.ErrorIs(t, requestContext.Err(), context.Canceled)
}

func TestContextTransportClosesCancelledRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	body := &trackedBody{Reader: strings.NewReader("request")}
	req, err := http.NewRequest(http.MethodPost, "https://example.com", body)
	require.NoError(t, err)
	transport := contextTransport{ctx: ctx, base: roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Error("a cancelled request must not reach the underlying transport")
		return nil, context.Canceled
	})}
	_, err = transport.RoundTrip(req)
	require.ErrorIs(t, err, context.Canceled)
	assert.True(t, body.closed)
}

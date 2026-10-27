/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: namedotcom_test.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created:  Saturday, December 26th 2020, 11:04:07 pm
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:15:30
 */

package target

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/mingcheng/simplyddns"
	namedotcom "github.com/namedotcom/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNameComContinuesAfterUnchangedRecord(t *testing.T) {
	var methods []string
	mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		if r.Method == http.MethodGet {
			assert.Equal(t, "/v4/domains/example.com/records", r.URL.Path)
			assert.Equal(t, "1", r.URL.Query().Get("page"))
			assert.Equal(t, "100", r.URL.Query().Get("perPage"))
			writeJSON(t, w, http.StatusOK, namedotcom.ListRecordsResponse{Records: []*namedotcom.Record{
				{ID: 1, DomainName: "example.com", Host: "a", Type: "A", Answer: "192.0.2.1"},
				{ID: 2, DomainName: "example.com", Host: "b", Type: "A", Answer: "192.0.2.2", TTL: 3600},
			}})
			return
		}
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/v4/domains/example.com/records/2", r.URL.Path)
		var record namedotcom.Record
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&record))
		assert.Equal(t, "192.0.2.1", record.Answer)
		assert.Equal(t, uint32(3600), record.TTL)
		writeJSON(t, w, http.StatusOK, record)
	})
	fn, err := simplyddns.TargetFuncByName("namedotcom")
	require.NoError(t, err)
	ip := net.ParseIP("192.0.2.1")
	require.NoError(t, fn(context.Background(), &ip, &simplyddns.TargetConfig{
		Key: "user", Token: "token", Domains: []string{"a.example.com", "b.example.com"},
	}))
	assert.Equal(t, []string{http.MethodGet, http.MethodPut}, methods)
}

func TestNameComCreatesIPv6Apex(t *testing.T) {
	var methods []string
	mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		if r.Method == http.MethodGet {
			writeJSON(t, w, http.StatusOK, namedotcom.ListRecordsResponse{Records: []*namedotcom.Record{
				{ID: 1, DomainName: "example.com", Type: "A", Answer: "192.0.2.1"},
			}})
			return
		}
		assert.Equal(t, http.MethodPost, r.Method)
		var record namedotcom.Record
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&record))
		assert.Equal(t, "AAAA", record.Type)
		assert.Equal(t, "2001:db8::1", record.Answer)
		assert.Equal(t, "example.com", record.DomainName)
		assert.Empty(t, record.Host)
		record.ID = 2
		writeJSON(t, w, http.StatusCreated, record)
	})
	fn, err := simplyddns.TargetFuncByName("namedotcom")
	require.NoError(t, err)
	ip := net.ParseIP("2001:db8::1")
	require.NoError(t, fn(context.Background(), &ip, &simplyddns.TargetConfig{
		Key: "user", Token: "token", Domains: []string{"example.com"},
	}))
	assert.Equal(t, []string{http.MethodGet, http.MethodPost}, methods)
}

func TestNameComPagination(t *testing.T) {
	pages := 0
	mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
		pages++
		assert.Equal(t, strconv.Itoa(pages), r.URL.Query().Get("page"))
		if pages == 1 {
			writeJSON(t, w, http.StatusOK, namedotcom.ListRecordsResponse{NextPage: 2})
			return
		}
		writeJSON(t, w, http.StatusOK, namedotcom.ListRecordsResponse{Records: []*namedotcom.Record{
			{ID: 2, DomainName: "example.com", Host: "@", Type: "A", Answer: "192.0.2.1"},
		}})
	})
	client, err := newNameCom(context.Background(), &simplyddns.TargetConfig{Key: "user", Token: "token"})
	require.NoError(t, err)
	record, err := client.FindRecord("example.com")
	require.NoError(t, err)
	require.NotNil(t, record)
	assert.Equal(t, int32(2), record.ID)
	_, err = client.FindRecord("example.com")
	require.NoError(t, err)
	assert.Equal(t, 2, pages, "a repeated lookup should use this client's cache")
}

func TestNameComLookupFailureDoesNotCreate(t *testing.T) {
	calls := 0
	mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		assert.Equal(t, http.MethodGet, r.Method)
		writeJSON(t, w, http.StatusForbidden, namedotcom.ErrorResponse{Message: "denied"})
	})
	fn, err := simplyddns.TargetFuncByName("namedotcom")
	require.NoError(t, err)
	ip := net.ParseIP("192.0.2.1")
	require.Error(t, fn(context.Background(), &ip, &simplyddns.TargetConfig{
		Key: "user", Token: "token", Domains: []string{"a.example.com", "b.example.com"},
	}))
	assert.Equal(t, 1, calls)
}

func TestNameComCachesAreIsolated(t *testing.T) {
	var updates atomic.Int32
	mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
		user, _, ok := r.BasicAuth()
		assert.True(t, ok)
		id, err := strconv.Atoi(strings.TrimPrefix(user, "user-"))
		assert.NoError(t, err)
		if r.Method == http.MethodGet {
			writeJSON(t, w, http.StatusOK, namedotcom.ListRecordsResponse{Records: []*namedotcom.Record{
				{ID: int32(id), DomainName: "example.com", Host: "home", Type: "A", Answer: "192.0.2.2"},
			}})
			return
		}
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, fmt.Sprintf("/v4/domains/example.com/records/%d", id), r.URL.Path)
		var record namedotcom.Record
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&record))
		assert.Equal(t, int32(id), record.ID)
		updates.Add(1)
		writeJSON(t, w, http.StatusOK, record)
	})
	fn, err := simplyddns.TargetFuncByName("namedotcom")
	require.NoError(t, err)
	var wg sync.WaitGroup
	for id := 1; id <= 8; id++ {
		wg.Go(func() {
			ip := net.ParseIP("192.0.2.1")
			err := fn(context.Background(), &ip, &simplyddns.TargetConfig{
				Key: fmt.Sprintf("user-%d", id), Token: "token", Domains: []string{"home.example.com"},
			})
			assert.NoError(t, err)
		})
	}
	wg.Wait()
	assert.Equal(t, int32(8), updates.Load())
}

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error { b.closed = true; return nil }

func TestNameComClosesResponseBodies(t *testing.T) {
	for _, body := range []string{`{}`, `invalid JSON`} {
		t.Run(body, func(t *testing.T) {
			tracked := &trackedBody{Reader: strings.NewReader(body)}
			mockTransport(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: tracked}, nil
			}))
			client, err := newNameCom(context.Background(), &simplyddns.TargetConfig{Key: "user", Token: "token"})
			require.NoError(t, err)
			err = client.Ping()
			if body == "{}" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			assert.True(t, tracked.closed)
		})
	}
}

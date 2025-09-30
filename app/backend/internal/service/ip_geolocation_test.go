package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/bareuptime/tms/internal/redis"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestIPGeolocationService(t *testing.T, cachePrefix string, cacheTTL time.Duration) (*IPGeolocationService, *miniredis.Miniredis) {
	t.Helper()

	server, err := miniredis.Run()
	require.NoError(t, err)

	redisService := redis.NewService(redis.RedisConfig{
		URL:         fmt.Sprintf("redis://%s", server.Addr()),
		Environment: "development",
	})

	t.Cleanup(func() {
		if err := redisService.Close(); err != nil {
			t.Errorf("failed to close redis service: %v", err)
		}
		server.Close()
	})

	return NewIPGeolocationService(redisService, cachePrefix, cacheTTL), server
}

func TestIPGeolocationService_PrivateIP(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service, _ := newTestIPGeolocationService(t, "ip_geo_test", time.Hour)

	service.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected API call for private IP: %s", req.URL)
		return nil, nil
	})}

	details, err := service.GetIPDetails(ctx, "192.168.1.10")
	require.NoError(t, err)

	require.Equal(t, "192.168.1.10", details.IP)
	require.Equal(t, "Private Network", details.Country)
	require.Equal(t, "USD", details.Currency)
}

func TestIPGeolocationService_CacheHit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cachePrefix := "cache_hit"
	cacheTTL := 30 * time.Minute

	service, _ := newTestIPGeolocationService(t, cachePrefix, cacheTTL)

	cachedDetails := &IPDetails{
		IP:       "8.8.4.4",
		City:     "Mountain View",
		Country:  "United States",
		Currency: "USD",
	}

	payload, err := json.Marshal(cachedDetails)
	require.NoError(t, err)

	cacheKey := fmt.Sprintf("%s:%s", cachePrefix, cachedDetails.IP)
	err = service.redisService.GetClient().Set(ctx, cacheKey, payload, cacheTTL).Err()
	require.NoError(t, err)

	service.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected API call on cache hit: %s", req.URL)
		return nil, nil
	})}

	details, err := service.GetIPDetails(ctx, cachedDetails.IP)
	require.NoError(t, err)

	require.Equal(t, cachedDetails.City, details.City)
	require.Equal(t, cachedDetails.Country, details.Country)
	require.Equal(t, cachedDetails.Currency, details.Currency)

	ttlRemaining, err := service.redisService.GetClient().TTL(ctx, cacheKey).Result()
	require.NoError(t, err)
	require.Greater(t, ttlRemaining, time.Duration(0))
	require.LessOrEqual(t, ttlRemaining, cacheTTL)
}

func TestIPGeolocationService_CacheMissFetchesAndStores(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cachePrefix := "cache_miss"
	cacheTTL := 12 * time.Hour

	service, _ := newTestIPGeolocationService(t, cachePrefix, cacheTTL)

	expected := &IPDetails{
		IP:       "1.1.1.1",
		City:     "Sydney",
		Country:  "Australia",
		Currency: "AUD",
	}

	body, err := json.Marshal(expected)
	require.NoError(t, err)

	var apiCalls int
	service.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		apiCalls++
		require.Equal(t, fmt.Sprintf("https://ipapi.co/%s/json", expected.IP), req.URL.String())

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})}

	details, err := service.GetIPDetails(ctx, expected.IP)
	require.NoError(t, err)
	require.Equal(t, expected.City, details.City)
	require.Equal(t, expected.Currency, details.Currency)
	require.Equal(t, 1, apiCalls)

	cacheKey := fmt.Sprintf("%s:%s", cachePrefix, expected.IP)
	cachedPayload, err := service.redisService.GetClient().Get(ctx, cacheKey).Result()
	require.NoError(t, err)
	require.NotEmpty(t, cachedPayload)

	ttlRemaining, err := service.redisService.GetClient().TTL(ctx, cacheKey).Result()
	require.NoError(t, err)
	require.Greater(t, ttlRemaining, cacheTTL-5*time.Second)
	require.LessOrEqual(t, ttlRemaining, cacheTTL)

	service.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected API call after result was cached: %s", req.URL)
		return nil, nil
	})}

	details, err = service.GetIPDetails(ctx, expected.IP)
	require.NoError(t, err)
	require.Equal(t, expected.City, details.City)
	require.Equal(t, expected.Currency, details.Currency)
}

func TestIPGeolocationService_InvalidateCacheAndStats(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cachePrefix := "cache_stats"
	cacheTTL := time.Hour

	service, _ := newTestIPGeolocationService(t, cachePrefix, cacheTTL)

	cacheKey := fmt.Sprintf("%s:%s", cachePrefix, "9.9.9.9")
	err := service.redisService.GetClient().Set(ctx, cacheKey, "{}", cacheTTL).Err()
	require.NoError(t, err)

	exists, ttlRemaining, err := service.GetCacheStats(ctx, "9.9.9.9")
	require.NoError(t, err)
	require.True(t, exists)
	require.Greater(t, ttlRemaining, time.Duration(0))

	err = service.InvalidateCache(ctx, "9.9.9.9")
	require.NoError(t, err)

	exists, ttlRemaining, err = service.GetCacheStats(ctx, "9.9.9.9")
	require.NoError(t, err)
	require.False(t, exists)
	require.Equal(t, time.Duration(0), ttlRemaining)
}

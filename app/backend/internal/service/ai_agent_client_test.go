package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAiAgentClientProcessMessageStream(t *testing.T) {
	t.Parallel()

	responsesPayload := []AgentResponse{
		{Type: "message", Content: "Hello"},
		{Type: "meta", Content: "done"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/chat/process", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		for _, resp := range responsesPayload {
			data, err := json.Marshal(resp)
			require.NoError(t, err)
			_, err = w.Write([]byte("data: " + string(data) + "\n\n"))
			require.NoError(t, err)
			flusher.Flush()
		}
	}))
	defer srv.Close()

	client := NewAgentClient(srv.URL)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	respCh, errCh := client.ProcessMessageStream(ctx, ChatRequest{Message: "hi"})

	var received []AgentResponse
	var errs []error

	for respCh != nil || errCh != nil {
		select {
		case resp, ok := <-respCh:
			if !ok {
				respCh = nil
				continue
			}
			received = append(received, resp)
		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			errs = append(errs, err)
		case <-ctx.Done():
			t.Fatalf("context cancelled before receiving all responses: %v", ctx.Err())
		}
	}

	require.Empty(t, errs)
	require.Equal(t, responsesPayload, received)
}

func TestAiAgentClientProcessMessageStreamErrorStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()

	client := NewAgentClient(srv.URL)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	respCh, errCh := client.ProcessMessageStream(ctx, ChatRequest{Message: "hi"})

	select {
	case err := <-errCh:
		require.Error(t, err)
		require.Contains(t, err.Error(), "status 500")
	case <-time.After(time.Second):
		t.Fatal("expected error but none received")
	}

	_, ok := <-respCh
	require.False(t, ok)
}

func TestAiAgentClientHealthCheck(t *testing.T) {
	t.Parallel()

	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/chat/health", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer healthy.Close()

	unhealthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer unhealthy.Close()

	client := NewAgentClient(healthy.URL)
	require.NoError(t, client.HealthCheck(context.Background()))

	client = NewAgentClient(unhealthy.URL)
	err := client.HealthCheck(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "status 503")
}

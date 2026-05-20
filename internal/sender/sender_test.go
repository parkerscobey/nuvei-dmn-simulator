package sender

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendPostsFormURLEncodedPayload(t *testing.T) {
	t.Parallel()

	var (
		requestMethod      string
		requestContentType string
		requestBody        string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMethod = r.Method
		requestContentType = r.Header.Get("Content-Type")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll request body error = %v", err)
		}
		requestBody = string(body)

		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	result, err := NewClient(server.Client()).Send(context.Background(), Request{
		TargetURL:      server.URL,
		EncodedPayload: "a=1&b=hello+world",
	})
	if err != nil {
		t.Fatalf("Send error = %v", err)
	}

	if requestMethod != http.MethodPost {
		t.Fatalf("method = %q, want %q", requestMethod, http.MethodPost)
	}
	if requestContentType != ContentTypeFormURLEncoded {
		t.Fatalf("Content-Type = %q, want %q", requestContentType, ContentTypeFormURLEncoded)
	}
	if requestBody != "a=1&b=hello+world" {
		t.Fatalf("body = %q, want %q", requestBody, "a=1&b=hello+world")
	}
	if result.StatusCode != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", result.StatusCode, http.StatusAccepted)
	}
	if result.Body != "ok" {
		t.Fatalf("response body = %q, want %q", result.Body, "ok")
	}
}

func TestSendReturnsValidationError(t *testing.T) {
	t.Parallel()

	_, err := NewClient(nil).Send(context.Background(), Request{})
	if err == nil {
		t.Fatal("Send error = nil, want validation error")
	}
}

func TestSendLimitsResponseBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("abcdef"))
	}))
	defer server.Close()

	result, err := (&Client{HTTPClient: server.Client(), MaxResponseBytes: 3}).Send(context.Background(), Request{
		TargetURL:      server.URL,
		EncodedPayload: "k=v",
	})
	if err != nil {
		t.Fatalf("Send error = %v", err)
	}
	if result.Body != "abc" {
		t.Fatalf("response body = %q, want %q", result.Body, "abc")
	}
}

func TestSendPropagatesTransportError(t *testing.T) {
	t.Parallel()

	_, err := NewClient(nil).Send(context.Background(), Request{
		TargetURL:      "http://127.0.0.1:1",
		EncodedPayload: "a=b",
	})
	if err == nil {
		t.Fatal("Send error = nil, want transport error")
	}
	if !strings.Contains(err.Error(), "send DMN request") {
		t.Fatalf("error = %v, want send DMN request context", err)
	}
}

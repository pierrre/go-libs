package httpurl

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"testing"

	"github.com/pierrre/go-libs/internal/testutil"
)

func TestGet(t *testing.T) {
	for _, tc := range []struct {
		name     string
		req      *http.Request
		expected string
	}{
		{
			name: "Full",
			req: &http.Request{
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
					Path:   "/test",
				},
			},
			expected: "http://example.com/test",
		},
		{
			name: "SchemeTLS",
			req: &http.Request{
				URL: &url.URL{
					Host: "example.com",
					Path: "/test",
				},
				TLS: &tls.ConnectionState{},
			},
			expected: "https://example.com/test",
		},
		{
			name: "SchemeXFP",
			req: &http.Request{
				URL: &url.URL{
					Host: "example.com",
					Path: "/test",
				},
				Header: http.Header{
					"X-Forwarded-Proto": {"https"},
				},
			},
			expected: "https://example.com/test",
		},
		{
			name: "SchemeHTTP",
			req: &http.Request{
				URL: &url.URL{
					Host: "example.com",
					Path: "/test",
				},
			},
			expected: "http://example.com/test",
		},
		{
			name: "Host",
			req: &http.Request{
				URL: &url.URL{
					Scheme: "http",
					Path:   "/test",
				},
				Host: "example.com",
			},
			expected: "http://example.com/test",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expected, err := url.Parse(tc.expected)
			if err != nil {
				t.Fatal(err)
			}
			u := Get(tc.req)
			testutil.Compare(t, "unexpected URL", u, expected)
		})
	}
}

// Package httpurl provides an helper to get the real URL of an HTTP request.
package httpurl

import (
	"net/http"
	"net/url"
)

// Get returns the real URL of the HTTP request.
//
// According to the Go HTTP documentation, the Scheme and Host fields are empty.
// This function returns the expected URL.
func Get(req *http.Request) *url.URL {
	u := copyURL(req.URL)
	scheme(req, u)
	host(req, u)
	return u
}

func copyURL(u *url.URL) *url.URL {
	tmp := *u
	return &tmp
}

func scheme(req *http.Request, u *url.URL) {
	if u.Scheme != "" {
		return
	}
	if req.TLS != nil {
		u.Scheme = "https"
		return
	}
	if req.Header.Get("X-Forwarded-Proto") == "https" {
		u.Scheme = "https"
		return
	}
	u.Scheme = "http"
}

func host(req *http.Request, u *url.URL) {
	if u.Host != "" {
		return
	}
	u.Host = req.Host
}

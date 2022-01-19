package ctxhttpsrv

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pierrre/go-libs/goroutine"
)

func TestListenAndServe(t *testing.T) {
	port := getTestFreePort(t)
	addr := net.JoinHostPort("localhost", strconv.Itoa(port))
	u := (&url.URL{
		Scheme: "http",
		Host:   addr,
	}).String()
	f := func(ctx context.Context, h http.Handler) error {
		return ListenAndServe(ctx, addr, h)
	}
	test(t, u, f)
}

func TestListenAndServeError(t *testing.T) {
	ctx := context.Background()
	addr := "invalid"
	err := ListenAndServe(ctx, addr, nil)
	if err == nil {
		t.Fatal("no error")
	}
}

func TestServe(t *testing.T) {
	l := getTestFreeListener(t)
	defer l.Close() //nolint:errcheck
	port := l.Addr().(*net.TCPAddr).Port
	addr := "localhost:" + strconv.Itoa(port)
	u := (&url.URL{
		Scheme: "http",
		Host:   addr,
	}).String()
	f := func(ctx context.Context, h http.Handler) error {
		return Serve(ctx, l, h)
	}
	test(t, u, f)
}

func TestServeError(t *testing.T) {
	ctx := context.Background()
	l := getTestFreeListener(t)
	_ = l.Close()
	err := Serve(ctx, l, nil)
	if err == nil {
		t.Fatal("no error")
	}
}

func test(tb testing.TB, u string, f func(context.Context, http.Handler) error) {
	tb.Helper()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var hCalled int64
	h := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		atomic.AddInt64(&hCalled, 1)
	})
	waitRequest := goroutine.Go(func() {
		defer cancel()
		time.Sleep(100 * time.Millisecond) // Wait for the server to start, prevents flaky tests.
		_, err := http.Get(u)
		if err != nil {
			tb.Error(err)
			return
		}
	})
	err := f(ctx, h)
	if err != nil {
		tb.Fatal(err)
	}
	waitRequest()
	if hCalled == 0 {
		tb.Fatal("handler not called")
	}
}

func getTestFreePort(tb testing.TB) int {
	tb.Helper()
	l := getTestFreeListener(tb)
	defer l.Close() //nolint:errcheck
	return l.Addr().(*net.TCPAddr).Port
}

func getTestFreeListener(tb testing.TB) *net.TCPListener {
	tb.Helper()
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		tb.Fatal(err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		tb.Fatal(err)
	}
	return l
}

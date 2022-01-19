// Package ctxhttpsrv provides a helper to stop an HTTP Server if a Context is canceled.
// It returns no error if the Context is canceled.
package ctxhttpsrv

import (
	"context"
	"net"
	"net/http"

	"github.com/pierrre/go-libs/errors"
	"github.com/pierrre/go-libs/goroutine"
)

// ListenAndServe is a replacement for net/http.ListenAndServe.
func ListenAndServe(ctx context.Context, addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	return ServerListenAndServe(ctx, srv)
}

// ServerListenAndServe is a replacement for net/http.Server.ListenAndServe.
func ServerListenAndServe(ctx context.Context, srv *http.Server) error {
	return runServer(ctx, srv, srv.ListenAndServe)
}

// Serve is a replacement for net/http.Serve.
func Serve(ctx context.Context, l net.Listener, handler http.Handler) error {
	srv := &http.Server{
		Handler: handler,
	}
	return ServerServe(ctx, srv, l)
}

// ServerServe is a replacement for net/http.Server.Serve.
func ServerServe(ctx context.Context, srv *http.Server, l net.Listener) error {
	return runServer(ctx, srv, func() error {
		return srv.Serve(l)
	})
}

func runServer(ctx context.Context, srv *http.Server, f func() error) error {
	errCh := make(chan error)
	wait := goroutine.Go(func() {
		select {
		case errCh <- f():
		case <-ctx.Done():
		}
	})
	defer wait()
	select {
	case err := <-errCh:
		return errors.WithStack(err)
	case <-ctx.Done():
		err := srv.Shutdown(context.Background())
		return errors.Wrap(err, "shutdown")
	}
}

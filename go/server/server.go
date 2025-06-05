// Package server wraps [http.Server] from "net/http" with a config, defaults,
// graceful shutdown via os signals, and optional lifecycle hooks. It's simple,
// stdlib-first, and not meant for complex cases.
//
// Common Usage:
//
//	srv, err := server.New(&server.Config{
//	  UseTLS:         true,
//	  TLSCertPath: "./cert.pem",
//	  TLSKeyPath:  "./key.pem",
//	  Handler:     myHandler,
//	  OnListen:    func() { /* do something after server starts */ },
//	  OnShutdown:  func() { /* cleanup database connections, websockets, etc. */ },
//	})
//	if err != nil {
//	  log.Fatalf("failed to create server: %v", err)
//	}
//	log.Fatal(srv.Listen())
//
// See [Config] for all options and defaults.
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	Addr        string // Address to listen on. Default is ":80" for HTTP and ":443" for HTTPS
	UseTLS      bool   // Whether to use TLS or not. TLS paths are ignored if false.
	TLSKeyPath  string // TLS key file path
	TLSCertPath string // TLS cert file path

	// Handler, typically a router or middleware chain
	Handler http.Handler

	// Timeouts, zero values are replaced with defaults and negatives passed
	// through to http.Server which treats them as no timeouts.

	ReadTimeout     time.Duration // default 5s
	WriteTimeout    time.Duration // default 10s
	IdleTimeout     time.Duration // default 120s
	ShutdownTimeout time.Duration // default 10s

	// OnListen, if non-nil, is called after the server starts listening. Simple and flexible.
	// Useful for validating the server is up and running, e.g. by checking a health endpoint.
	OnListen      func()
	OnListenDelay time.Duration // Delay after starting the server before calling OnListen. Default is 1s.

	// OnShutdown, if non-nil, is called during server shutdown, after the
	// server has stopped accepting new connections, but before closing idle ones.
	// Note: depending on the shutdown timeout, this may exceed the life of the server.
	// Note: if ShutdownTimeout is <= 0, this will not be called.
	OnShutdown func()
}

type Server struct {
	cfg    *Config      // Configuration for the server
	server *http.Server // The http or https server
}

func New(cfg *Config) (*Server, error) {
	copy := *cfg

	// validate config
	if copy.Handler == nil {
		return nil, fmt.Errorf("handler must be provided")
	}
	if copy.UseTLS {
		if copy.TLSKeyPath == "" || copy.TLSCertPath == "" {
			return nil, fmt.Errorf("TLS key and cert paths must be provided when using TLS")
		}
		if copy.Addr == "" {
			copy.Addr = ":443"
		}
	} else {
		if copy.Addr == "" {
			copy.Addr = ":80"
		}
	}

	// set default timeouts
	if copy.ReadTimeout == 0 {
		copy.ReadTimeout = 5 * time.Second
	}
	if copy.WriteTimeout == 0 {
		copy.WriteTimeout = 10 * time.Second
	}
	if copy.IdleTimeout == 0 {
		copy.IdleTimeout = 120 * time.Second
	}
	if copy.ShutdownTimeout == 0 {
		copy.ShutdownTimeout = 10 * time.Second
	}
	if copy.OnListenDelay == 0 {
		copy.OnListenDelay = 1 * time.Second
	}

	// create http server
	httpServer := &http.Server{
		Addr:         copy.Addr,
		Handler:      copy.Handler,
		ReadTimeout:  copy.ReadTimeout,
		WriteTimeout: copy.WriteTimeout,
		IdleTimeout:  copy.IdleTimeout,
		TLSConfig:    &tls.Config{MinVersion: tls.VersionTLS13},
	}

	// set shutdown hook if provided
	if copy.OnShutdown != nil && copy.ShutdownTimeout > 0 {
		httpServer.RegisterOnShutdown(copy.OnShutdown)
	}

	// return the server
	return &Server{
		cfg:    &copy,
		server: httpServer,
	}, nil
}

func (s *Server) Listen() error {
	// setup chans for listen and shutdown signals
	listenErrCh := make(chan error, 1)
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// start server
	go func() {
		if s.cfg.UseTLS {
			listenErrCh <- s.server.ListenAndServeTLS(s.cfg.TLSCertPath, s.cfg.TLSKeyPath)
		} else {
			listenErrCh <- s.server.ListenAndServe()
		}
	}()

	// setup OnListen. Tiny bit of a hack but should work fine for the vast majority of cases.
	onListenCh := make(chan struct{}, 1)
	if s.cfg.OnListen != nil {
		go func() {
			time.Sleep(s.cfg.OnListenDelay)
			onListenCh <- struct{}{}
		}()
	}

	// handle onListen, shutdown, and listen errors
	for {
		select {
		case <-onListenCh:
			s.cfg.OnListen()
		case <-shutdownCh:
			signal.Stop(shutdownCh)
			if s.cfg.ShutdownTimeout <= 0 {
				return s.server.Close()
			}
			ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
			defer cancel()
			return s.server.Shutdown(ctx) // shutdown causes listen to return [ErrServerClosed] immediately, no need to handle it.
		case err := <-listenErrCh:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				if errors.Is(err, syscall.EADDRINUSE) {
					return fmt.Errorf("address already in use: %w", err)
				}
				if errors.Is(err, syscall.EACCES) {
					return fmt.Errorf("permission denied: %w", err)
				}
				return err
			}
			return nil
		}
	}
}

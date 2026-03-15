package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/ubcent/tyria/internal/admin"
	"github.com/ubcent/tyria/internal/proxy"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Application error: %v", err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	proxyMux := http.NewServeMux()
	proxyMux.Handle("/", proxy.NewHandler("http://localhost:9999"))
	proxyServer := &http.Server{Addr: ":8080", Handler: proxyMux}

	adminMux := http.NewServeMux()
	adminMux.Handle("/admin", admin.NewHandler())
	adminServer := &http.Server{Addr: ":9090", Handler: adminMux}

	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error {
		return proxyServer.ListenAndServe()
	})

	g.Go(func() error {
		<-ctx.Done()
		shutDownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return proxyServer.Shutdown(shutDownCtx)
	})

	g.Go(func() error {
		return adminServer.ListenAndServe()
	})

	g.Go(func() error {
		<-ctx.Done()
		shutDownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return adminServer.Shutdown(shutDownCtx)
	})

	log.Println("Tyria started. Proxy :8080, Admin :9090")
	<-ctx.Done()
	log.Println("Shutting down...")

	return g.Wait()
}

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ubcent/edge.link/internal/admin"
)

func main() {
	var (
		addr        = flag.String("addr", ":3001", "Admin API server address")
		databaseURL = flag.String("db", os.Getenv("POSTGRES_URL"), "PostgreSQL database URL")
	)
	flag.Parse()

	if *databaseURL == "" {
		*databaseURL = "postgres://localhost:5432/edgelink?sslmode=disable"
	}

	// Create admin server
	server, err := admin.NewServer(*databaseURL)
	if err != nil {
		log.Fatalf("Failed to create admin server: %v", err)
	}
	defer server.Close()

	// Start server in goroutine
	go func() {
		if err := server.Start(*addr); err != nil {
			log.Fatalf("Admin server failed: %v", err)
		}
	}()

	log.Printf("Admin API server started on %s", *addr)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Admin API server stopped")
}
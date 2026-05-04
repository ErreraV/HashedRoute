package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	cfg, err := gatewayConfigFromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	gw, conn, err := connectGateway(cfg)
	if err != nil {
		log.Fatalf("fabric gateway: %v", err)
	}
	defer conn.Close()
	defer gw.Close()

	a := &api{
		gw:       gw,
		contract: fabricContract(cfg, gw),
	}

	addr := strings.TrimSpace(os.Getenv("HTTP_ADDR"))
	if addr == "" {
		addr = ":8080"
	}
	srv := &http.Server{Addr: addr, Handler: a.routes()}

	go func() {
		log.Printf("API listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

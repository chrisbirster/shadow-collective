package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/chrisbirster/shadow-collective/internal/config"
)

type Gateway struct {
	cfg config.Config
	log *slog.Logger
}

func New(cfg config.Config, log *slog.Logger) *Gateway {
	return &Gateway{cfg: cfg, log: log}
}

func (g *Gateway) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(g.cfg.HTTP)+len(g.cfg.TCP)+4)

	start := func(name string, fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(); err != nil {
				select {
				case errCh <- fmt.Errorf("%s: %w", name, err):
				case <-ctx.Done():
				}
			}
		}()
	}

	start("health", func() error { return g.serveHealth(ctx) })

	for _, svc := range g.cfg.HTTP {
		svc := svc
		start("http "+svc.Name, func() error { return g.serveHTTP(ctx, svc) })
	}
	for _, svc := range g.cfg.TCP {
		svc := svc
		start("tcp "+svc.Name, func() error { return g.serveTCP(ctx, svc.Name, svc.Listen, svc.Upstream) })
	}
	if g.cfg.DNS.Enabled {
		start("dns udp", func() error { return g.serveDNSUDP(ctx, g.cfg.DNS.Listen, g.cfg.DNS.Upstream) })
		start("dns tcp", func() error { return g.serveTCP(ctx, "dns", g.cfg.DNS.Listen, g.cfg.DNS.Upstream) })
	}

	select {
	case <-ctx.Done():
		wg.Wait()
		return nil
	case err := <-errCh:
		return err
	}
}

func (g *Gateway) serveHealth(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	server := &http.Server{Addr: g.cfg.HealthListen, Handler: mux}
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()
	g.log.Info("health listener ready", "listen", g.cfg.HealthListen)
	err := server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

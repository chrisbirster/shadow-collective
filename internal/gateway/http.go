package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/chrisbirster/shadow-collective/internal/config"
)

func (g *Gateway) serveHTTP(ctx context.Context, svc config.HTTPService) error {
	target, err := url.Parse(svc.Upstream)
	if err != nil {
		return fmt.Errorf("parse upstream: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		req.Host = target.Host
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		g.log.Warn("upstream request failed", "service", svc.Name, "upstream", svc.Upstream, "error", err)
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}

	server := &http.Server{Addr: svc.Listen, Handler: proxy}
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	g.log.Info("http proxy ready", "service", svc.Name, "listen", svc.Listen, "upstream", svc.Upstream)
	err = server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

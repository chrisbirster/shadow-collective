package gateway

import (
	"context"
	"errors"
	"io"
	"net"
	"time"
)

func (g *Gateway) serveTCP(ctx context.Context, name, listenAddr, upstream string) error {
	lc := net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	g.log.Info("tcp proxy ready", "service", name, "listen", listenAddr, "upstream", upstream)
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		go g.proxyTCP(ctx, name, conn, upstream)
	}
}

func (g *Gateway) proxyTCP(ctx context.Context, name string, client net.Conn, upstream string) {
	defer client.Close()

	dialer := net.Dialer{Timeout: 5 * time.Second}
	backend, err := dialer.DialContext(ctx, "tcp", upstream)
	if err != nil {
		g.log.Warn("tcp upstream dial failed", "service", name, "upstream", upstream, "error", err)
		return
	}
	defer backend.Close()

	done := make(chan struct{}, 2)
	copyConn := func(dst, src net.Conn) {
		_, _ = io.Copy(dst, src)
		if tcp, ok := dst.(*net.TCPConn); ok {
			_ = tcp.CloseWrite()
		}
		done <- struct{}{}
	}
	go copyConn(backend, client)
	go copyConn(client, backend)

	select {
	case <-ctx.Done():
	case <-done:
	}
}

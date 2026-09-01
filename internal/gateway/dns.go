package gateway

import (
	"context"
	"errors"
	"net"
	"time"
)

const maxDNSPacket = 65535

func (g *Gateway) serveDNSUDP(ctx context.Context, listenAddr, upstream string) error {
	pc, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		return err
	}
	defer pc.Close()

	go func() {
		<-ctx.Done()
		_ = pc.Close()
	}()

	g.log.Info("dns udp forwarder ready", "listen", listenAddr, "upstream", upstream)
	buf := make([]byte, maxDNSPacket)
	for {
		n, clientAddr, err := pc.ReadFrom(buf)
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		packet := append([]byte(nil), buf[:n]...)
		go g.forwardDNSPacket(ctx, pc, clientAddr, packet, upstream)
	}
}

func (g *Gateway) forwardDNSPacket(ctx context.Context, listener net.PacketConn, clientAddr net.Addr, packet []byte, upstream string) {
	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "udp", upstream)
	if err != nil {
		g.log.Warn("dns upstream dial failed", "upstream", upstream, "error", err)
		return
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	if _, err := conn.Write(packet); err != nil {
		return
	}
	response := make([]byte, maxDNSPacket)
	n, err := conn.Read(response)
	if err != nil {
		return
	}
	_, _ = listener.WriteTo(response[:n], clientAddr)
}

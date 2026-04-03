package proxy

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/xxf098/lite-proxy/tunnel"
)

type LocalSocksClient struct {
	ctx       context.Context
	proxyAddr string
	timeout   time.Duration
}

func NewLocalSocksClient(ctx context.Context, proxyAddr string) *LocalSocksClient {
	return &LocalSocksClient{ctx: ctx, proxyAddr: proxyAddr, timeout: 8 * time.Second}
}

func (c *LocalSocksClient) DialConn(addr *tunnel.Address, _ tunnel.Tunnel) (net.Conn, error) {
	if addr == nil {
		return nil, fmt.Errorf("nil address")
	}
	target := fmt.Sprintf("%s:%d", addr.String(), addr.Port)
	if addr.AddressType == tunnel.DomainName {
		target = fmt.Sprintf("%s:%d", addr.DomainName, addr.Port)
	}
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()
	return dialViaSocks5(ctx, c.proxyAddr, "tcp", target)
}

func (c *LocalSocksClient) Close() error { return nil }

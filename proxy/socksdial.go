package proxy

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	socksVersion5 = 0x05
	socksCmdConnect = 0x01
	socksAtypIPv4 = 0x01
	socksAtypDomain = 0x03
	socksAtypIPv6 = 0x04
	socksAuthNone = 0x00
	socksReplySucceeded = 0x00
	socksReplyGeneralFailure = 0x01
	socksReplyConnectionNotAllowed = 0x02
	socksReplyNetworkUnreachable = 0x03
	socksReplyHostUnreachable = 0x04
	socksReplyConnectionRefused = 0x05
	socksReplyTTLExpired = 0x06
	socksReplyCommandNotSupported = 0x07
	socksReplyAddressTypeNotSupported = 0x08
)

func NewSocks5DialFunc(proxyAddr string, timeout time.Duration) func(network, addr string) (net.Conn, error) {
	return func(network, addr string) (net.Conn, error) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return dialViaSocks5(ctx, proxyAddr, network, addr)
	}
}

func dialViaSocks5(ctx context.Context, proxyAddr, network, dstAddr string) (net.Conn, error) {
	switch strings.ToLower(network) {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, fmt.Errorf("socks5 dial only supports tcp, got %q", network)
	}

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("connect socks5 proxy failed: %w", err)
	}
	ok := false
	defer func() {
		if !ok {
			_ = conn.Close()
		}
	}()
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		_ = conn.SetDeadline(deadline)
	}
	if err := socks5Handshake(conn); err != nil {
		return nil, err
	}
	if err := socks5Connect(conn, dstAddr); err != nil {
		return nil, err
	}
	_ = conn.SetDeadline(time.Time{})
	ok = true
	return conn, nil
}

func socks5Handshake(conn net.Conn) error {
	req := []byte{socksVersion5, 0x01, socksAuthNone}
	if _, err := conn.Write(req); err != nil {
		return fmt.Errorf("write socks5 greeting failed: %w", err)
	}
	resp := make([]byte, 2)
	if _, err := io.ReadFull(conn, resp); err != nil {
		return fmt.Errorf("read socks5 greeting response failed: %w", err)
	}
	if resp[0] != socksVersion5 {
		return fmt.Errorf("invalid socks version: %d", resp[0])
	}
	if resp[1] != socksAuthNone {
		return fmt.Errorf("socks5 auth method not supported: %d", resp[1])
	}
	return nil
}

func socks5Connect(conn net.Conn, dstAddr string) error {
	host, port, err := splitHostPort(dstAddr)
	if err != nil {
		return err
	}
	addrField, atyp, err := encodeSocksAddr(host)
	if err != nil {
		return err
	}
	req := make([]byte, 0, 6+len(addrField))
	req = append(req, socksVersion5, socksCmdConnect, 0x00, atyp)
	req = append(req, addrField...)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	req = append(req, portBytes...)
	if _, err := conn.Write(req); err != nil {
		return fmt.Errorf("write socks5 connect request failed: %w", err)
	}
	head := make([]byte, 4)
	if _, err := io.ReadFull(conn, head); err != nil {
		return fmt.Errorf("read socks5 connect response head failed: %w", err)
	}
	if head[0] != socksVersion5 {
		return fmt.Errorf("invalid socks version in connect response: %d", head[0])
	}
	if head[1] != socksReplySucceeded {
		return fmt.Errorf("socks5 connect failed: %s", replyString(head[1]))
	}
	if err := discardBoundAddr(conn, head[3]); err != nil {
		return fmt.Errorf("read socks5 bound address failed: %w", err)
	}
	return nil
}

func splitHostPort(addr string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid target address %q: %w", addr, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return "", 0, fmt.Errorf("invalid target port %q", portStr)
	}
	return host, port, nil
}

func encodeSocksAddr(host string) ([]byte, byte, error) {
	if i := strings.LastIndex(host, "%"); i >= 0 {
		host = host[:i]
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			return ip4, socksAtypIPv4, nil
		}
		if ip16 := ip.To16(); ip16 != nil {
			return ip16, socksAtypIPv6, nil
		}
	}
	if len(host) == 0 {
		return nil, 0, fmt.Errorf("empty target host")
	}
	if len(host) > 255 {
		return nil, 0, fmt.Errorf("target host too long")
	}
	b := make([]byte, 1+len(host))
	b[0] = byte(len(host))
	copy(b[1:], host)
	return b, socksAtypDomain, nil
}

func discardBoundAddr(conn net.Conn, atyp byte) error {
	var addrLen int
	switch atyp {
	case socksAtypIPv4:
		addrLen = 4
	case socksAtypIPv6:
		addrLen = 16
	case socksAtypDomain:
		lb := make([]byte, 1)
		if _, err := io.ReadFull(conn, lb); err != nil {
			return err
		}
		addrLen = int(lb[0])
	default:
		return fmt.Errorf("unsupported bound atyp: %d", atyp)
	}
	buf := make([]byte, addrLen+2)
	_, err := io.ReadFull(conn, buf)
	return err
}

func replyString(code byte) string {
	switch code {
	case socksReplySucceeded:
		return "succeeded"
	case socksReplyGeneralFailure:
		return "general failure"
	case socksReplyConnectionNotAllowed:
		return "connection not allowed"
	case socksReplyNetworkUnreachable:
		return "network unreachable"
	case socksReplyHostUnreachable:
		return "host unreachable"
	case socksReplyConnectionRefused:
		return "connection refused"
	case socksReplyTTLExpired:
		return "TTL expired"
	case socksReplyCommandNotSupported:
		return "command not supported"
	case socksReplyAddressTypeNotSupported:
		return "address type not supported"
	default:
		return fmt.Sprintf("unknown reply code %d", code)
	}
}

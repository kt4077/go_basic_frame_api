package upload

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const remoteFetchTimeout = 30 * time.Second

// validateRemoteURL 只允许无凭据的 HTTP(S) 地址。目标 IP 的公网校验在实际拨号时再次执行，
// 防止域名解析到内网地址及 DNS rebinding 绕过。
func validateRemoteURL(raw string) (*url.URL, error) {
	u, err := url.ParseRequestURI(raw)
	if err != nil || u.Host == "" {
		return nil, errors.New("远程文件地址无效")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("远程文件地址仅支持 http/https")
	}
	if u.User != nil {
		return nil, errors.New("远程文件地址不能包含用户凭据")
	}
	if strings.TrimSpace(u.Hostname()) == "" {
		return nil, errors.New("远程文件地址缺少主机名")
	}
	return u, nil
}

func safeRemoteHTTPClient() *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	transport := &http.Transport{
		Proxy:                 nil,
		ForceAttemptHTTP2:     true,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		IdleConnTimeout:       30 * time.Second,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, fmt.Errorf("解析远程主机失败: %w", err)
			}
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, fmt.Errorf("解析远程主机失败: %w", err)
			}
			if len(ips) == 0 {
				return nil, errors.New("远程主机没有可用地址")
			}
			for _, item := range ips {
				if !isPublicIP(item.IP) {
					return nil, errors.New("禁止访问内网或保留地址")
				}
			}
			// 所有解析结果均已验证；逐个尝试，避免单个地址不可达导致误失败。
			var lastErr error
			for _, item := range ips {
				conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(item.IP.String(), port))
				if dialErr == nil {
					return conn, nil
				}
				lastErr = dialErr
			}
			return nil, lastErr
		},
	}
	return &http.Client{
		Transport: transport,
		Timeout:   remoteFetchTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("远程文件重定向次数过多")
			}
			_, err := validateRemoteURL(req.URL.String())
			return err
		},
	}
}

func isPublicIP(ip net.IP) bool {
	if ip == nil || !ip.IsGlobalUnicast() || ip.IsLoopback() || ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
		return false
	}
	// Go 的 IsPrivate 不包含运营商共享地址和文档/基准测试保留网段。
	// 对服务端抓取采取 allow-public 策略，避免这些地址被用来探测内部基础设施。
	for _, raw := range []string{
		"0.0.0.0/8", "100.64.0.0/10", "192.0.0.0/24", "192.0.2.0/24",
		"198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "240.0.0.0/4",
		"2001:db8::/32",
	} {
		_, network, _ := net.ParseCIDR(raw)
		if network.Contains(ip) {
			return false
		}
	}
	return true
}

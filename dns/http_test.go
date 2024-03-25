package dns

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
)

func TestHttpViaIP(t *testing.T) {
	ip := "2606:2800:220:1:248:1893:25c8:1946" // 替换为你要访问的IP地址
	url := "http://www.example.com/"           // 替换为你要访问的URL

	resp, err := fetchWithIP(ip, url)
	if err != nil {
		fmt.Println("Error:", err)
		t.Errorf(err.Error())
		return
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Response:", string(body))
}
func TestHttpsViaIP(t *testing.T) {
	ip := "2606:2800:220:1:248:1893:25c8:1946" // 替换为你要访问的IP地址
	url := "https://www.example.com/"          // 替换为你要访问的URL

	resp, err := fetchWithIP(ip, url)
	if err != nil {
		fmt.Println("Error:", err)
		t.Errorf(err.Error())
		return
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Response:", string(body))
}
func fetchWithIP(ip, url string) (*http.Response, error) {
	dialer := &net.Dialer{
		// Timeout:   30 * time.Second,
		// KeepAlive: 30 * time.Second,
	}
	/*  */
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
				if err != nil {
					return nil, err
				}
				fmt.Println("连接成功", host, port, conn.LocalAddr(), conn.RemoteAddr())
				return conn, err
			},
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
				if err != nil {
					return nil, err
				}
				//打印连接成功
				fmt.Println("连接成功", host, port, conn.LocalAddr(), conn.RemoteAddr())
				return tls.Client(conn, &tls.Config{
					ServerName: host, // 使用原始域名，而不是IP地址
					// 如果你需要跳过证书验证，可以设置 InsecureSkipVerify: true
				}), nil
			},
		},
	}

	return client.Get(url)
}

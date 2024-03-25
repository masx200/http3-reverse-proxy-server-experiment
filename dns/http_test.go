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

// TestHttpViaIP 通过指定的IP地址对给定的URL进行HTTP请求测试。
//
// 参数:
//
//	t *testing.T - 测试框架提供的测试上下文，用于报告测试失败和日志记录。
//
// 无返回值。
func TestHttpViaIP(t *testing.T) {
	ip := "2606:2800:220:1:248:1893:25c8:1946" // 要访问的目标IP地址
	url := "http://www.example.com/"           // 要访问的目标URL

	// 使用指定的IP地址发起HTTP GET请求
	resp, err := FetchWithIP(ip, url)
	if err != nil {
		fmt.Println("Error:", err)
		t.Errorf(err.Error())
		return
	}

	// 确保响应体在函数返回前被关闭
	defer resp.Body.Close()
	// 读取并打印响应体内容
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Response:", string(body))
}

// TestHttpsViaIP 通过指定的IP地址测试HTTPS连接。
// 参数:
//
//	t *testing.T - 测试框架提供的测试上下文，用于报告测试失败和日志记录。
//
// 无返回值。
func TestHttpsViaIP(t *testing.T) {
	ip := "2606:2800:220:1:248:1893:25c8:1946" // 要访问的IP地址
	url := "https://www.example.com/"          // 要访问的URL

	resp, err := FetchWithIP(ip, url)
	if err != nil {
		fmt.Println("Error:", err)
		t.Errorf(err.Error())
		return
	}

	defer resp.Body.Close() // 确保在函数返回前关闭响应体
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Response:", string(body)) // 打印响应内容
}
func FetchWithIP(ip, url string) (*http.Response, error) {
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

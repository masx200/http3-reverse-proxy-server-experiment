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

	//var // The `addresses` variable in the code is a slice containing two IP addresses: "93.184.216.34"
	// and "2606:2800:220:1:248:1893:25c8:1946". These IP addresses are used for testing HTTP and
	// HTTPS connections to a specified URL in the `TestHttpViaIP` and `TestHttpsViaIP` functions. The
	// code iterates over these addresses and performs HTTP requests using each IP address.
	var addresses = []string{"93.184.216.34", "2606:2800:220:1:248:1893:25c8:1946"}
	var eee error = nil
	for _, address := range addresses {
		ip := address                    // 要访问的目标IP地址
		url := "http://www.example.com/" // 要访问的目标URL

		// 使用指定的IP地址发起HTTP GET请求
		resp, err := FetchWithIP(ip, url)
		if err != nil {
			fmt.Println("Error:", err)
			// t.Errorf(err.Error())
			eee = err
			continue
		}

		// 确保响应体在函数返回前被关闭
		defer resp.Body.Close()
		// 读取并打印响应体内容
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Response:", string(body))
		return
	}
	if eee != nil {
		fmt.Println("Error:", eee)
		t.Errorf(eee.Error())
	}
}

// TestHttpsViaIP 通过指定的IP地址测试HTTPS连接。
// 参数:
//
//	t *testing.T - 测试框架提供的测试上下文，用于报告测试失败和日志记录。
//
// 无返回值。
func TestHttpsViaIP(t *testing.T) {
	var addresses = []string{"93.184.216.34", "2606:2800:220:1:248:1893:25c8:1946"}
	var eee error = nil
	for _, address := range addresses {
		ip := address                     // 要访问的目标IP地址
		url := "https://www.example.com/" // 要访问的目标URL

		// 使用指定的IP地址发起HTTP GET请求
		resp, err := FetchWithIP(ip, url)
		if err != nil {
			fmt.Println("Error:", err)
			// t.Errorf(err.Error())
			eee = err
			continue
		}

		// 确保响应体在函数返回前被关闭
		defer resp.Body.Close()
		// 读取并打印响应体内容
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Response:", string(body))
		return
	}
	if eee != nil {
		fmt.Println("Error:", eee)
		t.Errorf(eee.Error())
	}
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

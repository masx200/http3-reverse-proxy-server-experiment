package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/http2"
	// "golang.org/x/net/http2/h2c"
)

func main() {
	client := http.Client{

		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}

	resp, err := client.Get("http://localhost:8080")
	if err != nil {
		log.Fatalf("请求失败: %s", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatalf("读取响应失败: %s", err)
	}

	fmt.Printf("获取响应 %d: %s\n", resp.StatusCode, string(body))

}

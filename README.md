# http3-reverse-proxy-server-experiment

#### 介绍

http3反向代理服务器实验golang

#### 软件架构

软件架构说明

实现了随机选择的负载均衡算法和服务器被动健康检查功能

可以设定服务器被动健康检查的间隔时间和健康检查的方法和路径和状态码

本地服务器和上游服务器都支持http3和http2和http1.1协议

支持了多个上游服务器,本地服务器和上游服务器都支持http和https协议

添加了防环检测功能和被动健康检查的限速功能

添加了dns over https/dns over http3/dns over quic/dns over tls客户端的功能

增加了通过dns的https记录查询服务器支持http3的功能

增加了通过http2响应头alt-svc查询支持http3的功能

添加了通过自定义的ip地址访问http1/http2/http3的功能

添加了主动健康检查功能和dns over https负载均衡功能

实现了通过A,AAAA,HTTPS,CNAME记录解析域名的功能

#### 安装教程

```
go build main.go
```

#### 使用说明

```
Usage of main.exe:
  -http-port int
        http-port (default 18080)
  -https-port int
        https-port (default 18443)
  -listen-h2c
        listen-h2c (default true)
  -listen-hostname string
        listen-hostname (default "0.0.0.0")
  -listen-http
        listen-http (default true)
  -listen-http3
        listen-http3 (default true)
  -listen-tls
        listen-tls (default true)
  -tls-cert string
        tls-cert (default "cert.crt")
  -tls-key string
        tls-key (default "key.pem")
  -upstream-protocol string
        upstream-protocol (default "h3")
  -upstream-server string
        upstream-server (default "https://workers.cloudflare.com/")
```

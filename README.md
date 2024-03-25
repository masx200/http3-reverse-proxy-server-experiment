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

添加了dns over https客户端的功能

增加了通过dns的https记录查询服务器支持http3的功能

增加了通过http2响应头alt-svc查询支持http3的功能

添加了通过自定义的ip地址访问http1/http2/http3的功能

#### 安装教程

1. xxxx
2. xxxx
3. xxxx

#### 使用说明

1. xxxx
2. xxxx
3. xxxx

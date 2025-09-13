# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with
code in this repository.

## Build and Development Commands

### Build

```bash
go build main.go
```

### Run Tests

```bash
go test -v ./...
```

### Build and Test (as per CI/CD)

```bash
go build -v ./...
go test -v ./...
```

### Debug Pprof

Run with debug pprof enabled:

```bash
go run main.go -debug-pprof
```

## Architecture Overview

This is an HTTP/3 reverse proxy server experiment implemented in Go. The system
supports multiple protocols and advanced load balancing capabilities.

### Core Components

1. **Main Server** (`main.go`)
   - Entry point that configures and starts multiple HTTP servers
   - Supports HTTP/1.1, HTTP/2, HTTP/3 protocols simultaneously
   - Includes loop detection middleware
   - Supports debug profiling with pprof

2. **HTTP/3 Transport** (`h3/h3.go`)
   - Custom HTTP/3 implementation using QUIC protocol
   - Supports custom IP binding for network flexibility
   - Includes connection pooling and rotation to avoid rate limiting

3. **Load Balancer** (`load_balance/`)
   - Interface-based load balancing system
   - Supports active and passive health checks
   - Random load balancing algorithm with failover strategies
   - Server configuration management

4. **DNS Integration** (`dns/`)
   - DNS over HTTPS (DoH) support
   - DNS over QUIC (DoQ) support
   - DNS over TLS (DoT) support
   - A, AAAA, HTTPS, CNAME record resolution

5. **Protocol Adapters** (`adapter/`)
   - Custom HTTP round tripper implementations
   - Protocol abstraction layer
   - Connection management interfaces

### Key Features

- **Multi-Protocol Support**: HTTP/1.1, HTTP/2, HTTP/3 both client and server
  side
- **Load Balancing**: Random algorithm with configurable health checks
- **DNS Resolution**: Advanced DNS query capabilities over multiple protocols
- **Loop Detection**: Prevention of infinite proxy loops
- **Custom Transport**: Flexible network configuration with IP binding
- **Health Checks**: Active and passive server health monitoring
- **Debug Tools**: Built-in pprof profiling support

### Project Structure

```
├── main.go                    # Main server entry point
├── adapter/                  # Protocol adapter layer
├── h3/                       # HTTP/3 implementation
├── load_balance/             # Load balancing system
├── dns/                      # DNS resolution services
├── resolver/                 # DNS resolvers
├── print/                    # Debug printing utilities
├── h12/                      # HTTP1.2 experiments
├── h2c_client/               # HTTP/2 cleartext client
├── h2c_server/               # HTTP/2 cleartext server
├── doh_debugger/             # DoH debugging tool
└── doh3_debugger/           # DoH3 debugging tool
```

### Dependencies

- `github.com/quic-go/quic-go`: HTTP/3 and QUIC protocol implementation
- `github.com/miekg/dns`: DNS library for various DNS operations
- `github.com/gin-gonic/gin`: HTTP web framework
- `github.com/moznion/go-optional`: Optional type support
- `golang.org/x/net`: Extended networking libraries

### Testing

Run individual test files:

```bash
go test -v ./h3/
go test -v ./dns/
go test -v ./load_balance/
```

### Configuration

The server accepts command-line flags for configuration:

- `--upstream-server`: Target upstream server URL
- `--http-port`: HTTP server port (default: 18080)
- `--https-port`: HTTPS server port (default: 18443)
- `--upstream-protocol`: Protocol for upstream connections (h3, h2, h2c,
  http/1.1)
- `--listen-http`, `--listen-h2c`, `--listen-http3`: Enable/disable protocols
- `--tls-cert`, `--tls-key`: TLS certificate files
- `--debug-pprof`: Enable debug profiling

### Development Notes

- The system uses Gin framework for HTTP routing
- Connection pooling is implemented to avoid UDP rate limiting
- HTTP/3 transport rotates connections periodically
- Loop detection prevents infinite proxy chains
- Health checks are configurable for both active and passive monitoring

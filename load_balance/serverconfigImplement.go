package load_balance

import (
	"net/http"
	"sync"
	// "time"
)

type ServerConfigImplement struct {
	Identifier              string
	HealthMutex             sync.Mutex
	UpstreamServerURL       string
	IsHealthy               bool
	HealthCheckIntervalMs   int64
	unHealthyFailDurationMs int64
	unHealthyFailMaxCount   int64
	ActiveHealthyChecker    func(RoundTripper http.RoundTripper, url string) (bool, error) // 活跃健康检查函数，用于检查给定的传输和URL是否健康。
	RoundTripper            http.RoundTripper
	PassiveUnHealthyChecker func(response *http.Response) (bool, error) // 健康响应检查函数，用于基于HTTP响应检查客户端的健康状态。
}

func (s *ServerConfigImplement) GetUpStreamServerURL() string {
	return s.UpstreamServerURL
}

func (s *ServerConfigImplement) ActiveHealthyCheck() (bool, error) {
	// 这里应实现主动健康检查的逻辑
	// 示例仅返回健康状态和空错误
	return s.ActiveHealthyChecker(s.RoundTripper, s.UpstreamServerURL)
}

func (s *ServerConfigImplement) GetIdentifier() string {
	// 示例中简单返回URL作为唯一标识符，实际情况可能需要更复杂逻辑
	return s.Identifier
}

func (s *ServerConfigImplement) GetHealthy() bool {
	s.HealthMutex.Lock()
	defer s.HealthMutex.Unlock()
	return s.IsHealthy
}

func (s *ServerConfigImplement) SetHealthy(healthStatus bool) {
	s.HealthMutex.Lock()
	defer s.HealthMutex.Unlock()
	s.IsHealthy = healthStatus
}

func (s *ServerConfigImplement) PassiveUnHealthyCheck(response *http.Response) (bool, error) {
	// 这里应根据HTTP响应判断服务是否健康
	// 示例仅返回一个假设的结果和空错误
	// 实际情况可能基于HTTP状态码、响应体或其他响应数据来判断
	return s.PassiveUnHealthyChecker(response)
}

func (s *ServerConfigImplement) SetHealthyCheckInterval(intervalMs int64) {
	s.HealthCheckIntervalMs = intervalMs
}

func (s *ServerConfigImplement) GetHealthyCheckInterval() int64 {
	return s.HealthCheckIntervalMs
}

func (s *ServerConfigImplement) SetUnHealthyFailDurationMs(durationMs int64) {
	s.unHealthyFailDurationMs = durationMs
}

func (s *ServerConfigImplement) GetUnHealthyFailDurationMs() int64 {
	return int64(s.unHealthyFailDurationMs)
}

func (s *ServerConfigImplement) SetUnHealthyFailMaxCount(count int64) {
	s.unHealthyFailMaxCount = count
}

func (s *ServerConfigImplement) GetUnHealthyFailMaxCount() int64 {
	return s.unHealthyFailMaxCount
}

package load_balance

import (
	"fmt"
	"net/http"
	"sync"
	"time"
	// "time"
)

type ServerConfigImplement struct {
	Identifier              string
	HealthMutex             sync.Mutex
	FailureMutex            sync.Mutex
	UnHealthMutex           sync.Mutex
	UpstreamServerURL       string
	IsHealthy               bool
	HealthCheckIntervalMs   int64
	unHealthyFailDurationMs int64
	unHealthyFailMaxCount   int64
	ActiveHealthyChecker    func(RoundTripper http.RoundTripper, url string) (bool, error) // 活跃健康检查函数，用于检查给定的传输和URL是否健康。
	RoundTripper            http.RoundTripper
	PassiveUnHealthyChecker func(response *http.Response) (bool, error) // 健康响应检查函数，用于基于HTTP响应检查客户端的健康状态。
	UnHealthyFailCount      int64
}

// OnUpstreamFailure implements ServerConfigCommon.
func (s *ServerConfigImplement) OnUpstreamFailure() {

	fmt.Println("OnUpstreamFailure", s.GetIdentifier())
	s.FailureMutex.Lock()
	defer s.FailureMutex.Unlock()
	s.IncrementUnHealthyFailCount()
	failDuration := s.GetUnHealthyFailDurationMs() * int64(time.Millisecond)
	failCount := s.GetUnHealthyFailCount()

	// 如果unHealthyFailDurationMs大于0并且当前失败次数+1超过unHealthyFailMaxCount，则标记上游服务为不健康
	if s.unHealthyFailDurationMs > 0 && failCount >= s.GetUnHealthyFailMaxCount() {
		s.SetHealthy(false)
	} else if failCount == 1 { // 第一次失败时，启动计时器
		go func() {
			time.Sleep(time.Duration(failDuration))
			if s.GetUnHealthyFailCount() >= 1 { // 计时结束后，如果期间没有其他失败请求，则重置失败次数
				s.ResetUnHealthyFailCount()
			}
		}()
	}

}

// ServerConfigImplementConstructor 是用于构造ServerConfigImplement对象的函数。
// Identifier: 用于标识服务器的唯一字符串。
// UpStreamServerURL: 指定上游服务器的URL。
// RoundTripper: 实现http.RoundTripper接口的对象，用于进行HTTP请求。
// option: 可选参数，一系列函数，可用于修改ServerConfigImplement对象的配置。
// 返回值为配置好的ServerConfigCommon接口，实现了服务器配置的通用接口。
func ServerConfigImplementConstructor(Identifier string, UpStreamServerURL string, RoundTripper http.RoundTripper, option ...func(*ServerConfigImplement)) ServerConfigCommon {
	var s = &ServerConfigImplement{
		Identifier:              Identifier,
		HealthCheckIntervalMs:   HealthCheckIntervalMsDefault,
		UpstreamServerURL:       UpStreamServerURL,
		IsHealthy:               true,
		unHealthyFailDurationMs: unHealthyFailDurationMsDefault,
		unHealthyFailMaxCount:   UnHealthyFailMaxCountDefault,
		ActiveHealthyChecker:    ActiveHealthyCheckDefault,
		RoundTripper:            RoundTripper,
		PassiveUnHealthyChecker: HealthyResponseCheckDefault}
	for _, callback := range option {
		callback(s)

	}
	return s
}

// GetUnHealthyFailCount implements ServerConfigCommon.
func (s *ServerConfigImplement) GetUnHealthyFailCount() int64 {
	s.UnHealthMutex.Lock()
	defer s.UnHealthMutex.Unlock()
	return s.UnHealthyFailCount
}

// IncrementUnHealthyFailCount implements ServerConfigCommon.
func (s *ServerConfigImplement) IncrementUnHealthyFailCount() {
	s.UnHealthMutex.Lock()
	defer s.UnHealthMutex.Unlock()
	s.UnHealthyFailCount += 1
}

// ResetUnHealthyFailCount implements ServerConfigCommon.
func (s *ServerConfigImplement) ResetUnHealthyFailCount() {
	s.UnHealthMutex.Lock()
	defer s.UnHealthMutex.Unlock()
	s.UnHealthyFailCount = 0
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

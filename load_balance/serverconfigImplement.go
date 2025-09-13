package load_balance

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/masx200/http3-reverse-proxy-server-experiment/generic"
	// "time"
)

const PassiveUnHealthyCheckStatusCodeRangeDefaultStart = 500
const PassiveUnHealthyCheckStatusCodeRangeDefaultEnd = 600

type ServerConfigImplement struct {
	PassiveHealthyCheckEnabled        bool
	ActiveHealthyCheckEnabled         bool
	ActiveHealthyCheckStatusCodeRange generic.PairInterface[int, int]
	ActiveHealthyCheckMethod          string
	Identifier                        string
	HealthMutex                       sync.Mutex
	FailureMutex                      sync.Mutex
	FailCountMutex                    sync.Mutex
	UpStreamServerURL                 string
	IsHealthy                         bool
	HealthCheckIntervalMs             int64
	unHealthyFailDurationMs           int64
	unHealthyFailMaxCount             int64
	ActiveHealthyChecker              func(RoundTripper http.RoundTripper, url string, method string, statusCodeMin int, statusCodeMax int) (bool, error) // 活跃健康检查函数，用于检查给定的传输和URL是否健康。
	RoundTripper                      http.RoundTripper
	PassiveUnHealthyChecker           func(response *http.Response, UnHealthyStatusMin int, UnHealthyStatusMax int) (bool, error) // 健康响应检查函数，用于基于HTTP响应检查客户端的健康状态。
	UnHealthyFailCount                int64
	ActiveHealthyCheckURL             string

	PassiveUnHealthyCheckStatusCodeRange generic.PairInterface[int, int]
}

// GetActiveHealthyCheckEnabled implements ServerConfigCommon.
func (s *ServerConfigImplement) GetActiveHealthyCheckEnabled() bool {
	return s.ActiveHealthyCheckEnabled
}

// GetPassiveHealthyCheckEnabled implements ServerConfigCommon.
func (s *ServerConfigImplement) GetPassiveHealthyCheckEnabled() bool {
	return s.PassiveHealthyCheckEnabled
}

// OnUpstreamHealthy implements ServerConfigCommon.
func (s *ServerConfigImplement) OnUpstreamHealthy() {
	log.Println("OnUpstreamHealthy", s.GetIdentifier())
}

// SetActiveHealthyCheckEnabled implements ServerConfigCommon.
func (s *ServerConfigImplement) SetActiveHealthyCheckEnabled(e bool) {

	s.ActiveHealthyCheckEnabled = e
}

// SetPassiveHealthyCheckEnabled implements ServerConfigCommon.
func (s *ServerConfigImplement) SetPassiveHealthyCheckEnabled(e bool) {
	s.PassiveHealthyCheckEnabled = e
}

// GetPassiveUnHealthyCheckStatusCodeRange implements ServerConfigCommon.
func (s *ServerConfigImplement) GetPassiveUnHealthyCheckStatusCodeRange() generic.PairInterface[int, int] {
	return s.PassiveUnHealthyCheckStatusCodeRange
}

// GetActiveHealthyCheckMethod implements ServerConfigCommon.
func (s *ServerConfigImplement) GetActiveHealthyCheckMethod() string {
	return s.ActiveHealthyCheckMethod
}

// GetActiveHealthyCheckStatusCodeRange implements ServerConfigCommon.
func (s *ServerConfigImplement) GetActiveHealthyCheckStatusCodeRange() generic.PairInterface[int, int] {
	return s.ActiveHealthyCheckStatusCodeRange
}

// GetActiveHealthyCheckURL implements ServerConfigCommon.
func (s *ServerConfigImplement) GetActiveHealthyCheckURL() string {
	return s.ActiveHealthyCheckURL
}

// OnUpstreamFailure implements ServerConfigCommon.
func (s *ServerConfigImplement) OnUpstreamFailure() {

	log.Println("OnUpstreamFailure", s.GetIdentifier())

	if !s.PassiveHealthyCheckEnabled {
		return

	}
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

const ActiveHealthyCheckStatusCodeRangeDefaultStart = 200
const ActiveHealthyCheckStatusCodeRangeDefaultEnd = 300
const ActiveHealthyCheckMethodDefault = "HEAD"

// ServerConfigImplementConstructor 是用于构造ServerConfigImplement对象的函数。
// Identifier: 用于标识服务器的唯一字符串。
// UpStreamServerURL: 指定上游服务器的URL。
// RoundTripper: 实现http.RoundTripper接口的对象，用于进行HTTP请求。
// option: 可选参数，一系列函数，可用于修改ServerConfigImplement对象的配置。
// 返回值为配置好的ServerConfigCommon接口，实现了服务器配置的通用接口。
func ServerConfigImplementConstructor(Identifier string, UpStreamServerURL string, RoundTripper http.RoundTripper, option ...func(*ServerConfigImplement)) ServerConfigCommon {

	var s = &ServerConfigImplement{
		ActiveHealthyCheckStatusCodeRange:    generic.NewPairImplement(ActiveHealthyCheckStatusCodeRangeDefaultStart, ActiveHealthyCheckStatusCodeRangeDefaultEnd),
		ActiveHealthyCheckMethod:             ActiveHealthyCheckMethodDefault,
		Identifier:                           Identifier,
		HealthMutex:                          sync.Mutex{},
		FailureMutex:                         sync.Mutex{},
		FailCountMutex:                       sync.Mutex{},
		UpStreamServerURL:                    UpStreamServerURL,
		IsHealthy:                            true,
		HealthCheckIntervalMs:                HealthCheckIntervalMsDefault,
		unHealthyFailDurationMs:              unHealthyFailDurationMsDefault,
		unHealthyFailMaxCount:                UnHealthyFailMaxCountDefault,
		ActiveHealthyChecker:                 ActiveHealthyCheckDefault,
		RoundTripper:                         RoundTripper,
		PassiveUnHealthyChecker:              HealthyResponseCheckDefault,
		UnHealthyFailCount:                   0,
		ActiveHealthyCheckURL:                UpStreamServerURL,
		PassiveUnHealthyCheckStatusCodeRange: generic.NewPairImplement(PassiveUnHealthyCheckStatusCodeRangeDefaultStart, PassiveUnHealthyCheckStatusCodeRangeDefaultEnd),
	}
	for _, callback := range option {
		callback(s)

	}
	return s
}

// GetUnHealthyFailCount implements ServerConfigCommon.
func (s *ServerConfigImplement) GetUnHealthyFailCount() int64 {
	s.FailCountMutex.Lock()
	defer s.FailCountMutex.Unlock()
	return s.UnHealthyFailCount
}

// IncrementUnHealthyFailCount implements ServerConfigCommon.
func (s *ServerConfigImplement) IncrementUnHealthyFailCount() {
	s.FailCountMutex.Lock()
	defer s.FailCountMutex.Unlock()
	s.UnHealthyFailCount += 1
}

// ResetUnHealthyFailCount implements ServerConfigCommon.
func (s *ServerConfigImplement) ResetUnHealthyFailCount() {
	s.FailCountMutex.Lock()
	defer s.FailCountMutex.Unlock()
	s.UnHealthyFailCount = 0
}

func (s *ServerConfigImplement) GetUpStreamServerURL() string {
	return s.UpStreamServerURL
}

func (l *ServerConfigImplement) ActiveHealthyCheck() (bool, error) {

	x, x1 := l.ActiveHealthyChecker(l.RoundTripper, l.GetActiveHealthyCheckURL(), l.GetActiveHealthyCheckMethod(), l.GetActiveHealthyCheckStatusCodeRange().GetFirst(), l.GetActiveHealthyCheckStatusCodeRange().GetSecond())
	if x {
		l.OnUpstreamHealthy()
	}
	return x, x1
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
	return s.PassiveUnHealthyChecker(response, s.PassiveUnHealthyCheckStatusCodeRange.GetFirst(), s.PassiveUnHealthyCheckStatusCodeRange.GetSecond())
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

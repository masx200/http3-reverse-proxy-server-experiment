package load_balance

import (
	"net/http"

	"github.com/masx200/http3-reverse-proxy-server-experiment/generic"
)

type LoadBalanceService interface {
	GetUpStreams() generic.MapInterface[string, LoadBalanceAndUpStream]

	//选择一个可用的上游服务器
	// 参数：
	SelectAvailableServers() ([]LoadBalanceAndUpStream, error)

	LoadBalancePolicySelector() ([]LoadBalanceAndUpStream, error)

	GetActiveHealthyCheckEnabled() bool
	SetActiveHealthyCheckEnabled(bool)
	GetPassiveHealthyCheckEnabled() bool
	SetPassiveHealthyCheckEnabled(bool)
	HealthyCheckStart()
	HealthyCheckRunning() bool
	HealthyCheckStop()
	Close() error
	GetIdentifier() string
	RoundTrip(*http.Request) (*http.Response, error)

	FailoverAttemptStrategy(*http.Request) bool
}

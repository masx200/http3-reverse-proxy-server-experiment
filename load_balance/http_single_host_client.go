package load_balance

import (
	"net/http"
	"net/url"

	optional "github.com/moznion/go-optional"
)

func ActiveHealthyCheckDefault(RoundTripper http.RoundTripper) (bool, error) {
	return true, nil
}
func NewLoadBalanceSingleHostHTTPClient(identifier string, UpStreamServer url.URL) LoadBalanceAndUpStream {
	return &LoadBalanceSingleHostHTTPClient{
		Identifier:               identifier,
		ActiveHealthyChecker:     ActiveHealthyCheckDefault,
		IsHealthyResponseChecker: IsHealthyResponseDefault,
	}
}

type LoadBalanceSingleHostHTTPClient struct {
	ActiveHealthyChecker     func(RoundTripper http.RoundTripper) (bool, error)
	Identifier               string
	isHealthy                bool
	IsHealthyResponseChecker func(response *http.Response) (bool, error)
	RoundTripper             http.RoundTripper
}

// ActiveHealthyCheck implements LoadBalanceAndUpStream.
func (l *LoadBalanceSingleHostHTTPClient) ActiveHealthyCheck() (bool, error) {
	return l.ActiveHealthyChecker(l)
}

// Identifier implements LoadBalanceAndUpStream.
func (l *LoadBalanceSingleHostHTTPClient) GetIdentifier() string {
	return l.Identifier
}

// IsHealthy implements LoadBalanceAndUpStream.
func (l *LoadBalanceSingleHostHTTPClient) IsHealthy() bool {
	return l.isHealthy
}

// IsHealthyResponse implements LoadBalanceAndUpStream.
func (l *LoadBalanceSingleHostHTTPClient) IsHealthyResponse(response *http.Response) (bool, error) {
	return l.IsHealthyResponseChecker(response)
}
func IsHealthyResponseDefault(response *http.Response) (bool, error) {
	return true, nil
}

// RoundTrip implements LoadBalanceAndUpStream.
func (l *LoadBalanceSingleHostHTTPClient) RoundTrip(request *http.Request) (*http.Response, error) {
	return l.RoundTripper.RoundTrip(request)
}

// SelectAvailableServer implements LoadBalanceAndUpStream.
func (l *LoadBalanceSingleHostHTTPClient) SelectAvailableServer() (LoadBalanceAndUpStream, error) {
	return l, nil
}

// SetHealthy implements LoadBalanceAndUpStream.
func (l *LoadBalanceSingleHostHTTPClient) SetHealthy(healthy bool) {
	l.isHealthy = healthy
}

// UpStreams implements LoadBalanceAndUpStream.
func (l *LoadBalanceSingleHostHTTPClient) UpStreams() optional.Option[MapInterface[string, LoadBalanceAndUpStream]] {
	return optional.None[MapInterface[string, LoadBalanceAndUpStream]]()
}

package load_balance

import (
	optional "github.com/moznion/go-optional"
	"net/http"
)

func NewLoadBalanceHTTPSingleHostClient() LoadBalanceAndUpStream {
	return &LoadBalanceHTTPSingleHostClient{}
}

type LoadBalanceHTTPSingleHostClient struct {
}

// ActiveHealthyCheck implements LoadBalanceAndUpStream.
func (l *LoadBalanceHTTPSingleHostClient) ActiveHealthyCheck() (bool, error) {
	panic("unimplemented")
}

// Identifier implements LoadBalanceAndUpStream.
func (l *LoadBalanceHTTPSingleHostClient) Identifier() string {
	panic("unimplemented")
}

// IsHealthy implements LoadBalanceAndUpStream.
func (l *LoadBalanceHTTPSingleHostClient) IsHealthy() bool {
	panic("unimplemented")
}

// IsHealthyResponse implements LoadBalanceAndUpStream.
func (l *LoadBalanceHTTPSingleHostClient) IsHealthyResponse(*http.Response) (bool, error) {
	panic("unimplemented")
}

// RoundTrip implements LoadBalanceAndUpStream.
func (l *LoadBalanceHTTPSingleHostClient) RoundTrip(*http.Request) (*http.Response, error) {
	panic("unimplemented")
}

// SelectAvailableServer implements LoadBalanceAndUpStream.
func (l *LoadBalanceHTTPSingleHostClient) SelectAvailableServer() (LoadBalanceAndUpStream, error) {
	panic("unimplemented")
}

// SetHealthy implements LoadBalanceAndUpStream.
func (l *LoadBalanceHTTPSingleHostClient) SetHealthy(bool) {
	panic("unimplemented")
}

// UpStreams implements LoadBalanceAndUpStream.
func (l *LoadBalanceHTTPSingleHostClient) UpStreams() optional.Option[MapInterface[string, LoadBalanceAndUpStream]] {
	panic("unimplemented")
}

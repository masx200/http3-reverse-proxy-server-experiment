package load_balance

import (
	"fmt"
	"net/http"
	// "net/url"

	optional "github.com/moznion/go-optional"
)

func ActiveHealthyCheckDefault(RoundTripper http.RoundTripper) (bool, error) {
	return true, nil
}
func NewSingleHostHTTPClientOfAddress(identifier string, UpStreamServerURL string, address string) LoadBalanceAndUpStream {
	return &SingleHostHTTPClientOfAddress{
		Identifier:               identifier,
		ActiveHealthyChecker:     ActiveHealthyCheckDefault,
		IsHealthyResponseChecker: IsHealthyResponseDefault,
	}
}

type SingleHostHTTPClientOfAddress struct {
	ActiveHealthyChecker     func(RoundTripper http.RoundTripper) (bool, error)
	Identifier               string
	isHealthy                bool
	IsHealthyResponseChecker func(response *http.Response) (bool, error)
	RoundTripper             http.RoundTripper
}

// ActiveHealthyCheck implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) ActiveHealthyCheck() (bool, error) {
	return l.ActiveHealthyChecker(l)
}

// Identifier implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) GetIdentifier() string {
	return l.Identifier
}

// IsHealthy implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) IsHealthy() bool {
	return l.isHealthy
}

// IsHealthyResponse implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) IsHealthyResponse(response *http.Response) (bool, error) {
	return l.IsHealthyResponseChecker(response)
}
func IsHealthyResponseDefault(response *http.Response) (bool, error) {

	//check StatusCode<500
	if response.StatusCode >= 500 {
		return false, fmt.Errorf("StatusCode %d   is greater than 500", response.StatusCode)
	}
	return true, nil
}

// RoundTrip implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) RoundTrip(request *http.Request) (*http.Response, error) {
	return l.RoundTripper.RoundTrip(request)
}

// SelectAvailableServer implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) SelectAvailableServer() (LoadBalanceAndUpStream, error) {
	return l, nil
}

// SetHealthy implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) SetHealthy(healthy bool) {
	l.isHealthy = healthy
}

// UpStreams implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) UpStreams() optional.Option[MapInterface[string, LoadBalanceAndUpStream]] {
	return optional.None[MapInterface[string, LoadBalanceAndUpStream]]()
}

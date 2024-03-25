package load_balance

import (
	"fmt"
	"net/http"
	// "net/url"

	print_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/print"
	optional "github.com/moznion/go-optional"
)

func PrintResponse(resp *http.Response) {
	print_experiment.PrintResponse(resp)
}

func ActiveHealthyCheckDefault(RoundTripper http.RoundTripper, url string) (bool, error) {

	client := &http.Client{}

	client.Transport = RoundTripper
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false, err
	}
	PrintRequest(req)
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	PrintResponse(resp)
	return IsHealthyResponseDefault(resp)
	//return true, nil
}

func PrintRequest(req *http.Request) {
	print_experiment.PrintRequest(req)
}
func NewSingleHostHTTPClientOfAddress(identifier string, UpStreamServerURL string, address string) LoadBalanceAndUpStream {
	return &SingleHostHTTPClientOfAddress{
		Identifier:               identifier,
		ActiveHealthyChecker:     ActiveHealthyCheckDefault,
		IsHealthyResponseChecker: IsHealthyResponseDefault,
		UpStreamServerURL:        UpStreamServerURL,
	}
}

type SingleHostHTTPClientOfAddress struct {
	ActiveHealthyChecker     func(RoundTripper http.RoundTripper, url string) (bool, error)
	Identifier               string
	isHealthy                bool
	IsHealthyResponseChecker func(response *http.Response) (bool, error)
	RoundTripper             http.RoundTripper
	UpStreamServerURL        string
}

// ActiveHealthyCheck implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) ActiveHealthyCheck() (bool, error) {
	return l.ActiveHealthyChecker(l, l.UpStreamServerURL)
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

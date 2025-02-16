package probe

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/glendsoza/sprobe/status"
)

type HttpProbe interface {
	Probe(req *http.Request, timeout time.Duration) (status.Status, string, error)
}

type httpProbe struct {
	transport               *http.Transport
	followNonLocalRedirects bool
}

func NewHttpProbe(followNonLocalRedirects bool) *httpProbe {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	return NewWithTLSConfig(tlsConfig, followNonLocalRedirects)
}

func NewWithTLSConfig(config *tls.Config, followNonLocalRedirects bool) *httpProbe {

	transport :=
		&http.Transport{
			TLSClientConfig:    config,
			DisableKeepAlives:  true,
			DisableCompression: true,
		}

	return &httpProbe{transport, followNonLocalRedirects}
}

func (pr *httpProbe) Probe(req *http.Request, timeout time.Duration) (status.Status, string, error) {
	client := &http.Client{
		Timeout:       timeout,
		Transport:     pr.transport,
		CheckRedirect: RedirectChecker(pr.followNonLocalRedirects),
	}
	return DoHTTPProbe(req, client)
}

type GetHTTPInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

func DoHTTPProbe(req *http.Request, client GetHTTPInterface) (status.Status, string, error) {
	res, err := client.Do(req)
	if err != nil {
		return status.Failure, err.Error(), nil
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return status.Failure, "", err
	}
	body := string(b)
	if res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusBadRequest {
		if res.StatusCode >= http.StatusMultipleChoices {
			return status.Warning, fmt.Sprintf("Probe terminated redirects, Response body: %v", body), nil
		}
		return status.Success, body, nil
	}

	failureMsg := fmt.Sprintf("HTTP probe failed with statuscode: %d", res.StatusCode)
	return status.Failure, failureMsg, nil
}

func RedirectChecker(followNonLocalRedirects bool) func(*http.Request, []*http.Request) error {
	if followNonLocalRedirects {
		return nil
	}

	return func(req *http.Request, via []*http.Request) error {
		if req.URL.Hostname() != via[0].URL.Hostname() {
			return http.ErrUseLastResponse
		}

		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}
}

package httputil

import (
	"errors"
	"net/http"
	"net/url"
)

func NewClient(proxy string, insecureSkipVerify bool, caCertFiles []string) (http.Client, error) {
	transport, err := NewTransport(proxy, insecureSkipVerify, caCertFiles)
	if err != nil {
		return http.Client{}, err
	}
	return http.Client{
		Transport: transport,
	}, nil
}

func NewTransport(proxy string, insecureSkipVerify bool, caCertFiles []string) (*http.Transport, error) {
	tlsConfig, err := LoadTLSConfig(insecureSkipVerify, caCertFiles)
	if err != nil {
		return nil, err
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig

	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		if proxyURL.Scheme == "" || proxyURL.Host == "" {
			return nil, errors.New("proxy URL must include scheme and host")
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	return transport, nil
}

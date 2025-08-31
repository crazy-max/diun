package utl

import (
	"crypto/tls"
	"crypto/x509"
	"os"
)

func LoadTLSConfig(insecureSkipVerify bool, caCertFiles []string) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}
	if len(caCertFiles) > 0 {
		certPool := x509.NewCertPool()
		for _, caCertFile := range caCertFiles {
			caCert, err := os.ReadFile(caCertFile)
			if err != nil {
				return nil, err
			}
			certPool.AppendCertsFromPEM(caCert)
		}
		tlsConfig.RootCAs = certPool
	}
	return tlsConfig, nil
}

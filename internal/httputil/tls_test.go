package httputil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadTLSConfig(t *testing.T) {
	t.Run("without CA cert files", func(t *testing.T) {
		tlsConfig, err := LoadTLSConfig(true, nil)
		require.NoError(t, err)
		require.True(t, tlsConfig.InsecureSkipVerify)
		require.Nil(t, tlsConfig.RootCAs)
	})

	t.Run("with CA cert file", func(t *testing.T) {
		cert, certPEM := newTestCACert(t)
		certFile := filepath.Join(t.TempDir(), "ca.pem")
		require.NoError(t, os.WriteFile(certFile, certPEM, 0o644))

		tlsConfig, err := LoadTLSConfig(false, []string{certFile})
		require.NoError(t, err)
		require.False(t, tlsConfig.InsecureSkipVerify)
		require.NotNil(t, tlsConfig.RootCAs)

		_, err = cert.Verify(x509.VerifyOptions{
			Roots: tlsConfig.RootCAs,
		})
		require.NoError(t, err)
	})

	t.Run("missing CA cert file", func(t *testing.T) {
		_, err := LoadTLSConfig(false, []string{filepath.Join(t.TempDir(), "missing.pem")})
		require.Error(t, err)
	})
}

func newTestCACert(t *testing.T) (*x509.Certificate, []byte) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "diun-test-ca",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	return cert, pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: der,
	})
}

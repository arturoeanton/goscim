package scim

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Certificate verification was disabled unconditionally, which spent the cost
// of TLS without buying its protection. It is now on unless asked otherwise.
func TestCouchbaseConnectionVerifiesByDefault(t *testing.T) {
	t.Setenv("SCIM_COUCHBASE_TLS", "")
	t.Setenv("SCIM_COUCHBASE_TLS_SKIP_VERIFY", "")
	t.Setenv("SCIM_COUCHBASE_CA_CERT", "")

	connection, security, err := couchbaseConnection("db.example")
	if err != nil {
		t.Fatal(err)
	}
	if connection != "couchbases://db.example" {
		t.Errorf("connection string = %s", connection)
	}
	if security.TLSSkipVerify {
		t.Error("verification is disabled by default")
	}
	if security.TLSRootCAs != nil {
		t.Error("no CA was configured, so the system pool should be used")
	}
}

func TestCouchbaseConnectionOptions(t *testing.T) {
	t.Run("verification can be turned off deliberately", func(t *testing.T) {
		t.Setenv("SCIM_COUCHBASE_TLS_SKIP_VERIFY", "true")
		_, security, err := couchbaseConnection("db.example")
		if err != nil {
			t.Fatal(err)
		}
		if !security.TLSSkipVerify {
			t.Error("SCIM_COUCHBASE_TLS_SKIP_VERIFY was ignored")
		}
	})

	t.Run("TLS can be turned off for a local cluster", func(t *testing.T) {
		t.Setenv("SCIM_COUCHBASE_TLS", "false")
		connection, security, err := couchbaseConnection("localhost")
		if err != nil {
			t.Fatal(err)
		}
		if connection != "couchbase://localhost" {
			t.Errorf("connection string = %s", connection)
		}
		if security.TLSSkipVerify {
			t.Error("skip-verify should be irrelevant without TLS")
		}
	})

	t.Run("a CA certificate is loaded", func(t *testing.T) {
		path := writeTestCA(t)
		t.Setenv("SCIM_COUCHBASE_CA_CERT", path)

		_, security, err := couchbaseConnection("db.example")
		if err != nil {
			t.Fatal(err)
		}
		if security.TLSRootCAs == nil {
			t.Fatal("the CA was not loaded")
		}
	})

	t.Run("a missing CA file is an error", func(t *testing.T) {
		t.Setenv("SCIM_COUCHBASE_CA_CERT", filepath.Join(t.TempDir(), "absent.pem"))
		if _, _, err := couchbaseConnection("db.example"); err == nil {
			t.Error("a missing CA file was accepted")
		}
	})

	t.Run("a CA file that is not a certificate is an error", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "junk.pem")
		if err := os.WriteFile(path, []byte("this is not a certificate"), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("SCIM_COUCHBASE_CA_CERT", path)
		if _, _, err := couchbaseConnection("db.example"); err == nil {
			t.Error("a file with no PEM certificate was accepted")
		}
	})
}

func TestEnvBool(t *testing.T) {
	cases := []struct {
		value    string
		fallback bool
		want     bool
	}{
		{"", true, true},
		{"", false, false},
		{"true", false, true},
		{"TRUE", false, true},
		{"1", false, true},
		{"yes", false, true},
		{"on", false, true},
		{"false", true, false},
		{"0", true, false},
		{"nonsense", true, false},
	}
	for _, tc := range cases {
		t.Setenv("SCIM_TEST_BOOL", tc.value)
		if got := envBool("SCIM_TEST_BOOL", tc.fallback); got != tc.want {
			t.Errorf("envBool(%q, %v) = %v, want %v", tc.value, tc.fallback, got, tc.want)
		}
	}
}

// writeTestCA writes a self-signed certificate and returns its path.
func writeTestCA(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "goscim test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "ca.pem")
	encoded := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

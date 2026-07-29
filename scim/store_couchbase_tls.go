package scim

import (
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/couchbase/gocb/v2"
)

// couchbaseConnection builds the connection string and TLS settings for the
// cluster from the environment.
//
// TLS is on by default and the certificate is verified. Skipping verification
// used to be hardcoded, which meant the connection paid the cost of TLS while
// giving none of its protection: anyone able to intercept the link to the
// database could present any certificate. Turning it off is still possible for
// a local cluster with a self-signed certificate, but it has to be asked for.
//
//	SCIM_COUCHBASE_TLS               "false" to connect over plain couchbase://
//	SCIM_COUCHBASE_CA_CERT           path to the PEM that signed the cluster's certificate
//	SCIM_COUCHBASE_TLS_SKIP_VERIFY   "true" to accept any certificate
func couchbaseConnection(endpoint string) (string, gocb.SecurityConfig, error) {
	security := gocb.SecurityConfig{}

	if !envBool("SCIM_COUCHBASE_TLS", true) {
		return "couchbase://" + endpoint, security, nil
	}

	if path := strings.TrimSpace(os.Getenv("SCIM_COUCHBASE_CA_CERT")); path != "" {
		pem, err := os.ReadFile(path)
		if err != nil {
			return "", security, fmt.Errorf("reading SCIM_COUCHBASE_CA_CERT: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return "", security, fmt.Errorf("SCIM_COUCHBASE_CA_CERT %q holds no PEM certificate", path)
		}
		security.TLSRootCAs = pool
	}

	security.TLSSkipVerify = envBool("SCIM_COUCHBASE_TLS_SKIP_VERIFY", false)
	return "couchbases://" + endpoint, security, nil
}

// envBool reads a boolean environment variable, falling back when it is unset.
func envBool(name string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "":
		return fallback
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

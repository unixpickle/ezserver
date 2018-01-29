package ezserver

import (
	"crypto/tls"
	"crypto/x509"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

// KeyCert represents a TLS key/centificate pair.
type KeyCert struct {
	Key         string `json:"key"`
	Certificate string `json:"certificate"`
}

// TLSConfig represents the TLS configuration for an HTTPS server.
type TLSConfig struct {
	Named   map[string]KeyCert `json:"named"`
	RootCAs []string           `json:"root_ca"`
	Default KeyCert            `json:"default"`

	ACMEDirectoryURL string   `json:"acme_dir_url"`
	ACMEHosts        []string `json:"acme_hosts"`
}

// Clone produces a deep copy of a TLSConfig object.
func (c *TLSConfig) Clone() *TLSConfig {
	named := map[string]KeyCert{}
	for key, val := range c.Named {
		named[key] = val
	}

	roots := make([]string, len(c.RootCAs))
	for i, x := range c.RootCAs {
		roots[i] = x
	}

	return &TLSConfig{named, roots, c.Default, c.ACMEDirectoryURL, c.ACMEHosts}
}

// ToConfig turns a TLSConfig into a tls.Config.
func (c *TLSConfig) ToConfig() (*tls.Config, *autocert.Manager, error) {
	var err error
	res := &tls.Config{}

	res.NextProtos = []string{"http/1.1", "h2"}

	res.Certificates = make([]tls.Certificate, 1)
	res.Certificates[0], err = tls.X509KeyPair([]byte(c.Default.Certificate),
		[]byte(c.Default.Key))
	if err != nil {
		return nil, nil, err
	}

	res.NameToCertificate = map[string]*tls.Certificate{}
	for name, pair := range c.Named {
		loaded, err := tls.X509KeyPair([]byte(pair.Certificate),
			[]byte(pair.Key))
		if err != nil {
			return nil, nil, err
		}
		idx := len(res.Certificates)
		res.Certificates = append(res.Certificates, loaded)
		res.NameToCertificate[name] = &res.Certificates[idx]
	}

	if len(c.RootCAs) > 0 {
		pool := x509.NewCertPool()
		for _, pemData := range c.RootCAs {
			if !pool.AppendCertsFromPEM([]byte(pemData)) {
				return nil, nil, ErrInvalidRootCA
			}
		}
		res.RootCAs = pool
	}

	var manager *autocert.Manager
	if len(c.ACMEHosts) > 0 {
		manager = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(c.ACMEHosts...),
		}
		if c.ACMEDirectoryURL != "" {
			manager.Client = &acme.Client{DirectoryURL: c.ACMEDirectoryURL}
		}
		res.GetCertificate = manager.GetCertificate
	}

	return res, manager, nil
}

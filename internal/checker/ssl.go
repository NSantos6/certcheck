package checker

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

type SSLResult struct {
	Domain    string
	ExpiresAt time.Time
	DaysLeft  int
	Issuer    string
	Error     error
}

func CheckSSL(domain string) SSLResult {
	result := SSLResult{Domain: domain}

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		net.JoinHostPort(domain, "443"),
		&tls.Config{ServerName: domain},
	)
	if err != nil {
		result.Error = fmt.Errorf("connection failed: %w", err)
		return result
	}
	defer conn.Close()

	cert := conn.ConnectionState().PeerCertificates[0]
	result.ExpiresAt = cert.NotAfter
	result.DaysLeft = int(time.Until(cert.NotAfter).Hours() / 24)
	result.Issuer = cert.Issuer.Organization[0]

	return result
}

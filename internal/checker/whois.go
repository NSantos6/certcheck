package checker

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

type DomainResult struct {
	Domain    string
	ExpiresAt time.Time
	DaysLeft  int
	Registrar string
	Error     error
}

type whoisServer struct {
	host       string
	expiryKey  string
	dateFormat string
}

var whoisServers = map[string]whoisServer{
	"br": {
		host:       "whois.registro.br:43",
		expiryKey:  "expires:",
		dateFormat: "20060102",
	},
	"default": {
		host:       "whois.iana.org:43",
		expiryKey:  "Expiry Date:",
		dateFormat: time.RFC3339,
	},
}

var dateFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05Z",
	"2006-01-02",
	"20060102",
	"02/01/2006",
	"2006-01-02 15:04:05",
}

func CheckDomain(domain string) DomainResult {
	result := DomainResult{Domain: domain}

	tld := extractTLD(domain)
	server, ok := whoisServers[tld]
	if !ok {
		server = whoisServers["default"]
		// For unknown TLDs, try to find the right WHOIS server
		server.host = fmt.Sprintf("whois.nic.%s:43", tld)
	}

	raw, err := queryWhois(server.host, domain)
	if err != nil {
		// Fallback to IANA for unknown TLDs
		raw, err = queryWhois("whois.iana.org:43", tld)
		if err != nil {
			result.Error = fmt.Errorf("whois query failed: %w", err)
			return result
		}
		// Extract referral WHOIS server from IANA response and retry
		referral := extractField(raw, "whois:")
		if referral != "" {
			raw, err = queryWhois(referral+":43", domain)
			if err != nil {
				result.Error = fmt.Errorf("whois query failed: %w", err)
				return result
			}
		}
	}

	expiry := extractExpiry(raw)
	if expiry.IsZero() {
		result.Error = fmt.Errorf("could not parse expiry date from WHOIS response")
		return result
	}

	result.ExpiresAt = expiry
	result.DaysLeft = int(time.Until(expiry).Hours() / 24)
	result.Registrar = extractField(raw, "owner:")
	if result.Registrar == "" {
		result.Registrar = extractField(raw, "Registrar:")
	}

	return result
}

func queryWhois(server, domain string) (string, error) {
	conn, err := net.DialTimeout("tcp", server, 10*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(15 * time.Second))
	fmt.Fprintf(conn, "%s\r\n", domain)

	var sb strings.Builder
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		sb.WriteString(scanner.Text() + "\n")
	}

	return sb.String(), nil
}

func extractExpiry(raw string) time.Time {
	expiryKeys := []string{
		"expires:", "Expiry Date:", "Registry Expiry Date:",
		"Expiration Date:", "paid-till:", "expiration:",
		"Registrar Registration Expiration Date:",
	}

	for _, line := range strings.Split(raw, "\n") {
		lower := strings.ToLower(line)
		for _, key := range expiryKeys {
			if strings.Contains(lower, strings.ToLower(key)) {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) < 2 {
					continue
				}
				val := strings.TrimSpace(parts[1])
				for _, format := range dateFormats {
					if t, err := time.Parse(format, val); err == nil {
						return t
					}
				}
			}
		}
	}

	return time.Time{}
}

func extractField(raw, key string) string {
	for _, line := range strings.Split(raw, "\n") {
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, strings.ToLower(key)) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

func extractTLD(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

func validateDomain(domain string) error {
	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("%q não é um domínio válido", domain)
	}
	return nil
}

func resolveDomains(args []string, file string) ([]string, error) {
	if file == "" {
		for _, d := range args {
			if err := validateDomain(d); err != nil {
				return nil, err
			}
		}
		return args, nil
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer f.Close()

	var domains []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") && strings.TrimSpace(strings.TrimPrefix(line, "-")) == "" {
			continue
		}
		// Handle YAML list items: "- example.com"
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimSpace(line)
		if line != "" {
			domains = append(domains, line)
		}
	}

	all := append(domains, args...)
	for _, d := range all {
		if err := validateDomain(d); err != nil {
			return nil, err
		}
	}
	return all, scanner.Err()
}

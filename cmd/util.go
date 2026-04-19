package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func resolveDomains(args []string, file string) ([]string, error) {
	if file == "" {
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

	return append(domains, args...), scanner.Err()
}

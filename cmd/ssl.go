package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/NSantos6/certcheck/internal/checker"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var sslWarnDays int
var sslOutputJSON bool
var sslFile string

var sslCmd = &cobra.Command{
	Use:   "ssl [domain...]",
	Short: "Check SSL certificate expiry",
	Example: `  certcheck ssl example.com
  certcheck ssl example.com google.com
  certcheck ssl --file domains.yaml --warn-days 14`,
	RunE: func(cmd *cobra.Command, args []string) error {
		domains, err := resolveDomains(args, sslFile)
		if err != nil {
			return err
		}
		if len(domains) == 0 {
			return fmt.Errorf("no domains provided")
		}

		results := runSSLChecks(domains)

		if sslOutputJSON {
			return printSSLJSON(results)
		}
		printSSLTable(results, sslWarnDays)
		return nil
	},
}

func init() {
	sslCmd.Flags().IntVar(&sslWarnDays, "warn-days", 30, "warn when cert expires within N days")
	sslCmd.Flags().BoolVar(&sslOutputJSON, "json", false, "output as JSON")
	sslCmd.Flags().StringVar(&sslFile, "file", "", "file with list of domains (one per line or YAML)")
	rootCmd.AddCommand(sslCmd)
}

func runSSLChecks(domains []string) []checker.SSLResult {
	results := make([]checker.SSLResult, len(domains))
	var wg sync.WaitGroup

	for i, d := range domains {
		wg.Add(1)
		go func(idx int, domain string) {
			defer wg.Done()
			results[idx] = checker.CheckSSL(domain)
		}(i, d)
	}

	wg.Wait()
	return results
}

func printSSLTable(results []checker.SSLResult, warnDays int) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Domain", "Expires", "Days Left", "Issuer", "Status"})
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
	})

	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	for _, r := range results {
		if r.Error != nil {
			table.Append([]string{r.Domain, "-", "-", "-", red("ERROR: " + r.Error.Error())})
			continue
		}

		status := green("OK")
		days := fmt.Sprintf("%d", r.DaysLeft)
		if r.DaysLeft < 0 {
			status = red("EXPIRED")
			days = red(days)
		} else if r.DaysLeft <= warnDays {
			status = yellow(fmt.Sprintf("EXPIRING SOON"))
			days = yellow(days)
		}

		table.Append([]string{
			r.Domain,
			r.ExpiresAt.Format("2006-01-02"),
			days,
			r.Issuer,
			status,
		})
	}

	table.Render()
}

func printSSLJSON(results []checker.SSLResult) error {
	type jsonResult struct {
		Domain    string `json:"domain"`
		ExpiresAt string `json:"expires_at,omitempty"`
		DaysLeft  int    `json:"days_left,omitempty"`
		Issuer    string `json:"issuer,omitempty"`
		Error     string `json:"error,omitempty"`
	}

	out := make([]jsonResult, len(results))
	for i, r := range results {
		out[i] = jsonResult{Domain: r.Domain, DaysLeft: r.DaysLeft, Issuer: r.Issuer}
		if r.Error != nil {
			out[i].Error = r.Error.Error()
		} else {
			out[i].ExpiresAt = r.ExpiresAt.Format("2006-01-02")
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
